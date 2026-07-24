use chrono::NaiveDateTime;
use serde::{Deserialize, Serialize};

#[derive(Debug, Clone, Serialize, Deserialize, sqlx::FromRow)]
pub struct Notification {
    pub id: i64,
    pub purchase_id: Option<i32>,
    pub recipient_user_id: Option<i32>,
    pub recipient_permission: Option<String>,
    pub notification_type: String,
    pub message: String,
    pub read_at: Option<NaiveDateTime>,
    pub created_at: NaiveDateTime,
}
