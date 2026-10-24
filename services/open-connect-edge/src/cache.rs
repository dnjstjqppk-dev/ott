use bytes::Bytes;
use lru::LruCache;
use parking_lot::Mutex;
use sha2::{Digest, Sha256};
use std::num::NonZeroUsize;
use std::path::PathBuf;
use std::sync::Arc;

pub struct TieredSegmentCache {
    mem_lru: Mutex<LruCache<String, Bytes>>,
    nvme_dir: PathBuf,
}

impl TieredSegmentCache {
    pub fn new(capacity_items: usize, nvme_dir: PathBuf) -> Arc<Self> {
        let cap = NonZeroUsize::new(capacity_items).unwrap_or(NonZeroUsize::new(1024).unwrap());
        Arc::new(Self {
            mem_lru: Mutex::new(LruCache::new(cap)),
            nvme_dir,
        })
    }

    pub async fn get_segment(&self, key: &str) -> Option<Bytes> {
        if let Some(hit) = self.mem_lru.lock().get(key).cloned() {
            return Some(hit);
        }
        let disk_path = self.nvme_dir.join(Self::safe_filename(key));
        if let Ok(raw) = tokio::fs::read(&disk_path).await {
            let b = Bytes::from(raw);
            self.mem_lru.lock().put(key.to_string(), b.clone());
            return Some(b);
        }
        None
    }

    pub async fn put_segment(&self, key: &str, data: Bytes) {
        self.mem_lru.lock().put(key.to_string(), data.clone());
        let disk_path = self.nvme_dir.join(Self::safe_filename(key));
        let _ = tokio::fs::write(disk_path, &data).await;
    }

    fn safe_filename(key: &str) -> String {
        let mut hasher = Sha256::new();
        hasher.update(key.as_bytes());
        hex::encode(hasher.finalize())
    }
}

