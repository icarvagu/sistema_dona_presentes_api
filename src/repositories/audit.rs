pub async fn log(
    _pool: &sqlx::PgPool,
    _user_id: Option<i32>,
    _action: &str,
    _entity: &str,
    _detail: &str,
    _ip: &str,
) {
}
