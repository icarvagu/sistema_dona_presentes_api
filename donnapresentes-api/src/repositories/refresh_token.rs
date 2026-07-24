use crate::error::AppError;
use sqlx::PgPool;

pub async fn revoke_all(pool: &PgPool, user_id: i32) -> Result<(), AppError> {
    sqlx::query("UPDATE refresh_tokens SET revoked = true WHERE user_id = $1")
        .bind(user_id)
        .execute(pool)
        .await
        .map_err(|e| AppError::internal(e.to_string()))?;
    Ok(())
}
