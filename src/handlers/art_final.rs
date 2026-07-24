use axum::{
    extract::{Path, State},
    Extension, Json,
};

use crate::error::AppError;
use crate::middleware::auth::UserContext;
use crate::models::{
    ArtFinalStoryInput, ArtFinalTaskInput, LayoutDecisionInput, LayoutJobInput,
    LayoutMessageInput, LayoutRequestInput, LayoutTransitionInput, LayoutVersionInput,
    StoryLifecycleInput,
};
use crate::response::{created_response, no_content, ok_response};
use crate::AppState;
use sqlx::Column;
use sqlx::Row;

pub async fn dashboard(
    State(state): State<AppState>,
    _user: Extension<UserContext>,
) -> Result<impl axum::response::IntoResponse, AppError> {
    let task_count: i64 = sqlx::query_scalar("SELECT COUNT(*) FROM art_final_tasks")
        .fetch_one(&state.db).await
        .map_err(|e| AppError::internal(e.to_string()))?;
    Ok(ok_response(serde_json::json!({"total_tasks": task_count})))
}

pub async fn list_tasks(
    State(state): State<AppState>,
    _user: Extension<UserContext>,
) -> Result<impl axum::response::IntoResponse, AppError> {
    let rows = sqlx::query("SELECT * FROM art_final_tasks ORDER BY id DESC")
        .fetch_all(&state.db).await
        .map_err(|e| AppError::internal(e.to_string()))?;

    let tasks: Vec<serde_json::Value> = rows.iter().map(|r| {
        let mut map = serde_json::Map::new();
        for col in r.columns() {
            let name = col.name().to_string();
            let val: Option<String> = r.try_get(name.as_str()).unwrap_or(None);
            map.insert(name, serde_json::Value::String(val.unwrap_or_default()));
        }
        serde_json::Value::Object(map)
    }).collect();

    Ok(ok_response(tasks))
}

pub async fn create_task(
    State(state): State<AppState>,
    _user: Extension<UserContext>,
    Json(input): Json<ArtFinalTaskInput>,
) -> Result<impl axum::response::IntoResponse, AppError> {
    let id = sqlx::query_scalar::<_, i64>(
        "INSERT INTO art_final_tasks (panel, category, title, description, sale_id, purchase_id,
                due_date, due_at, assigned_to, channel, format, priority, tags, attachment_url, status)
         VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15) RETURNING id"
    )
    .bind(&input.panel).bind(&input.category).bind(&input.title)
    .bind(&input.description).bind(input.sale_id).bind(input.purchase_id)
    .bind(input.due_date).bind(input.due_at).bind(input.assigned_to)
    .bind(&input.channel).bind(&input.format).bind(input.priority.unwrap_or(0))
    .bind(&input.tags).bind(&input.attachment_url).bind(input.status.as_deref().unwrap_or("backlog"))
    .fetch_one(&state.db).await
    .map_err(|e| AppError::internal(e.to_string()))?;
    Ok(created_response(serde_json::json!({"id": id})))
}

pub async fn update_task(
    State(state): State<AppState>,
    _user: Extension<UserContext>,
    Path(id): Path<i64>,
    Json(input): Json<ArtFinalTaskInput>,
) -> Result<impl axum::response::IntoResponse, AppError> {
    sqlx::query(
        "UPDATE art_final_tasks SET category=$1, title=$2, description=$3, assigned_to=$4,
                priority=$5, status=$6, updated_at=NOW() WHERE id=$7"
    )
    .bind(&input.category).bind(&input.title).bind(&input.description)
    .bind(input.assigned_to).bind(input.priority.unwrap_or(0))
    .bind(input.status.as_deref().unwrap_or("backlog")).bind(id)
    .execute(&state.db).await
    .map_err(|e| AppError::internal(e.to_string()))?;
    Ok(ok_response(serde_json::json!({"message": "Tarefa atualizada"})))
}

pub async fn create_story(
    State(state): State<AppState>,
    _user: Extension<UserContext>,
    Json(input): Json<ArtFinalStoryInput>,
) -> Result<impl axum::response::IntoResponse, AppError> {
    let id: i64 = sqlx::query_scalar(
        "INSERT INTO art_final_stories (sale_id, sale_item_id, title, tags)
         VALUES ($1,$2,$3,$4) RETURNING id"
    )
    .bind(input.sale_id).bind(input.sale_item_id).bind(&input.title).bind(&input.tags)
    .fetch_one(&state.db).await
    .map_err(|e| AppError::internal(e.to_string()))?;
    Ok(created_response(serde_json::json!({"id": id})))
}

pub async fn check_story(
    State(state): State<AppState>,
    _user: Extension<UserContext>,
    Path(id): Path<i64>,
) -> Result<impl axum::response::IntoResponse, AppError> {
    sqlx::query("UPDATE art_final_stories SET checked = true, checked_at = NOW() WHERE id = $1")
        .bind(id).execute(&state.db).await
        .map_err(|e| AppError::internal(e.to_string()))?;
    Ok(ok_response(serde_json::json!({"message": "Story verificada"})))
}

pub async fn delete_story(
    State(state): State<AppState>,
    _user: Extension<UserContext>,
    Path(id): Path<i64>,
) -> Result<impl axum::response::IntoResponse, AppError> {
    sqlx::query("DELETE FROM art_final_stories WHERE id = $1").bind(id)
        .execute(&state.db).await
        .map_err(|e| AppError::internal(e.to_string()))?;
    Ok(no_content())
}

pub async fn update_story_lifecycle(
    State(state): State<AppState>,
    _user: Extension<UserContext>,
    Path(id): Path<i64>,
    Json(input): Json<StoryLifecycleInput>,
) -> Result<impl axum::response::IntoResponse, AppError> {
    sqlx::query(
        "UPDATE art_final_stories SET status = $1, observation = $2, expires_at = $3, updated_at = NOW() WHERE id = $4"
    )
    .bind(&input.status).bind(&input.observation).bind(input.expires_at).bind(id)
    .execute(&state.db).await
    .map_err(|e| AppError::internal(e.to_string()))?;
    Ok(ok_response(serde_json::json!({"message": "Status atualizado"})))
}

pub async fn list_layout_requests(
    State(state): State<AppState>,
    _user: Extension<UserContext>,
) -> Result<impl axum::response::IntoResponse, AppError> {
    let rows = sqlx::query("SELECT * FROM layout_requests ORDER BY id DESC")
        .fetch_all(&state.db).await
        .map_err(|e| AppError::internal(e.to_string()))?;

    let requests: Vec<serde_json::Value> = rows.iter().map(|r| {
        let mut map = serde_json::Map::new();
        for col in r.columns() {
            let name = col.name().to_string();
            let val: Option<String> = r.try_get(name.as_str()).unwrap_or(None);
            map.insert(name, serde_json::Value::String(val.unwrap_or_default()));
        }
        serde_json::Value::Object(map)
    }).collect();

    Ok(ok_response(requests))
}

pub async fn create_layout_request(
    State(state): State<AppState>,
    _user: Extension<UserContext>,
    Json(input): Json<LayoutRequestInput>,
) -> Result<impl axum::response::IntoResponse, AppError> {
    let id: i64 = sqlx::query_scalar(
        "INSERT INTO layout_requests (source_type, source_id, title, instructions, file_mode,
                common_file_url, priority, due_at, assigned_to)
         VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9) RETURNING id"
    )
    .bind(&input.source_type).bind(input.source_id).bind(&input.title)
    .bind(&input.instructions).bind(&input.file_mode).bind(&input.common_file_url)
    .bind(input.priority.unwrap_or(0)).bind(input.due_at).bind(input.assigned_to)
    .fetch_one(&state.db).await
    .map_err(|e| AppError::internal(e.to_string()))?;

    for item in &input.items {
        sqlx::query(
            "INSERT INTO layout_request_items (request_id, entity_type, item_id, group_key, source_file_url)
             VALUES ($1,$2,$3,$4,$5)"
        )
        .bind(id).bind(&item.entity_type).bind(item.item_id)
        .bind(&item.group_key).bind(&item.source_file_url)
        .execute(&state.db).await.ok();
    }

    Ok(created_response(serde_json::json!({"id": id})))
}

pub async fn get_layout_request(
    State(state): State<AppState>,
    _user: Extension<UserContext>,
    Path(id): Path<i64>,
) -> Result<impl axum::response::IntoResponse, AppError> {
    let rows = sqlx::query("SELECT * FROM layout_requests WHERE id = $1")
        .bind(id)
        .fetch_all(&state.db).await
        .map_err(|e| AppError::internal(e.to_string()))?;

    if rows.is_empty() {
        return Err(AppError::not_found("Layout request"));
    }

    let r = &rows[0];
    let mut map = serde_json::Map::new();
    for col in r.columns() {
        let name = col.name().to_string();
        let val: Option<String> = r.try_get(name.as_str()).unwrap_or(None);
        map.insert(name, serde_json::Value::String(val.unwrap_or_default()));
    }

    Ok(ok_response(serde_json::Value::Object(map)))
}

pub async fn transition_layout(
    State(state): State<AppState>,
    _user: Extension<UserContext>,
    Path(id): Path<i64>,
    Json(input): Json<LayoutTransitionInput>,
) -> Result<impl axum::response::IntoResponse, AppError> {
    sqlx::query("UPDATE layout_requests SET status = $1, updated_at = NOW() WHERE id = $2")
        .bind(&input.status).bind(id)
        .execute(&state.db).await
        .map_err(|e| AppError::internal(e.to_string()))?;
    Ok(ok_response(serde_json::json!({"message": "Status atualizado"})))
}

pub async fn add_layout_message(
    State(state): State<AppState>,
    _user: Extension<UserContext>,
    Path(id): Path<i64>,
    Json(input): Json<LayoutMessageInput>,
) -> Result<impl axum::response::IntoResponse, AppError> {
    sqlx::query(
        "INSERT INTO layout_messages (request_id, message, file_url) VALUES ($1,$2,$3)"
    )
    .bind(id).bind(&input.message).bind(&input.file_url)
    .execute(&state.db).await
    .map_err(|e| AppError::internal(e.to_string()))?;
    Ok(created_response(serde_json::json!({"message": "Mensagem adicionada"})))
}

pub async fn add_layout_version(
    State(state): State<AppState>,
    _user: Extension<UserContext>,
    Path(item_id): Path<i64>,
    Json(input): Json<LayoutVersionInput>,
) -> Result<impl axum::response::IntoResponse, AppError> {
    let id: i64 = sqlx::query_scalar(
        "INSERT INTO layout_versions (item_id, file_url, label) VALUES ($1,$2,$3) RETURNING id"
    )
    .bind(item_id).bind(&input.file_url).bind(&input.label).fetch_one(&state.db).await
    .map_err(|e| AppError::internal(e.to_string()))?;
    Ok(created_response(serde_json::json!({"id": id})))
}

pub async fn decide_layout(
    State(state): State<AppState>,
    _user: Extension<UserContext>,
    Path(version_id): Path<i64>,
    Json(input): Json<LayoutDecisionInput>,
) -> Result<impl axum::response::IntoResponse, AppError> {
    sqlx::query("UPDATE layout_versions SET status = $1, note = $2, decided_at = NOW() WHERE id = $3")
        .bind(&input.status).bind(&input.note).bind(version_id)
        .execute(&state.db).await
        .map_err(|e| AppError::internal(e.to_string()))?;
    Ok(ok_response(serde_json::json!({"message": "Decisão registrada"})))
}

pub async fn upsert_layout_job(
    State(state): State<AppState>,
    _user: Extension<UserContext>,
    Path((item_id, kind)): Path<(i64, String)>,
    Json(input): Json<LayoutJobInput>,
) -> Result<impl axum::response::IntoResponse, AppError> {
    sqlx::query(
        "INSERT INTO layout_jobs (item_id, kind, status, responsible_id, external_contact,
                due_at, file_url, tags, reason)
         VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)
         ON CONFLICT (item_id, kind) DO UPDATE SET
                status = EXCLUDED.status, responsible_id = EXCLUDED.responsible_id,
                external_contact = EXCLUDED.external_contact, due_at = EXCLUDED.due_at,
                file_url = EXCLUDED.file_url, tags = EXCLUDED.tags,
                reason = EXCLUDED.reason"
    )
    .bind(item_id).bind(&kind).bind(&input.status).bind(input.responsible_id)
    .bind(&input.external_contact).bind(input.due_at).bind(&input.file_url)
    .bind(&input.tags).bind(&input.reason)
    .execute(&state.db).await
    .map_err(|e| AppError::internal(e.to_string()))?;
    Ok(ok_response(serde_json::json!({"message": "Job salvo"})))
}

pub async fn confirm_layout_product(
    State(state): State<AppState>,
    _user: Extension<UserContext>,
    Path(item_id): Path<i64>,
) -> Result<impl axum::response::IntoResponse, AppError> {
    sqlx::query("UPDATE layout_request_items SET product_received = true, received_at = NOW() WHERE id = $1")
        .bind(item_id).execute(&state.db).await
        .map_err(|e| AppError::internal(e.to_string()))?;
    Ok(ok_response(serde_json::json!({"message": "Produto confirmado"})))
}
