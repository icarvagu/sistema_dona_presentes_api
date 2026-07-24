use crate::error::AppError;
use crate::models::{Sale, SaleInput, SaleItem};

pub async fn get_all(pool: &sqlx::PgPool) -> Result<Vec<Sale>, AppError> {
    let sales = sqlx::query_as::<_, Sale>(SALE_SELECT_ALL)
        .fetch_all(pool)
        .await
        .map_err(|e| AppError::internal(e.to_string()))?;
    attach_items_bulk(pool, sales).await
}

pub async fn get_by_id(pool: &sqlx::PgPool, id: i32) -> Result<Sale, AppError> {
    let mut sale = sqlx::query_as::<_, Sale>(&format!("{SALE_SELECT_ALL} WHERE s.id = $1"))
        .bind(id)
        .fetch_optional(pool)
        .await
        .map_err(|e| AppError::internal(e.to_string()))?
        .ok_or_else(|| AppError::not_found("Venda"))?;
    sale.items = sqlx::query_as::<_, SaleItem>(
        "SELECT id, sale_id, product_id, quantity, unit_price, total_price,
                discount, price_formation, engravings, created_at, updated_at
         FROM sale_items WHERE sale_id = $1 ORDER BY id",
    )
    .bind(id)
    .fetch_all(pool)
    .await
    .map_err(|e| AppError::internal(e.to_string()))?;
    sale.carriers = get_carriers(pool, id).await?;
    Ok(sale)
}

pub async fn create(pool: &sqlx::PgPool, input: &SaleInput) -> Result<Sale, AppError> {
    let mut tx = pool.begin().await.map_err(|e| AppError::internal(e.to_string()))?;

    let total: f64 = input.items.iter().map(|i| i.unit_price * i.quantity as f64 * (1.0 - i.discount.unwrap_or(0.0))).sum();

    let sale = sqlx::query_as::<_, Sale>(
        "INSERT INTO sales (seller_id, customer_id, payment_method, installments,
                payment_term_days, first_installment_start, installment_dates,
                status, is_event, delivery_address, delivery_date, departure_date,
                arrival_date, priority, care_of, invoice_email, financial_email,
                purchase_order, external_notes, internal_notes, layout_urls,
                total_value)
         VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20,$21,$22)
         RETURNING id, seller_id, customer_id, payment_method, installments,
                   payment_term_days, first_installment_start, installment_dates,
                   status, is_event, delivery_address, delivery_date, departure_date,
                   arrival_date, priority, care_of, invoice_email, financial_email,
                   purchase_order, external_notes, internal_notes, layout_urls,
                   total_value, created_at, updated_at",
    )
    .bind(input.seller_id).bind(input.customer_id)
    .bind(&input.payment_method).bind(input.installments)
    .bind(input.payment_term_days).bind(input.first_installment_start)
    .bind(&input.installment_dates)
    .bind(input.status.as_deref().unwrap_or("Pendente"))
    .bind(input.is_event.unwrap_or(false))
    .bind(&input.delivery_address).bind(input.delivery_date)
    .bind(input.departure_date).bind(input.arrival_date)
    .bind(&input.priority).bind(&input.care_of)
    .bind(&input.invoice_email).bind(&input.financial_email)
    .bind(&input.purchase_order).bind(&input.external_notes)
    .bind(&input.internal_notes).bind(&input.layout_urls)
    .bind(total)
    .fetch_one(&mut *tx)
    .await
    .map_err(|e| AppError::internal(e.to_string()))?;

    for item in &input.items {
        let iprice = item.unit_price * item.quantity as f64 * (1.0 - item.discount.unwrap_or(0.0));
        sqlx::query(
            "INSERT INTO sale_items (sale_id, product_id, quantity, unit_price,
                    total_price, discount, engravings)
             VALUES ($1,$2,$3,$4,$5,$6,$7)",
        )
        .bind(sale.id).bind(item.product_id).bind(item.quantity)
        .bind(item.unit_price).bind(iprice).bind(item.discount.unwrap_or(0.0))
        .bind(&item.engravings)
        .execute(&mut *tx).await
        .map_err(|e| AppError::internal(e.to_string()))?;
    }

    for carrier_id in &input.carrier_ids {
        sqlx::query("INSERT INTO sale_carriers (sale_id, carrier_id) VALUES ($1,$2)")
            .bind(sale.id).bind(carrier_id)
            .execute(&mut *tx).await.map_err(|e| AppError::internal(e.to_string()))?;
    }

    tx.commit().await.map_err(|e| AppError::internal(e.to_string()))?;
    get_by_id(pool, sale.id).await
}

pub async fn update(pool: &sqlx::PgPool, id: i32, input: &SaleInput) -> Result<Sale, AppError> {
    let mut tx = pool.begin().await.map_err(|e| AppError::internal(e.to_string()))?;

    let total: f64 = input.items.iter().map(|i| i.unit_price * i.quantity as f64 * (1.0 - i.discount.unwrap_or(0.0))).sum();

    sqlx::query(
        "UPDATE sales SET seller_id=$1, customer_id=$2, payment_method=$3,
                installments=$4, payment_term_days=$5, first_installment_start=$6,
                installment_dates=$7, status=$8, is_event=$9,
                delivery_address=$10, delivery_date=$11, departure_date=$12,
                arrival_date=$13, priority=$14, care_of=$15,
                invoice_email=$16, financial_email=$17, purchase_order=$18,
                external_notes=$19, internal_notes=$20, layout_urls=$21,
                total_value=$22, updated_at=NOW()
         WHERE id=$23",
    )
    .bind(input.seller_id).bind(input.customer_id)
    .bind(&input.payment_method).bind(input.installments)
    .bind(input.payment_term_days).bind(input.first_installment_start)
    .bind(&input.installment_dates)
    .bind(input.status.as_deref().unwrap_or("Pendente"))
    .bind(input.is_event.unwrap_or(false))
    .bind(&input.delivery_address).bind(input.delivery_date)
    .bind(input.departure_date).bind(input.arrival_date)
    .bind(&input.priority).bind(&input.care_of)
    .bind(&input.invoice_email).bind(&input.financial_email)
    .bind(&input.purchase_order).bind(&input.external_notes)
    .bind(&input.internal_notes).bind(&input.layout_urls)
    .bind(total).bind(id)
    .execute(&mut *tx).await
    .map_err(|e| AppError::internal(e.to_string()))?;

    sqlx::query("DELETE FROM sale_items WHERE sale_id = $1").bind(id)
        .execute(&mut *tx).await.map_err(|e| AppError::internal(e.to_string()))?;
    sqlx::query("DELETE FROM sale_carriers WHERE sale_id = $1").bind(id)
        .execute(&mut *tx).await.map_err(|e| AppError::internal(e.to_string()))?;

    for item in &input.items {
        let iprice = item.unit_price * item.quantity as f64 * (1.0 - item.discount.unwrap_or(0.0));
        sqlx::query(
            "INSERT INTO sale_items (sale_id, product_id, quantity, unit_price,
                    total_price, discount, engravings)
             VALUES ($1,$2,$3,$4,$5,$6,$7)",
        )
        .bind(id).bind(item.product_id).bind(item.quantity)
        .bind(item.unit_price).bind(iprice).bind(item.discount.unwrap_or(0.0))
        .bind(&item.engravings)
        .execute(&mut *tx).await
        .map_err(|e| AppError::internal(e.to_string()))?;
    }

    for carrier_id in &input.carrier_ids {
        sqlx::query("INSERT INTO sale_carriers (sale_id, carrier_id) VALUES ($1,$2)")
            .bind(id).bind(carrier_id)
            .execute(&mut *tx).await.map_err(|e| AppError::internal(e.to_string()))?;
    }

    tx.commit().await.map_err(|e| AppError::internal(e.to_string()))?;
    get_by_id(pool, id).await
}

pub async fn update_layout(pool: &sqlx::PgPool, id: i32, urls: &[String]) -> Result<Sale, AppError> {
    sqlx::query("UPDATE sales SET layout_urls = $1, updated_at = NOW() WHERE id = $2")
        .bind(urls).bind(id)
        .execute(pool).await
        .map_err(|e| AppError::internal(e.to_string()))?;
    get_by_id(pool, id).await
}

pub async fn delete(pool: &sqlx::PgPool, id: i32) -> Result<(), AppError> {
    sqlx::query("DELETE FROM sales WHERE id = $1").bind(id).execute(pool).await
        .map_err(|e| AppError::internal(e.to_string()))?;
    Ok(())
}

const SALE_SELECT_ALL: &str = "
    SELECT s.id, s.seller_id, s.customer_id, s.payment_method, s.installments,
           s.payment_term_days, s.first_installment_start, s.installment_dates,
           s.status, s.is_event, s.delivery_address, s.delivery_date,
           s.departure_date, s.arrival_date, s.priority, s.care_of,
           s.invoice_email, s.financial_email, s.purchase_order,
           s.external_notes, s.internal_notes, s.layout_urls, s.total_value,
           s.created_at, s.updated_at
    FROM sales s";

async fn attach_items_bulk(pool: &sqlx::PgPool, mut sales: Vec<Sale>) -> Result<Vec<Sale>, AppError> {
    let ids: Vec<i32> = sales.iter().map(|s| s.id).collect();
    if ids.is_empty() { return Ok(sales); }
    let items = sqlx::query_as::<_, SaleItem>(
        "SELECT id, sale_id, product_id, quantity, unit_price, total_price,
                discount, price_formation, engravings, created_at, updated_at
         FROM sale_items WHERE sale_id = ANY($1) ORDER BY id")
    .bind(&ids).fetch_all(pool).await
    .map_err(|e| AppError::internal(e.to_string()))?;
    let mut map = std::collections::HashMap::new();
    for item in items { map.entry(item.sale_id).or_insert_with(Vec::new).push(item); }
    for s in &mut sales { s.items = map.remove(&s.id).unwrap_or_default(); }
    Ok(sales)
}

async fn get_carriers(pool: &sqlx::PgPool, sale_id: i32) -> Result<Vec<crate::models::Carrier>, AppError> {
    sqlx::query_as::<_, crate::models::Carrier>(
        "SELECT c.id, c.name, c.cnpj, c.carrier_type, c.email, c.landline_phone,
                c.mobile_phone, c.full_address, c.contact_name, c.contact_phone,
                c.website, c.created_at, c.updated_at
         FROM carriers c
         JOIN sale_carriers sc ON c.id = sc.carrier_id
         WHERE sc.sale_id = $1")
    .bind(sale_id).fetch_all(pool).await
    .map_err(|e| AppError::internal(e.to_string()))
}
