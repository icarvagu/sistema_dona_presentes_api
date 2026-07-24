use crate::models::{AdditionalContact, Address, Customer, CustomerInput};

pub async fn get_all(pool: &sqlx::PgPool) -> Result<Vec<Customer>, crate::error::AppError> {
    let customers = sqlx::query_as::<_, Customer>(
        "SELECT id, customer_type, status, name, trade_name, company_name, state_registration,
                city_registration, responsible, contact_financial_name, contact_financial_email,
                contact_financial_phone, contact_nf_name, contact_nf_email, contact_nf_phone,
                contact_commercial_name, contact_commercial_email, contact_commercial_phone,
                cnpj, cpf, email, business_phone, mobile_phone, website, notes, created_at, updated_at
         FROM customers ORDER BY id",
    )
    .fetch_all(pool)
    .await
    .map_err(|e| crate::error::AppError::internal(e.to_string()))?;

    let ids: Vec<i32> = customers.iter().map(|c| c.id).collect();
    let addresses = get_addresses_bulk(pool, &ids).await?;
    let contacts = get_contacts_bulk(pool, &ids).await?;

    Ok(customers
        .into_iter()
        .map(|mut c| {
            let id = c.id;
            c.addresses = addresses.get(&id).cloned().unwrap_or_default();
            c.additional_contacts = contacts.get(&id).cloned().unwrap_or_default();
            c
        })
        .collect())
}

pub async fn get_by_id(pool: &sqlx::PgPool, id: i32) -> Result<Customer, crate::error::AppError> {
    let mut customer = sqlx::query_as::<_, Customer>(
        "SELECT id, customer_type, status, name, trade_name, company_name, state_registration,
                city_registration, responsible, contact_financial_name, contact_financial_email,
                contact_financial_phone, contact_nf_name, contact_nf_email, contact_nf_phone,
                contact_commercial_name, contact_commercial_email, contact_commercial_phone,
                cnpj, cpf, email, business_phone, mobile_phone, website, notes, created_at, updated_at
         FROM customers WHERE id = $1",
    )
    .bind(id)
    .fetch_optional(pool)
    .await
    .map_err(|e| crate::error::AppError::internal(e.to_string()))?
    .ok_or_else(|| crate::error::AppError::not_found("Cliente"))?;

    customer.addresses = get_addresses(pool, id).await?;
    customer.additional_contacts = get_contacts(pool, id).await?;
    Ok(customer)
}

pub async fn create(
    pool: &sqlx::PgPool,
    input: &CustomerInput,
) -> Result<Customer, crate::error::AppError> {
    let mut tx = pool.begin().await.map_err(|e| crate::error::AppError::internal(e.to_string()))?;

    let customer = sqlx::query_as::<_, Customer>(
        "INSERT INTO customers (customer_type, status, name, trade_name, company_name,
                state_registration, city_registration, responsible, contact_financial_name,
                contact_financial_email, contact_financial_phone, contact_nf_name,
                contact_nf_email, contact_nf_phone, contact_commercial_name,
                contact_commercial_email, contact_commercial_phone, cnpj, cpf, email,
                business_phone, mobile_phone, website, notes)
         VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17,
                 $18, $19, $20, $21, $22, $23, $24)
         RETURNING id, customer_type, status, name, trade_name, company_name, state_registration,
                   city_registration, responsible, contact_financial_name, contact_financial_email,
                   contact_financial_phone, contact_nf_name, contact_nf_email, contact_nf_phone,
                   contact_commercial_name, contact_commercial_email, contact_commercial_phone,
                   cnpj, cpf, email, business_phone, mobile_phone, website, notes,
                   created_at, updated_at",
    )
    .bind(&input.customer_type)
    .bind(input.status.as_deref().unwrap_or("Ativo"))
    .bind(&input.name)
    .bind(&input.trade_name)
    .bind(&input.company_name)
    .bind(&input.state_registration)
    .bind(&input.city_registration)
    .bind(&input.responsible)
    .bind(&input.contact_financial_name)
    .bind(&input.contact_financial_email)
    .bind(&input.contact_financial_phone)
    .bind(&input.contact_nf_name)
    .bind(&input.contact_nf_email)
    .bind(&input.contact_nf_phone)
    .bind(&input.contact_commercial_name)
    .bind(&input.contact_commercial_email)
    .bind(&input.contact_commercial_phone)
    .bind(&input.cnpj)
    .bind(&input.cpf)
    .bind(&input.email)
    .bind(&input.business_phone)
    .bind(&input.mobile_phone)
    .bind(&input.website)
    .bind(&input.notes)
    .fetch_one(&mut *tx)
    .await
    .map_err(|e| crate::error::AppError::internal(e.to_string()))?;

    for addr in &input.addresses {
        sqlx::query(
            "INSERT INTO customer_addresses (customer_id, address_type, address)
             VALUES ($1, $2, $3)",
        )
        .bind(customer.id)
        .bind(&addr.address_type)
        .bind(&addr.address)
        .execute(&mut *tx)
        .await
        .map_err(|e| crate::error::AppError::internal(e.to_string()))?;
    }

    for contact in &input.additional_contacts {
        sqlx::query(
            "INSERT INTO additional_contacts (customer_id, name, email, phone)
             VALUES ($1, $2, $3, $4)",
        )
        .bind(customer.id)
        .bind(&contact.name)
        .bind(&contact.email)
        .bind(&contact.phone)
        .execute(&mut *tx)
        .await
        .map_err(|e| crate::error::AppError::internal(e.to_string()))?;
    }

    tx.commit().await.map_err(|e| crate::error::AppError::internal(e.to_string()))?;

    let mut c = customer;
    c.addresses = get_addresses(pool, c.id).await?;
    c.additional_contacts = get_contacts(pool, c.id).await?;
    Ok(c)
}

pub async fn update(
    pool: &sqlx::PgPool,
    id: i32,
    input: &CustomerInput,
) -> Result<Customer, crate::error::AppError> {
    let mut tx = pool.begin().await.map_err(|e| crate::error::AppError::internal(e.to_string()))?;

    let customer = sqlx::query_as::<_, Customer>(
        "UPDATE customers SET customer_type = $1, status = $2, name = $3, trade_name = $4,
                company_name = $5, state_registration = $6, city_registration = $7,
                responsible = $8, contact_financial_name = $9, contact_financial_email = $10,
                contact_financial_phone = $11, contact_nf_name = $12, contact_nf_email = $13,
                contact_nf_phone = $14, contact_commercial_name = $15,
                contact_commercial_email = $16, contact_commercial_phone = $17,
                cnpj = $18, cpf = $19, email = $20, business_phone = $21, mobile_phone = $22,
                website = $23, notes = $24, updated_at = NOW()
         WHERE id = $25
         RETURNING id, customer_type, status, name, trade_name, company_name, state_registration,
                   city_registration, responsible, contact_financial_name, contact_financial_email,
                   contact_financial_phone, contact_nf_name, contact_nf_email, contact_nf_phone,
                   contact_commercial_name, contact_commercial_email, contact_commercial_phone,
                   cnpj, cpf, email, business_phone, mobile_phone, website, notes,
                   created_at, updated_at",
    )
    .bind(&input.customer_type)
    .bind(input.status.as_deref().unwrap_or("Ativo"))
    .bind(&input.name)
    .bind(&input.trade_name)
    .bind(&input.company_name)
    .bind(&input.state_registration)
    .bind(&input.city_registration)
    .bind(&input.responsible)
    .bind(&input.contact_financial_name)
    .bind(&input.contact_financial_email)
    .bind(&input.contact_financial_phone)
    .bind(&input.contact_nf_name)
    .bind(&input.contact_nf_email)
    .bind(&input.contact_nf_phone)
    .bind(&input.contact_commercial_name)
    .bind(&input.contact_commercial_email)
    .bind(&input.contact_commercial_phone)
    .bind(&input.cnpj)
    .bind(&input.cpf)
    .bind(&input.email)
    .bind(&input.business_phone)
    .bind(&input.mobile_phone)
    .bind(&input.website)
    .bind(&input.notes)
    .bind(id)
    .fetch_one(&mut *tx)
    .await
    .map_err(|e| crate::error::AppError::internal(e.to_string()))?;

    sqlx::query("DELETE FROM customer_addresses WHERE customer_id = $1")
        .bind(id)
        .execute(&mut *tx)
        .await
        .map_err(|e| crate::error::AppError::internal(e.to_string()))?;

    sqlx::query("DELETE FROM additional_contacts WHERE customer_id = $1")
        .bind(id)
        .execute(&mut *tx)
        .await
        .map_err(|e| crate::error::AppError::internal(e.to_string()))?;

    for addr in &input.addresses {
        sqlx::query(
            "INSERT INTO customer_addresses (customer_id, address_type, address)
             VALUES ($1, $2, $3)",
        )
        .bind(id)
        .bind(&addr.address_type)
        .bind(&addr.address)
        .execute(&mut *tx)
        .await
        .map_err(|e| crate::error::AppError::internal(e.to_string()))?;
    }

    for contact in &input.additional_contacts {
        sqlx::query(
            "INSERT INTO additional_contacts (customer_id, name, email, phone)
             VALUES ($1, $2, $3, $4)",
        )
        .bind(id)
        .bind(&contact.name)
        .bind(&contact.email)
        .bind(&contact.phone)
        .execute(&mut *tx)
        .await
        .map_err(|e| crate::error::AppError::internal(e.to_string()))?;
    }

    tx.commit().await.map_err(|e| crate::error::AppError::internal(e.to_string()))?;

    let mut c = customer;
    c.addresses = get_addresses(pool, c.id).await?;
    c.additional_contacts = get_contacts(pool, c.id).await?;
    Ok(c)
}

pub async fn delete(pool: &sqlx::PgPool, id: i32) -> Result<(), crate::error::AppError> {
    sqlx::query("DELETE FROM customers WHERE id = $1")
        .bind(id)
        .execute(pool)
        .await
        .map_err(|e| crate::error::AppError::internal(e.to_string()))?;
    Ok(())
}

async fn get_addresses(
    pool: &sqlx::PgPool,
    customer_id: i32,
) -> Result<Vec<Address>, crate::error::AppError> {
    sqlx::query_as::<_, Address>(
        "SELECT id, customer_id, address_type, address, created_at, updated_at
         FROM customer_addresses WHERE customer_id = $1",
    )
    .bind(customer_id)
    .fetch_all(pool)
    .await
    .map_err(|e| crate::error::AppError::internal(e.to_string()))
}

async fn get_contacts(
    pool: &sqlx::PgPool,
    customer_id: i32,
) -> Result<Vec<AdditionalContact>, crate::error::AppError> {
    sqlx::query_as::<_, AdditionalContact>(
        "SELECT id, customer_id, name, email, phone, created_at, updated_at
         FROM additional_contacts WHERE customer_id = $1",
    )
    .bind(customer_id)
    .fetch_all(pool)
    .await
    .map_err(|e| crate::error::AppError::internal(e.to_string()))
}

async fn get_addresses_bulk(
    pool: &sqlx::PgPool,
    ids: &[i32],
) -> Result<std::collections::HashMap<i32, Vec<Address>>, crate::error::AppError> {
    let rows = sqlx::query_as::<_, Address>(
        "SELECT id, customer_id, address_type, address, created_at, updated_at
         FROM customer_addresses WHERE customer_id = ANY($1)",
    )
    .bind(ids)
    .fetch_all(pool)
    .await
    .map_err(|e| crate::error::AppError::internal(e.to_string()))?;

    let mut map: std::collections::HashMap<i32, Vec<Address>> = std::collections::HashMap::new();
    for row in rows {
        map.entry(row.customer_id).or_default().push(row);
    }
    Ok(map)
}

async fn get_contacts_bulk(
    pool: &sqlx::PgPool,
    ids: &[i32],
) -> Result<std::collections::HashMap<i32, Vec<AdditionalContact>>, crate::error::AppError> {
    let rows = sqlx::query_as::<_, AdditionalContact>(
        "SELECT id, customer_id, name, email, phone, created_at, updated_at
         FROM additional_contacts WHERE customer_id = ANY($1)",
    )
    .bind(ids)
    .fetch_all(pool)
    .await
    .map_err(|e| crate::error::AppError::internal(e.to_string()))?;

    let mut map: std::collections::HashMap<i32, Vec<AdditionalContact>> =
        std::collections::HashMap::new();
    for row in rows {
        map.entry(row.customer_id).or_default().push(row);
    }
    Ok(map)
}
