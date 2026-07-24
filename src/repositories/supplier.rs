use crate::models::{Supplier, SupplierInput};

pub async fn get_all(pool: &sqlx::PgPool) -> Result<Vec<Supplier>, crate::error::AppError> {
    sqlx::query_as::<_, Supplier>(
        "SELECT id, name, cnpj, state_registration, contact_person, email, landline_phone,
                mobile_phone, responsible_email, commercial_address, postal_code, website,
                created_at, updated_at
         FROM suppliers ORDER BY id",
    )
    .fetch_all(pool)
    .await
    .map_err(|e| crate::error::AppError::internal(e.to_string()))
}

pub async fn get_by_id(pool: &sqlx::PgPool, id: i32) -> Result<Supplier, crate::error::AppError> {
    sqlx::query_as::<_, Supplier>(
        "SELECT id, name, cnpj, state_registration, contact_person, email, landline_phone,
                mobile_phone, responsible_email, commercial_address, postal_code, website,
                created_at, updated_at
         FROM suppliers WHERE id = $1",
    )
    .bind(id)
    .fetch_optional(pool)
    .await
    .map_err(|e| crate::error::AppError::internal(e.to_string()))?
    .ok_or_else(|| crate::error::AppError::not_found("Fornecedor"))
}

pub async fn create(
    pool: &sqlx::PgPool,
    input: &SupplierInput,
) -> Result<Supplier, crate::error::AppError> {
    sqlx::query_as::<_, Supplier>(
        "INSERT INTO suppliers (name, cnpj, state_registration, contact_person, email,
                landline_phone, mobile_phone, responsible_email, commercial_address,
                postal_code, website)
         VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
         RETURNING id, name, cnpj, state_registration, contact_person, email, landline_phone,
                   mobile_phone, responsible_email, commercial_address, postal_code, website,
                   created_at, updated_at",
    )
    .bind(&input.name)
    .bind(&input.cnpj)
    .bind(&input.state_registration)
    .bind(&input.contact_person)
    .bind(&input.email)
    .bind(&input.landline_phone)
    .bind(&input.mobile_phone)
    .bind(&input.responsible_email)
    .bind(&input.commercial_address)
    .bind(&input.postal_code)
    .bind(&input.website)
    .fetch_one(pool)
    .await
    .map_err(|e| crate::error::AppError::internal(e.to_string()))
}

pub async fn update(
    pool: &sqlx::PgPool,
    id: i32,
    input: &SupplierInput,
) -> Result<Supplier, crate::error::AppError> {
    sqlx::query_as::<_, Supplier>(
        "UPDATE suppliers SET name = $1, cnpj = $2, state_registration = $3, contact_person = $4,
                email = $5, landline_phone = $6, mobile_phone = $7, responsible_email = $8,
                commercial_address = $9, postal_code = $10, website = $11, updated_at = NOW()
         WHERE id = $12
         RETURNING id, name, cnpj, state_registration, contact_person, email, landline_phone,
                   mobile_phone, responsible_email, commercial_address, postal_code, website,
                   created_at, updated_at",
    )
    .bind(&input.name)
    .bind(&input.cnpj)
    .bind(&input.state_registration)
    .bind(&input.contact_person)
    .bind(&input.email)
    .bind(&input.landline_phone)
    .bind(&input.mobile_phone)
    .bind(&input.responsible_email)
    .bind(&input.commercial_address)
    .bind(&input.postal_code)
    .bind(&input.website)
    .bind(id)
    .fetch_one(pool)
    .await
    .map_err(|e| crate::error::AppError::internal(e.to_string()))
}

pub async fn delete(pool: &sqlx::PgPool, id: i32) -> Result<(), crate::error::AppError> {
    sqlx::query("DELETE FROM suppliers WHERE id = $1")
        .bind(id)
        .execute(pool)
        .await
        .map_err(|e| crate::error::AppError::internal(e.to_string()))?;
    Ok(())
}
