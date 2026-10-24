mod ffmpeg;
mod pipeline;

use aws_sdk_s3::Client as S3Client;
use hmac::{Hmac, Mac};
use sha2::Sha256;
use sqlx::{PgPool, Row};
use std::net::SocketAddr;
use std::path::PathBuf;
use std::time::{SystemTime, UNIX_EPOCH};
use tonic::{transport::Server, Request, Response, Status};
use uuid::Uuid;

pub mod mediav1 {
    tonic::include_proto!("netflix.media.v1");
}
use mediav1::media_pipeline_service_server::{MediaPipelineService, MediaPipelineServiceServer};
use mediav1::*;

type HmacSha256 = Hmac<Sha256>;

pub struct MediaWorkerService {
    pool: PgPool,
    s3: S3Client,
    edge_hmac_secret: String,
}

#[tonic::async_trait]
impl MediaPipelineService for MediaWorkerService {
    async fn dispatch_encoding_job(&self, req: Request<DispatchEncodingJobRequest>) -> Result<Response<EncodingJobStatusDto>, Status> {
        let inner = req.into_inner();
        let ep_id = Uuid::parse_str(&inner.episode_id).map_err(|_| Status::invalid_argument("bad episode_id"))?;
        let ing_id = Uuid::parse_str(&inner.ingest_session_id).map_err(|_| Status::invalid_argument("bad ingest_id"))?;
        let job_id = Uuid::new_v4();

        sqlx::query("INSERT INTO encoding_jobs (id, episode_id, ingest_session_id, state, min_target_vmaf) VALUES ($1, $2, $3, 'QUEUED', $4)")
            .bind(job_id).bind(ep_id).bind(ing_id).bind(inner.min_target_vmaf)
            .execute(&self.pool).await.map_err(|e| Status::internal(e.to_string()))?;

        let pool_cl = self.pool.clone();
        let s3_cl = self.s3.clone();
        tokio::spawn(async move {
            let src = PathBuf::from(format!("/tmp/ingest_{}.mp4", ing_id));
            let work = PathBuf::from(format!("/tmp/work_{}", job_id));
            let _ = pipeline::execute_encoding_pipeline(&pool_cl, &s3_cl, job_id, ep_id, &src, &work, inner.min_target_vmaf).await;
        });

        Ok(Response::new(EncodingJobStatusDto {
            job_id: job_id.to_string(),
            episode_id: ep_id.to_string(),
            state: "QUEUED".into(),
            total_chunks: 0,
            completed_chunks: 0,
            master_hls_key: String::new(),
            master_dash_key: String::new(),
            error_message: String::new(),
            representations: vec![],
        }))
    }

    async fn get_encoding_job_status(&self, req: Request<GetEncodingJobStatusRequest>) -> Result<Response<EncodingJobStatusDto>, Status> {
        let jid = Uuid::parse_str(&req.into_inner().job_id).map_err(|_| Status::invalid_argument("invalid job_id"))?;
        let row = sqlx::query("SELECT episode_id, state::text, total_chunks, completed_chunks, COALESCE(master_hls_key,'') AS hls, COALESCE(master_dash_key,'') AS dash, COALESCE(error_message,'') AS err FROM encoding_jobs WHERE id = $1")
            .bind(jid).fetch_one(&self.pool).await.map_err(|_| Status::not_found("job not found"))?;

        Ok(Response::new(EncodingJobStatusDto {
            job_id: jid.to_string(),
            episode_id: row.get::<Uuid, _>("episode_id").to_string(),
            state: row.get("state"),
            total_chunks: row.get("total_chunks"),
            completed_chunks: row.get("completed_chunks"),
            master_hls_key: row.get("hls"),
            master_dash_key: row.get("dash"),
            error_message: row.get("err"),
            representations: vec![],
        }))
    }

    async fn retry_failed_chunk(&self, req: Request<RetryFailedChunkRequest>) -> Result<Response<EncodingJobStatusDto>, Status> {
        let jid = req.into_inner().job_id;
        self.get_encoding_job_status(Request::new(GetEncodingJobStatusRequest { job_id: jid })).await
    }

    async fn select_optimal_edge_node(&self, req: Request<SelectOptimalEdgeNodeRequest>) -> Result<Response<SelectOptimalEdgeNodeResponse>, Status> {
        let inner = req.into_inner();
        let row = sqlx::query("SELECT id, base_url FROM open_connect_edge_nodes WHERE country_code = $1 AND status = 'HEALTHY' ORDER BY current_bandwidth_gbps ASC LIMIT 1")
            .bind(&inner.country_code).fetch_one(&self.pool).await.map_err(|_| Status::unavailable("no healthy edge node"))?;

        let node_id: Uuid = row.get("id");
        let base_url: String = row.get("base_url");
        let exp = SystemTime::now().duration_since(UNIX_EPOCH).unwrap_or_default().as_secs() as i64 + 14400;
        let mut mac = HmacSha256::new_from_slice(self.edge_hmac_secret.as_bytes()).map_err(|_| Status::internal("hmac key err"))?;
        mac.update(format!("{}:{}:{}", inner.episode_id, inner.client_ip, exp).as_bytes());
        let token = hex::encode(mac.finalize().into_bytes());

        Ok(Response::new(SelectOptimalEdgeNodeResponse {
            edge_node_id: node_id.to_string(),
            edge_base_url: base_url,
            signed_hmac_token: token,
            token_expires_at_unix: exp,
        }))
    }

    async fn set_edge_node_drain_mode(&self, req: Request<SetEdgeNodeDrainModeRequest>) -> Result<Response<EdgeNodeDto>, Status> {
        let inner = req.into_inner();
        let nid = Uuid::parse_str(&inner.edge_node_id).map_err(|_| Status::invalid_argument("invalid edge_node_id"))?;
        let r = sqlx::query("UPDATE open_connect_edge_nodes SET status = $2 WHERE id = $1 RETURNING node_code, region_code, status, active_streams")
            .bind(nid).bind(&inner.target_status).fetch_one(&self.pool).await.map_err(|_| Status::not_found("node not found"))?;

        Ok(Response::new(EdgeNodeDto {
            id: nid.to_string(),
            node_code: r.get("node_code"),
            region_code: r.get("region_code"),
            status: r.get("status"),
            current_bandwidth_gbps: 0.0,
            cache_hit_ratio: 0.98,
            active_streams: r.get("active_streams"),
        }))
    }
}

#[tokio::main]
async fn main() -> Result<(), Box<dyn std::error::Error>> {
    let db_url = std::env::var("DATABASE_URL")?;
    let addr: SocketAddr = std::env::var("MEDIA_WORKER_GRPC_ADDR")?.parse()?;
    let secret = std::env::var("EDGE_HMAC_SECRET")?;
    let pool = PgPool::connect(&db_url).await?;
    let aws_cfg = aws_config::load_from_env().await;
    let s3 = S3Client::new(&aws_cfg);

    Server::builder()
        .add_service(MediaPipelineServiceServer::new(MediaWorkerService { pool, s3, edge_hmac_secret: secret }))
        .serve(addr)
        .await?;
    Ok(())
}

