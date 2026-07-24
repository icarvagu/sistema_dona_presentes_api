use std::collections::HashMap;
use std::sync::Arc;
use std::time::{Duration, Instant};
use tokio::sync::Mutex;

#[derive(Clone)]
pub struct RateLimiter {
    store: Arc<Mutex<HashMap<String, RateLimitEntry>>>,
}

struct RateLimitEntry {
    count: u32,
    reset_at: Instant,
}

impl RateLimiter {
    pub fn new() -> Self {
        Self {
            store: Arc::new(Mutex::new(HashMap::new())),
        }
    }

    pub async fn check(&self, key: &str, max_requests: u32, window: Duration) -> bool {
        let mut store = self.store.lock().await;
        let now = Instant::now();

        let entry = store
            .entry(key.to_string())
            .or_insert_with(|| RateLimitEntry {
                count: 0,
                reset_at: now + window,
            });

        if now > entry.reset_at {
            entry.count = 0;
            entry.reset_at = now + window;
        }

        if entry.count >= max_requests {
            return false;
        }

        entry.count += 1;
        true
    }

    pub async fn cleanup(&self) {
        let mut store = self.store.lock().await;
        let now = Instant::now();
        store.retain(|_, entry| now <= entry.reset_at);
    }
}
