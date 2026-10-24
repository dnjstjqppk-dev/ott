mod cache;

use aws_sdk_s3::Client as S3Client;
use axum::{
    body::Body,
    extract::{Path, Query, State},
    http::{header, HeaderMap, StatusCode},
    response::Response,
    routing::get,
    Router,
};
use bytes::Bytes;
use cache::TieredSegmentCache;
use hmac::{Hmac, Mac};
use serde::Deserialize;
use sha2::Sha256;
use std::net::SocketAddr;
use std::path::PathBuf;
use std::sync::Arc;
use std::time::{SystemTime, UNIX_EPOCH};

type HmacSha256 = Hmac<Sha256>;

#[derive(Clone)]
struct EdgeState {
    cache: Arc<TieredSegmentCache>,
    s3: S3Client,
    bucket: String,
    hmac_secret: String,
}

#[derive(Deserialize)]
struct SignedQuery {
    token: String,
    exp: i64,
    ep: String,
    ip: String,
}

#[tokio::main]
async fn main() {
    let nvme_path = PathBuf::from(std::env::var("EDGE_NVME_CACHE_DIR").unwrap_or_else(|_| "/tmp/open_connect_cache".into()));
    let _ = tokio::fs::create_dir_all(&nvme_path).await;
    let aws_cfg = aws_config::load_from_env().await;
    let state = EdgeState {
        cache: TieredSegmentCache::new(4096, nvme_path),
        s3: S3Client::new(&aws_cfg),
        bucket: std::env::var("PACKAGED_BUCKET").unwrap_or_else(|_| "nf-media-packaged".into()),
        hmac_secret: std::env::var("EDGE_HMAC_SECRET").expect("EDGE_HMAC_SECRET required"),
    };

    let app = Router::new()
        .route("/stream/*object_key", get(handle_byte_range_stream))
        .with_state(state);

    let addr: SocketAddr = std::env::var("EDGE_HTTP_ADDR").unwrap_or_else(|_| "0.0.0.0:8090".into()).parse().unwrap();
    let listener = tokio::net::TcpListener::bind(addr).await.unwrap();
    axum::serve(listener, app).await.unwrap();
}

async fn handle_byte_range_stream(
    State(st): State<EdgeState>,
    Path(object_key): Path<String>,
    Query(q): Query<SignedQuery>,
    headers: HeaderMap,
) -> Result<Response, StatusCode> {
    verify_hmac_signature(&st.hmac_secret, &q)?;
    let full_bytes = fetch_or_populate_cache(&st, &object_key).await?;
    let total_len = full_bytes.len();
    if total_len == 0 {
        return Err(StatusCode::NOT_FOUND);
    }

    if let Some(range_val) = headers.get(header::RANGE) {
        let range_str = range_val.to_str().map_err(|_| StatusCode::BAD_REQUEST)?;
        let (start, end) = parse_byte_range(range_str, total_len).ok_or(StatusCode::RANGE_NOT_SATISFIABLE)?;
        let slice = full_bytes.slice(start..=end);
        let content_range = format!("bytes {}-{}/{}", start, end, total_len);
        return Ok(Response::builder()
            .status(StatusCode::PARTIAL_CONTENT)
            .header(header::CONTENT_TYPE, detect_mime(&object_key))
            .header(header::ACCEPT_RANGES, "bytes")
            .header(header::CONTENT_RANGE, content_range)
            .header(header::CONTENT_LENGTH, slice.len().to_string())
            .header(header::ACCESS_CONTROL_ALLOW_ORIGIN, "*")
            .body(Body::from(slice))
            .unwrap());
    }

    Ok(Response::builder()
        .status(StatusCode::OK)
        .header(header::CONTENT_TYPE, detect_mime(&object_key))
        .header(header::ACCEPT_RANGES, "bytes")
        .header(header::CONTENT_LENGTH, total_len.to_string())
        .header(header::ACCESS_CONTROL_ALLOW_ORIGIN, "*")
        .body(Body::from(full_bytes))
        .unwrap())
}

fn verify_hmac_signature(secret: &str, q: &SignedQuery) -> Result<(), StatusCode> {
    let now = SystemTime::now().duration_since(UNIX_EPOCH).unwrap_or_default().as_secs() as i64;
    if q.exp < now {
        return Err(StatusCode::FORBIDDEN);
    }
    let mut mac = HmacSha256::new_from_slice(secret.as_bytes()).map_err(|_| StatusCode::INTERNAL_SERVER_ERROR)?;
    mac.update(format!("{}:{}:{}", q.ep, q.ip, q.exp).as_bytes());
    let expected = hex::encode(mac.finalize().into_bytes());
    if expected != q.token {
        return Err(StatusCode::FORBIDDEN);
    }
    Ok(())
}

async fn fetch_or_populate_cache(st: &EdgeState, key: &str) -> Result<Bytes, StatusCode> {
    if let Some(cached) = st.cache.get_segment(key).await {
        return Ok(cached);
    }
    let out = st.s3.get_object().bucket(&st.bucket).key(key).send().await.map_err(|_| StatusCode::NOT_FOUND)?;
    let collected = out.body.collect().await.map_err(|_| StatusCode::BAD_GATEWAY)?;
    let data = collected.into_bytes();
    st.cache.put_segment(key, data.clone()).await;
    Ok(data)
}

fn parse_byte_range(header_val: &str, total_len: usize) -> Option<(usize, usize)> {
    let stripped = header_val.strip_prefix("bytes=")?;
    let mut parts = stripped.splitn(2, '-');
    let start: usize = parts.next()?.parse().ok()?;
    if start >= total_len {
        return None;
    }
    let end_str = parts.next().unwrap_or("");
    let end = if end_str.is_empty() {
        total_len - 1
    } else {
        end_str.parse::<usize>().ok()?.min(total_len - 1)
    };
    if start > end {
        return None;
    }
    Some((start, end))
}

fn detect_mime(key: &str) -> &'static str {
    if key.ends_with(".m3u8") {
        "application/vnd.apple.mpegurl"
    } else if key.ends_with(".mpd") {
        "application/dash+xml"
    } else if key.ends_with(".vtt") {
        "text/vtt"
    } else {
        "video/iso.segment"
    }
}

