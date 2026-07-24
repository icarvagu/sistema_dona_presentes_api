use sqlx::PgPool;

pub struct AuditService;

impl AuditService {
    pub async fn log(
        pool: &PgPool,
        user_id: Option<i32>,
        action: &str,
        entity: &str,
        detail: &str,
        ip: &str,
    ) {
        let _ = sqlx::query(
            "INSERT INTO audit_logs (user_id, action, entity, detail, ip_address, created_at)
             VALUES ($1, $2, $3, $4, $5, NOW())",
        )
        .bind(user_id)
        .bind(action)
        .bind(entity)
        .bind(detail)
        .bind(ip)
        .execute(pool)
        .await;
    }
}
