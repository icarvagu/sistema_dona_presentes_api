use chrono::NaiveDateTime;
use serde::{Deserialize, Serialize};
use serde_json::Value;

#[derive(Debug, Clone, Serialize, Deserialize, sqlx::FromRow)]
pub struct QuoteFeedbackEvent {
    pub id: i64,
    pub quote_id: i32,
    pub scheduled_at: Option<NaiveDateTime>,
    pub observation: Option<String>,
    pub created_by: i32,
    pub created_at: chrono::DateTime<chrono::Utc>,
}

#[derive(Debug, Clone, Serialize, Deserialize, sqlx::FromRow)]
pub struct ItemLayoutVersion {
    pub id: i64,
    pub entity_type: String,
    pub item_id: i32,
    pub version: i32,
    pub label: Option<String>,
    pub file_url: String,
    pub item_snapshot: Option<Value>,
    pub approval_status: String,
    pub approval_note: Option<String>,
    pub approved_by: Option<i32>,
    pub approved_at: Option<NaiveDateTime>,
    pub created_by: i32,
    pub created_at: chrono::DateTime<chrono::Utc>,
}

#[derive(Debug, Deserialize)]
pub struct FinancialAnalysisInput {
    pub status: String,
    pub tags: Option<Vec<String>>,
    pub observation: Option<String>,
    pub attachment_url: Option<String>,
}

#[derive(Debug, Deserialize)]
pub struct SalePendingInput {
    pub sector: String,
    pub description: String,
    pub blocking: bool,
}

#[derive(Debug, Deserialize)]
pub struct EngravingApprovalInput {
    pub response: String,
    pub observation: Option<String>,
}

#[derive(Debug, Deserialize)]
pub struct QuoteConversionInput {
    pub withdrawal_dates: std::collections::HashMap<String, Option<NaiveDateTime>>,
}

#[derive(Debug, Clone, Serialize, Deserialize, sqlx::FromRow)]
pub struct WorkflowAlert {
    pub id: i64,
    pub user_id: i32,
    pub sale_id: Option<i32>,
    pub quote_id: Option<i32>,
    pub title: String,
    pub description: String,
    pub scheduled_at: chrono::DateTime<chrono::Utc>,
    pub status: String,
    pub resolved_at: Option<chrono::DateTime<chrono::Utc>>,
    pub created_at: chrono::DateTime<chrono::Utc>,
    pub updated_at: chrono::DateTime<chrono::Utc>,
}

#[derive(Debug, Clone, Deserialize)]
pub struct WorkflowAlertInput {
    pub sale_id: Option<i32>,
    pub quote_id: Option<i32>,
    pub title: String,
    pub description: String,
    pub scheduled_at: chrono::DateTime<chrono::Utc>,
}
