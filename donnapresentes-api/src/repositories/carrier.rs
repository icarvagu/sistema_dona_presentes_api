use crate::models::{Carrier, CarrierInput};

pub async fn get_all(pool: &sqlx::PgPool) -> Result<Vec<Carrier>, crate::error::AppError> {
    sqlx::query_as::<_, Carrier>(
        "SELECT id, name, cnpj, carrier_type, email, landline_phone, mobile_phone,
                full_address, contact_name, contact_phone, website, created_at, updated_at
         FROM carriers ORDER BY id",
    )
    .fetch_all(pool)
    .await
    .map_err(|e| crate::error::AppError::internal(e.to_string()))
}

pub async fn get_by_id(pool: &sqlx::PgPool, id: i32) -> Result<Carrier, crate::error::AppError> {
    sqlx::query_as::<_, Carrier>(
        "SELECT id, name, cnpj, carrier_type, email, landline_phone, mobile_phone,
                full_address, contact_name, contact_phone, website, created_at, updated_at
         FROM carriers WHERE id = $1",
    )
    .bind(id)
    .fetch_optional(pool)
    .await
    .map_err(|e| crate::error::AppError::internal(e.to_string()))?
    .ok_or_else(|| crate::error::AppError::not_found("Transportadora"))
}

pub async fn create(
    pool: &sqlx::PgPool,
    input: &CarrierInput,
) -> Result<Carrier, crate::error::AppError> {
    sqlx::query_as::<_, Carrier>(
        "INSERT INTO carriers (name, cnpj, carrier_type, email, landline_phone, mobile_phone,
                full_address, contact_name, contact_phone, website)
         VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
         RETURNING id, name, cnpj, carrier_type, email, landline_phone, mobile_phone,
                   full_address, contact_name, contact_phone, website, created_at, updated_at",
    )
    .bind(&input.name)
    .bind(&input.cnpj)
    .bind(&input.carrier_type)
    .bind(&input.email)
    .bind(&input.landline_phone)
    .bind(&input.mobile_phone)
    .bind(&input.full_address)
    .bind(&input.contact_name)
    .bind(&input.contact_phone)
    .bind(&input.website)
    .fetch_one(pool)
    .await
    .map_err(|e| crate::error::AppError::internal(e.to_string()))
}

pub async fn update(
    pool: &sqlx::PgPool,
    id: i32,
    input: &CarrierInput,
) -> Result<Carrier, crate::error::AppError> {
    sqlx::query_as::<_, Carrier>(
        "UPDATE carriers SET name = $1, cnpj = $2, carrier_type = $3, email = $4,
                landline_phone = $5, mobile_phone = $6, full_address = $7, contact_name = $8,
                contact_phone = $9, website = $10, updated_at = NOW()
         WHERE id = $11
         RETURNING id, name, cnpj, carrier_type, email, landline_phone, mobile_phone,
                   full_address, contact_name, contact_phone, website, created_at, updated_at",
    )
    .bind(&input.name)
    .bind(&input.cnpj)
    .bind(&input.carrier_type)
    .bind(&input.email)
    .bind(&input.landline_phone)
    .bind(&input.mobile_phone)
    .bind(&input.full_address)
    .bind(&input.contact_name)
    .bind(&input.contact_phone)
    .bind(&input.website)
    .bind(id)
    .fetch_one(pool)
    .await
    .map_err(|e| crate::error::AppError::internal(e.to_string()))
}

pub async fn delete(pool: &sqlx::PgPool, id: i32) -> Result<(), crate::error::AppError> {
    sqlx::query("DELETE FROM carriers WHERE id = $1")
        .bind(id)
        .execute(pool)
        .await
        .map_err(|e| crate::error::AppError::internal(e.to_string()))?;
    Ok(())
}
