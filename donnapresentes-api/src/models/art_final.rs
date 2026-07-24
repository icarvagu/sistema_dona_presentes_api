use chrono::NaiveDateTime;
use serde::Deserialize;

#[derive(Debug, Deserialize)]
pub struct LayoutRequestInput {
    pub source_type: String,
    pub source_id: i32,
    pub title: String,
    pub instructions: Option<String>,
    pub file_mode: Option<String>,
    pub common_file_url: Option<String>,
    pub priority: Option<i32>,
    pub due_at: Option<NaiveDateTime>,
    pub assigned_to: Option<i32>,
    pub items: Vec<LayoutRequestItemInput>,
}

#[derive(Debug, Deserialize)]
pub struct LayoutRequestItemInput {
    pub entity_type: String,
    pub item_id: i32,
    pub group_key: Option<String>,
    pub source_file_url: Option<String>,
}

#[derive(Debug, Deserialize)]
pub struct LayoutTransitionInput {
    pub status: String,
    pub note: Option<String>,
}

#[derive(Debug, Deserialize)]
pub struct LayoutMessageInput {
    pub message: String,
    pub file_url: Option<String>,
}

#[derive(Debug, Deserialize)]
pub struct LayoutVersionInput {
    pub file_url: String,
    pub label: Option<String>,
}

#[derive(Debug, Deserialize)]
pub struct LayoutDecisionInput {
    pub status: String,
    pub note: Option<String>,
}

#[derive(Debug, Deserialize)]
pub struct LayoutJobInput {
    pub status: String,
    pub responsible_id: Option<i32>,
    pub external_contact: Option<String>,
    pub due_at: Option<NaiveDateTime>,
    pub file_url: Option<String>,
    pub tags: Option<Vec<String>>,
    pub reason: Option<String>,
}

#[derive(Debug, Deserialize)]
pub struct StoryLifecycleInput {
    pub status: String,
    pub observation: Option<String>,
    pub expires_at: Option<NaiveDateTime>,
}

#[derive(Debug, Deserialize)]
pub struct ArtFinalTaskInput {
    pub panel: Option<String>,
    pub category: String,
    pub title: String,
    pub description: Option<String>,
    pub sale_id: Option<i32>,
    pub purchase_id: Option<i32>,
    pub due_date: Option<NaiveDateTime>,
    pub due_at: Option<NaiveDateTime>,
    pub assigned_to: Option<i32>,
    pub channel: Option<String>,
    pub format: Option<String>,
    pub priority: Option<i32>,
    pub tags: Option<Vec<String>>,
    pub attachment_url: Option<String>,
    pub status: Option<String>,
}

#[derive(Debug, Deserialize)]
pub struct ArtFinalStoryInput {
    pub sale_id: i32,
    pub sale_item_id: Option<i32>,
    pub title: String,
    pub tags: Option<Vec<String>>,
}
