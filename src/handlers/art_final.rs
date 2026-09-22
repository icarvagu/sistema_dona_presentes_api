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
    Extension(user): Extension<UserContext>,
) -> Result<impl axum::response::IntoResponse, AppError> {
    let task_count: i64 = sqlx::query_scalar("SELECT COUNT(*) FROM art_final_tasks")
        .fetch_one(&state.db).await
        .map_err(|e| AppError::internal(e.to_string()))?;
    let open_layouts: i64 = sqlx::query_scalar(
        "SELECT COUNT(*) FROM layout_requests WHERE status NOT IN ('approved','cancelled')",
    )
    .fetch_one(&state.db).await
    .map_err(|e| AppError::internal(e.to_string()))?;
    let approved_layouts: i64 = sqlx::query_scalar(
        "SELECT COUNT(*) FROM layout_requests WHERE status = 'approved'",
    )
    .fetch_one(&state.db).await
    .map_err(|e| AppError::internal(e.to_string()))?;
    let permissions = &user.permissions;
    let admin = user.role == "admin";
    Ok(ok_response(serde_json::json!({
        "total_tasks": task_count,
        "open_layouts": open_layouts,
        "approved_layouts": approved_layouts,
        "layout": admin || permissions.iter().any(|p| p == "arte_final"),
        "manage": admin || permissions.iter().any(|p| p == "arte_final"),
        "marketing": admin || permissions.iter().any(|p| p == "marketing" || p == "arte_final"),
        "corel": admin || permissions.iter().any(|p| p == "arte_final"),
        "engraving": admin || permissions.iter().any(|p| p == "producao" || p == "arte_final")
    })))
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
    let rows = sqlx::query(
        "SELECT r.*,
                u.username AS requested_name,
                COALESCE(i.item_count, 0)::BIGINT AS item_count,
                COALESCE(i.approved_items, 0)::BIGINT AS approved_items
         FROM layout_requests r
         LEFT JOIN users u ON u.id = r.requested_by
         LEFT JOIN (
            SELECT request_id,
                   COUNT(*) AS item_count,
                   COUNT(*) FILTER (WHERE status = 'approved') AS approved_items
            FROM layout_request_items
            GROUP BY request_id
         ) i ON i.request_id = r.id
         ORDER BY r.priority DESC, r.due_at NULLS LAST, r.id DESC"
    )
        .fetch_all(&state.db).await
        .map_err(|e| AppError::internal(e.to_string()))?;

    let requests: Vec<serde_json::Value> = rows.iter().map(|r| {
        let mut map = serde_json::Map::new();
        for col in r.columns() {
            let name = col.name().to_string();
            if let Ok(val) = r.try_get::<serde_json::Value, _>(name.as_str()) {
                map.insert(name, val);
            } else if let Ok(val) = r.try_get::<Option<chrono::DateTime<chrono::Utc>>, _>(name.as_str()) {
                map.insert(name, val.map(|v| serde_json::Value::String(v.to_rfc3339())).unwrap_or(serde_json::Value::Null));
            } else if let Ok(val) = r.try_get::<Option<i64>, _>(name.as_str()) {
                map.insert(name, val.map(serde_json::Value::from).unwrap_or(serde_json::Value::Null));
            } else if let Ok(val) = r.try_get::<Option<i32>, _>(name.as_str()) {
                map.insert(name, val.map(serde_json::Value::from).unwrap_or(serde_json::Value::Null));
            } else if let Ok(val) = r.try_get::<Option<String>, _>(name.as_str()) {
                map.insert(name, val.map(serde_json::Value::String).unwrap_or(serde_json::Value::Null));
            }
        }
        serde_json::Value::Object(map)
    }).collect();

    Ok(ok_response(requests))
}

pub async fn create_layout_request(
    State(state): State<AppState>,
    Extension(user): Extension<UserContext>,
    Json(input): Json<LayoutRequestInput>,
) -> Result<impl axum::response::IntoResponse, AppError> {
    let id: i64 = sqlx::query_scalar(
        "INSERT INTO layout_requests (source_type, source_id, title, instructions, file_mode,
                common_file_url, priority, due_at, assigned_to, requested_by, source_snapshot, status)
         VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,'requested') RETURNING id"
    )
    .bind(&input.source_type).bind(input.source_id).bind(&input.title)
    .bind(&input.instructions).bind(&input.file_mode).bind(&input.common_file_url)
    .bind(input.priority.unwrap_or(0)).bind(input.due_at).bind(input.assigned_to)
    .bind(user.user_id).bind(serde_json::json!({}))
    .fetch_one(&state.db).await
    .map_err(|e| AppError::internal(e.to_string()))?;

    for item in &input.items {
        sqlx::query(
            "INSERT INTO layout_request_items (request_id, entity_type, item_id, group_key, source_file_url, item_snapshot)
             VALUES ($1,$2,$3,$4,$5,$6)"
        )
        .bind(id).bind(&item.entity_type).bind(item.item_id)
        .bind(&item.group_key).bind(&item.source_file_url)
        .bind(serde_json::json!({}))
        .execute(&state.db).await.ok();
    }

    Ok(created_response(serde_json::json!({"id": id})))
}

pub async fn get_layout_request(
    State(state): State<AppState>,
    _user: Extension<UserContext>,
    Path(id): Path<i64>,
) -> Result<impl axum::response::IntoResponse, AppError> {
    let rows = sqlx::query(
        "SELECT r.*, u.username AS requested_name
         FROM layout_requests r
         LEFT JOIN users u ON u.id = r.requested_by
         WHERE r.id = $1"
    )
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
        if let Ok(val) = r.try_get::<serde_json::Value, _>(name.as_str()) {
            map.insert(name, val);
        } else if let Ok(val) = r.try_get::<Option<chrono::DateTime<chrono::Utc>>, _>(name.as_str()) {
            map.insert(name, val.map(|v| serde_json::Value::String(v.to_rfc3339())).unwrap_or(serde_json::Value::Null));
        } else if let Ok(val) = r.try_get::<Option<i64>, _>(name.as_str()) {
            map.insert(name, val.map(serde_json::Value::from).unwrap_or(serde_json::Value::Null));
        } else if let Ok(val) = r.try_get::<Option<i32>, _>(name.as_str()) {
            map.insert(name, val.map(serde_json::Value::from).unwrap_or(serde_json::Value::Null));
        } else if let Ok(val) = r.try_get::<Option<String>, _>(name.as_str()) {
            map.insert(name, val.map(serde_json::Value::String).unwrap_or(serde_json::Value::Null));
        }
    }

    let items = sqlx::query(
        "SELECT i.*,
                COALESCE((
                    SELECT json_agg(json_build_object(
                        'id', v.id,
                        'version', v.version,
                        'label', v.label,
                        'file_url', v.file_url,
                        'approval_status', v.approval_status,
                        'approval_note', v.approval_note,
                        'created_at', v.created_at
                    ) ORDER BY v.version DESC)
                    FROM item_layout_versions v
                    WHERE v.request_item_id = i.id
                ), '[]'::json) AS versions,
                row_to_json(c.*) AS corel,
                row_to_json(e.*) AS engraving
         FROM layout_request_items i
         LEFT JOIN layout_corel_jobs c ON c.request_item_id = i.id
         LEFT JOIN layout_engraving_jobs e ON e.request_item_id = i.id
         WHERE i.request_id = $1
         ORDER BY i.id"
    )
    .bind(id)
    .fetch_all(&state.db).await
    .map_err(|e| AppError::internal(e.to_string()))?;

    let items_json: Vec<serde_json::Value> = items.iter().map(|row| {
        let mut item = serde_json::Map::new();
        for col in row.columns() {
            let name = col.name().to_string();
            if let Ok(val) = row.try_get::<serde_json::Value, _>(name.as_str()) {
                item.insert(name, val);
            } else if let Ok(val) = row.try_get::<Option<chrono::DateTime<chrono::Utc>>, _>(name.as_str()) {
                item.insert(name, val.map(|v| serde_json::Value::String(v.to_rfc3339())).unwrap_or(serde_json::Value::Null));
            } else if let Ok(val) = row.try_get::<Option<i64>, _>(name.as_str()) {
                item.insert(name, val.map(serde_json::Value::from).unwrap_or(serde_json::Value::Null));
            } else if let Ok(val) = row.try_get::<Option<i32>, _>(name.as_str()) {
                item.insert(name, val.map(serde_json::Value::from).unwrap_or(serde_json::Value::Null));
            } else if let Ok(val) = row.try_get::<Option<String>, _>(name.as_str()) {
                item.insert(name, val.map(serde_json::Value::String).unwrap_or(serde_json::Value::Null));
            }
        }
        serde_json::Value::Object(item)
    }).collect();

    let timeline = sqlx::query(
        "SELECT id, action, details, created_at
         FROM art_final_audit_log
         WHERE entity_type = 'layout_request' AND entity_id = $1
         ORDER BY created_at DESC"
    )
    .bind(id)
    .fetch_all(&state.db).await
    .map_err(|e| AppError::internal(e.to_string()))?;
    let timeline_json: Vec<serde_json::Value> = timeline.iter().map(|row| serde_json::json!({
        "id": row.try_get::<i64, _>("id").unwrap_or_default(),
        "action": row.try_get::<String, _>("action").unwrap_or_default(),
        "details": row.try_get::<serde_json::Value, _>("details").unwrap_or_else(|_| serde_json::json!({})),
        "created_at": row.try_get::<chrono::DateTime<chrono::Utc>, _>("created_at").map(|v| v.to_rfc3339()).unwrap_or_default()
    })).collect();

    Ok(ok_response(serde_json::json!({
        "request": serde_json::Value::Object(map),
        "items": items_json,
        "timeline": timeline_json
    })))
}

pub async fn transition_layout(
    State(state): State<AppState>,
    Extension(user): Extension<UserContext>,
    Path(id): Path<i64>,
    Json(input): Json<LayoutTransitionInput>,
) -> Result<impl axum::response::IntoResponse, AppError> {
    sqlx::query("UPDATE layout_requests SET status = $1, updated_at = NOW() WHERE id = $2")
        .bind(&input.status).bind(id)
        .execute(&state.db).await
        .map_err(|e| AppError::internal(e.to_string()))?;
    sqlx::query(
        "INSERT INTO art_final_audit_log (entity_type, entity_id, action, details, user_id)
         VALUES ('layout_request', $1, 'transition', $2, $3)"
    )
    .bind(id)
    .bind(serde_json::json!({"status": input.status, "note": input.note}))
    .bind(user.user_id)
    .execute(&state.db).await.ok();
    Ok(ok_response(serde_json::json!({"message": "Status atualizado"})))
}

pub async fn add_layout_message(
    State(state): State<AppState>,
    Extension(user): Extension<UserContext>,
    Path(id): Path<i64>,
    Json(input): Json<LayoutMessageInput>,
) -> Result<impl axum::response::IntoResponse, AppError> {
    sqlx::query(
        "INSERT INTO art_final_audit_log (entity_type, entity_id, action, details, user_id)
         VALUES ('layout_request', $1, 'message', $2, $3)"
    )
    .bind(id)
    .bind(serde_json::json!({"message": input.message, "file_url": input.file_url}))
    .bind(user.user_id)
    .execute(&state.db).await
    .map_err(|e| AppError::internal(e.to_string()))?;
    Ok(created_response(serde_json::json!({"message": "Mensagem adicionada"})))
}

pub async fn add_layout_version(
    State(state): State<AppState>,
    Extension(user): Extension<UserContext>,
    Path(item_id): Path<i64>,
    Json(input): Json<LayoutVersionInput>,
) -> Result<impl axum::response::IntoResponse, AppError> {
    let id: i64 = sqlx::query_scalar(
        "INSERT INTO item_layout_versions (entity_type, item_id, request_item_id, version, label, file_url, created_by)
         SELECT i.entity_type, i.item_id, i.id,
                COALESCE((SELECT MAX(version) FROM item_layout_versions WHERE request_item_id = i.id), 0) + 1,
                COALESCE($2, 'Layout'), $3, $4
         FROM layout_request_items i
         WHERE i.id = $1
         RETURNING id"
    )
    .bind(item_id).bind(&input.label).bind(&input.file_url).bind(user.user_id)
    .fetch_one(&state.db).await
    .map_err(|e| AppError::internal(e.to_string()))?;
    Ok(created_response(serde_json::json!({"id": id})))
}

pub async fn decide_layout(
    State(state): State<AppState>,
    Extension(user): Extension<UserContext>,
    Path(version_id): Path<i64>,
    Json(input): Json<LayoutDecisionInput>,
) -> Result<impl axum::response::IntoResponse, AppError> {
    sqlx::query("UPDATE item_layout_versions SET approval_status = $1, approval_note = COALESCE($2, ''), approved_by = $3, approved_at = NOW() WHERE id = $4")
        .bind(&input.status).bind(&input.note).bind(user.user_id).bind(version_id)
        .execute(&state.db).await
        .map_err(|e| AppError::internal(e.to_string()))?;
    sqlx::query(
        "INSERT INTO layout_approval_events (layout_version_id, status, note, created_by)
         VALUES ($1, $2, COALESCE($3, ''), $4)"
    )
    .bind(version_id).bind(&input.status).bind(&input.note).bind(user.user_id)
    .execute(&state.db).await.ok();
    Ok(ok_response(serde_json::json!({"message": "Decisão registrada"})))
}

pub async fn upsert_layout_job(
    State(state): State<AppState>,
    Extension(user): Extension<UserContext>,
    Path((item_id, kind)): Path<(i64, String)>,
    Json(input): Json<LayoutJobInput>,
) -> Result<impl axum::response::IntoResponse, AppError> {
    if kind == "corel" {
        sqlx::query(
            "INSERT INTO layout_corel_jobs (request_item_id, status, responsible_id, external_contact,
                    due_at, file_url, tags, reason, updated_by)
             VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)
             ON CONFLICT (request_item_id) DO UPDATE SET
                    status = EXCLUDED.status, responsible_id = EXCLUDED.responsible_id,
                    external_contact = EXCLUDED.external_contact, due_at = EXCLUDED.due_at,
                    file_url = EXCLUDED.file_url, tags = EXCLUDED.tags,
                    reason = EXCLUDED.reason, updated_by = EXCLUDED.updated_by, updated_at = NOW()",
        )
        .bind(item_id).bind(&input.status).bind(input.responsible_id)
        .bind(&input.external_contact).bind(input.due_at).bind(&input.file_url)
        .bind(&input.tags).bind(&input.reason).bind(user.user_id)
        .execute(&state.db).await
        .map_err(|e| AppError::internal(e.to_string()))?;
    } else {
        sqlx::query(
            "INSERT INTO layout_engraving_jobs (request_item_id, status, responsible_id,
                    due_at, file_url, tags, reason, updated_by)
             VALUES ($1,$2,$3,$4,$5,$6,$7,$8)
             ON CONFLICT (request_item_id) DO UPDATE SET
                    status = EXCLUDED.status, responsible_id = EXCLUDED.responsible_id,
                    due_at = EXCLUDED.due_at, file_url = EXCLUDED.file_url,
                    tags = EXCLUDED.tags, reason = EXCLUDED.reason,
                    updated_by = EXCLUDED.updated_by, updated_at = NOW()",
        )
        .bind(item_id).bind(&input.status).bind(input.responsible_id)
        .bind(input.due_at).bind(&input.file_url).bind(&input.tags)
        .bind(&input.reason).bind(user.user_id)
        .execute(&state.db).await
        .map_err(|e| AppError::internal(e.to_string()))?;
    }
    Ok(ok_response(serde_json::json!({"message": "Job salvo"})))
}

pub async fn confirm_layout_product(
    State(state): State<AppState>,
    Extension(user): Extension<UserContext>,
    Path(item_id): Path<i64>,
) -> Result<impl axum::response::IntoResponse, AppError> {
    sqlx::query("UPDATE layout_request_items SET product_received_at = NOW(), product_received_by = $1 WHERE id = $2")
        .bind(user.user_id).bind(item_id).execute(&state.db).await
        .map_err(|e| AppError::internal(e.to_string()))?;
    Ok(ok_response(serde_json::json!({"message": "Produto confirmado"})))
}
