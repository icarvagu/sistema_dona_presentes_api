use crate::models::{User, UserInput};

pub async fn get_all(pool: &sqlx::PgPool) -> Result<Vec<User>, crate::error::AppError> {
    sqlx::query_as::<_, User>(
        "SELECT id, username, password_hash, role, permissions, full_name, cpf, cpf_hash,
                rg, birth_date, gender, status, contact_email, full_address, contact_phone,
                notes, failed_login_attempts, locked_until, must_change_password,
                created_at, updated_at
         FROM users ORDER BY id",
    )
    .fetch_all(pool)
    .await
    .map_err(|e| crate::error::AppError::internal(e.to_string()))
}

pub async fn get_by_id(pool: &sqlx::PgPool, id: i32) -> Result<User, crate::error::AppError> {
    sqlx::query_as::<_, User>(
        "SELECT id, username, password_hash, role, permissions, full_name, cpf, cpf_hash,
                rg, birth_date, gender, status, contact_email, full_address, contact_phone,
                notes, failed_login_attempts, locked_until, must_change_password,
                created_at, updated_at
         FROM users WHERE id = $1",
    )
    .bind(id)
    .fetch_optional(pool)
    .await
    .map_err(|e| crate::error::AppError::internal(e.to_string()))?
    .ok_or_else(|| crate::error::AppError::not_found("Usuário"))
}

pub async fn create(
    pool: &sqlx::PgPool,
    input: &UserInput,
    password_hash: &str,
) -> Result<User, crate::error::AppError> {
    sqlx::query_as::<_, User>(
        "INSERT INTO users (username, password_hash, role, permissions, full_name, cpf,
                rg, birth_date, gender, status, contact_email, full_address, contact_phone,
                notes, must_change_password)
         VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
         RETURNING id, username, password_hash, role, permissions, full_name, cpf, cpf_hash,
                   rg, birth_date, gender, status, contact_email, full_address, contact_phone,
                   notes, failed_login_attempts, locked_until, must_change_password,
                   created_at, updated_at",
    )
    .bind(&input.username)
    .bind(password_hash)
    .bind(input.role.as_deref().unwrap_or("standard"))
    .bind(&input.permissions)
    .bind(&input.full_name)
    .bind(&input.cpf)
    .bind(&input.rg)
    .bind(input.birth_date)
    .bind(&input.gender)
    .bind(input.status.as_deref().unwrap_or("Ativo"))
    .bind(&input.contact_email)
    .bind(&input.full_address)
    .bind(&input.contact_phone)
    .bind(&input.notes)
    .bind(input.must_change_password)
    .fetch_one(pool)
    .await
    .map_err(|e| crate::error::AppError::internal(e.to_string()))
}

pub async fn update(
    pool: &sqlx::PgPool,
    id: i32,
    input: &UserInput,
) -> Result<User, crate::error::AppError> {
    sqlx::query_as::<_, User>(
        "UPDATE users SET
            full_name = $1, cpf = $2, rg = $3, birth_date = $4, gender = $5,
            status = $6, contact_email = $7, full_address = $8, contact_phone = $9,
            notes = $10, role = $11, permissions = $12, must_change_password = $13,
            updated_at = NOW()
         WHERE id = $14
         RETURNING id, username, password_hash, role, permissions, full_name, cpf, cpf_hash,
                   rg, birth_date, gender, status, contact_email, full_address, contact_phone,
                   notes, failed_login_attempts, locked_until, must_change_password,
                   created_at, updated_at",
    )
    .bind(&input.full_name)
    .bind(&input.cpf)
    .bind(&input.rg)
    .bind(input.birth_date)
    .bind(&input.gender)
    .bind(input.status.as_deref().unwrap_or("Ativo"))
    .bind(&input.contact_email)
    .bind(&input.full_address)
    .bind(&input.contact_phone)
    .bind(&input.notes)
    .bind(input.role.as_deref().unwrap_or("standard"))
    .bind(&input.permissions)
    .bind(input.must_change_password)
    .bind(id)
    .fetch_one(pool)
    .await
    .map_err(|e| crate::error::AppError::internal(e.to_string()))
}

pub async fn delete(pool: &sqlx::PgPool, id: i32) -> Result<(), crate::error::AppError> {
    sqlx::query("DELETE FROM users WHERE id = $1")
        .bind(id)
        .execute(pool)
        .await
        .map_err(|e| crate::error::AppError::internal(e.to_string()))?;
    Ok(())
}

pub async fn get_by_cpf(pool: &sqlx::PgPool, cpf: &str) -> Result<Option<User>, crate::error::AppError> {
    sqlx::query_as::<_, User>(
        "SELECT id, username, password_hash, role, permissions, full_name, cpf, cpf_hash,
                rg, birth_date, gender, status, contact_email, full_address, contact_phone,
                notes, failed_login_attempts, locked_until, must_change_password,
                created_at, updated_at
         FROM users WHERE cpf = $1",
    )
    .bind(cpf)
    .fetch_optional(pool)
    .await
    .map_err(|e| crate::error::AppError::internal(e.to_string()))
}
