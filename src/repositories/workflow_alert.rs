use crate::error::AppError;
use crate::models::{WorkflowAlert, WorkflowAlertInput};

pub async fn list_by_user(
    pool: &sqlx::PgPool,
    user_id: i32,
) -> Result<Vec<WorkflowAlert>, AppError> {
    sqlx::query_as::<_, WorkflowAlert>(
        "SELECT id, user_id, sale_id, quote_id, title, description, scheduled_at,
                status, resolved_at, created_at, updated_at
         FROM workflow_alerts
         WHERE user_id = $1
         ORDER BY CASE WHEN status = 'scheduled' THEN 0 ELSE 1 END,
                  scheduled_at ASC,
                  created_at DESC",
    )
    .bind(user_id)
    .fetch_all(pool)
    .await
    .map_err(|e| AppError::internal(e.to_string()))
}

pub async fn create(
    pool: &sqlx::PgPool,
    user_id: i32,
    input: &WorkflowAlertInput,
) -> Result<WorkflowAlert, AppError> {
    sqlx::query_as::<_, WorkflowAlert>(
        "INSERT INTO workflow_alerts (user_id, sale_id, quote_id, title, description, scheduled_at)
         VALUES ($1, $2, $3, $4, $5, $6)
         RETURNING id, user_id, sale_id, quote_id, title, description, scheduled_at,
                   status, resolved_at, created_at, updated_at",
    )
    .bind(user_id)
    .bind(input.sale_id)
    .bind(input.quote_id)
    .bind(&input.title)
    .bind(&input.description)
    .bind(input.scheduled_at)
    .fetch_one(pool)
    .await
    .map_err(|e| AppError::internal(e.to_string()))
}

pub async fn resolve(
    pool: &sqlx::PgPool,
    user_id: i32,
    alert_id: i64,
) -> Result<WorkflowAlert, AppError> {
    sqlx::query_as::<_, WorkflowAlert>(
        "UPDATE workflow_alerts
         SET status = 'done',
             resolved_at = COALESCE(resolved_at, NOW()),
             updated_at = NOW()
         WHERE id = $1
           AND user_id = $2
         RETURNING id, user_id, sale_id, quote_id, title, description, scheduled_at,
                   status, resolved_at, created_at, updated_at",
    )
    .bind(alert_id)
    .bind(user_id)
    .fetch_optional(pool)
    .await
    .map_err(|e| AppError::internal(e.to_string()))?
    .ok_or_else(|| AppError::not_found("Alerta"))
}

pub async fn delete(
    pool: &sqlx::PgPool,
    user_id: i32,
    alert_id: i64,
) -> Result<(), AppError> {
    let result = sqlx::query(
        "DELETE FROM workflow_alerts
         WHERE id = $1
           AND user_id = $2",
    )
    .bind(alert_id)
    .bind(user_id)
    .execute(pool)
    .await
    .map_err(|e| AppError::internal(e.to_string()))?;

    if result.rows_affected() == 0 {
        return Err(AppError::not_found("Alerta"));
    }

    Ok(())
}
