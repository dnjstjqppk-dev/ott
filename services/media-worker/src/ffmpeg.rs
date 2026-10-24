use anyhow::{Context, Result};
use serde::Deserialize;
use std::path::Path;
use tokio::process::Command;

#[derive(Debug, Clone)]
pub struct SceneChunkSpec {
    pub chunk_index: i32,
    pub start_ms: i64,
    pub end_ms: i64,
    pub start_frame: i64,
    pub end_frame: i64,
    pub spatial_complexity: f64,
    pub temporal_complexity: f64,
}

#[derive(Debug, Deserialize)]
struct VmafRoot {
    pooled_metrics: VmafPooled,
}

#[derive(Debug, Deserialize)]
struct VmafPooled {
    vmaf: VmafMetric,
}

#[derive(Debug, Deserialize)]
struct VmafMetric {
    harmonic_mean: f64,
    min: f64,
}

pub async fn detect_scenes(input_path: &Path) -> Result<Vec<SceneChunkSpec>> {
    let output = Command::new("ffprobe")
        .args([
            "-v", "error",
            "-show_entries", "Packet=pts_time,flags",
            "-select_streams", "v:0",
            "-of", "csv=p=0",
            input_path.to_str().unwrap_or_default(),
        ])
        .output()
        .await
        .context("failed to execute ffprobe for scene detection")?;

    let stdout = String::from_utf8_lossy(&output.stdout);
    Ok(parse_keyframe_intervals(&stdout))
}

fn parse_keyframe_intervals(csv_out: &str) -> Vec<SceneChunkSpec> {
    let mut specs = Vec::new();
    let mut last_ms: i64 = 0;
    let mut idx: i32 = 0;
    for line in csv_out.lines() {
        if line.contains('K') {
            if let Some(ts_str) = line.split(',').next() {
                if let Ok(sec) = ts_str.parse::<f64>() {
                    let cur_ms = (sec * 1000.0) as i64;
                    if cur_ms - last_ms >= 4000 {
                        specs.push(SceneChunkSpec {
                            chunk_index: idx,
                            start_ms: last_ms,
                            end_ms: cur_ms,
                            start_frame: (last_ms * 24) / 1000,
                            end_frame: (cur_ms * 24) / 1000,
                            spatial_complexity: 68.5,
                            temporal_complexity: 42.1,
                        });
                        last_ms = cur_ms;
                        idx += 1;
                    }
                }
            }
        }
    }
    if specs.is_empty() {
        specs.push(SceneChunkSpec {
            chunk_index: 0, start_ms: 0, end_ms: 4000,
            start_frame: 0, end_frame: 96,
            spatial_complexity: 65.0, temporal_complexity: 40.0,
        });
    }
    specs
}

pub async fn transcode_cmaf_segment(
    src: &Path,
    out_dir: &Path,
    codec: &str,
    width: i32,
    height: i32,
    bitrate_kbps: i32,
    drm_key_hex: &str,
    drm_kid_hex: &str,
) -> Result<()> {
    let encoder = match codec {
        "H265_HEVC" => "libx265",
        "AV1" => "libsvtav1",
        _ => "libx264",
    };
    let scale_filter = format!("scale={}:{}", width, height);
    let br = format!("{}k", bitrate_kbps);
    let m3u8_out = out_dir.join("stream.m3u8");
    let init_name = "init.mp4";
    let seg_pattern = out_dir.join("segment_%d.m4s");

    let status = Command::new("ffmpeg")
        .args([
            "-y", "-i", src.to_str().unwrap_or_default(),
            "-vf", &scale_filter,
            "-c:v", encoder, "-b:v", &br, "-maxrate", &br, "-bufsize", &br,
            "-f", "hls", "-hls_segment_type", "fmp4",
            "-hls_time", "4",
            "-hls_fmp4_init_filename", init_name,
            "-encryption_scheme", "cenc-aes-ctr",
            "-encryption_key", drm_key_hex,
            "-encryption_kid", drm_kid_hex,
            "-hls_segment_filename", seg_pattern.to_str().unwrap_or_default(),
            m3u8_out.to_str().unwrap_or_default(),
        ])
        .status()
        .await
        .context("ffmpeg cmaf transcode failed")?;

    if !status.success() {
        anyhow::bail!("ffmpeg exited with non-zero status");
    }
    Ok(())
}

pub async fn evaluate_vmaf(reference: &Path, distorted: &Path, vmaf_json: &Path) -> Result<(f64, f64)> {
    let filter = format!("libvmaf=log_fmt=json:log_path={}", vmaf_json.to_str().unwrap_or_default());
    let status = Command::new("ffmpeg")
        .args([
            "-i", distorted.to_str().unwrap_or_default(),
            "-i", reference.to_str().unwrap_or_default(),
            "-lavfi", &filter, "-f", "null", "-",
        ])
        .status()
        .await?;
    if !status.success() {
        anyhow::bail!("libvmaf evaluation failed");
    }
    let raw = tokio::fs::read_to_string(vmaf_json).await?;
    let parsed: VmafRoot = serde_json::from_str(&raw)?;
    Ok((parsed.pooled_metrics.vmaf.harmonic_mean, parsed.pooled_metrics.vmaf.min))
}

