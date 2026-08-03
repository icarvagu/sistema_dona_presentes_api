use crate::error::AppError;
use crate::models::{FinancialReportItem, PaginatedProductResponse, Product, ProductInput, ProductItem};
use sqlx::{Postgres, QueryBuilder};

pub async fn get_all(pool: &sqlx::PgPool) -> Result<Vec<Product>, AppError> {
    let products = sqlx::query_as::<_, Product>(PRODUCT_SELECT)
        .fetch_all(pool)
        .await
        .map_err(|e| AppError::internal(e.to_string()))?;
    let items = get_items_bulk(pool).await?;
    Ok(attach_items(products, items))
}

pub async fn get_by_id(pool: &sqlx::PgPool, id: i32) -> Result<Product, AppError> {
    let mut product = sqlx::query_as::<_, Product>(&format!("{PRODUCT_SELECT} WHERE p.id = $1"))
        .bind(id)
        .fetch_optional(pool)
        .await
        .map_err(|e| AppError::internal(e.to_string()))?
        .ok_or_else(|| AppError::not_found("Produto"))?;
    product.items = get_items(pool, id).await?;
    Ok(product)
}

pub async fn get_paginated(
    pool: &sqlx::PgPool,
    filter: &str,
    group: Option<&str>,
    only_new: bool,
    page: i32,
    limit: i32,
) -> Result<PaginatedProductResponse, AppError> {
    let offset = (page - 1) * limit;

    let mut count_query = QueryBuilder::<Postgres>::new("SELECT COUNT(*) FROM products p");
    push_product_filters(&mut count_query, filter, group, only_new);
    let total: i64 = count_query
        .build_query_scalar()
        .fetch_one(pool)
        .await
        .map_err(|e| AppError::internal(e.to_string()))?;

    let mut data_query = QueryBuilder::<Postgres>::new(PRODUCT_SELECT);
    push_product_filters(&mut data_query, filter, group, only_new);
    data_query
        .push(" ORDER BY p.id LIMIT ")
        .push_bind(limit as i64)
        .push(" OFFSET ")
        .push_bind(offset as i64);

    let data = data_query
        .build_query_as::<Product>()
        .fetch_all(pool)
        .await
        .map_err(|e| AppError::internal(e.to_string()))?;

    let items = get_items_bulk(pool).await?;
    let data = attach_items(data, items);

    Ok(PaginatedProductResponse {
        data,
        total,
        page,
        limit,
        total_pages: ((total + limit as i64 - 1) / limit as i64) as i32,
    })
}

fn push_product_filters(
    query: &mut QueryBuilder<'_, Postgres>,
    filter: &str,
    group: Option<&str>,
    only_new: bool,
) {
    let mut has_filter = false;
    let mut push_conjunction = |query: &mut QueryBuilder<'_, Postgres>| {
        if has_filter {
            query.push(" AND ");
        } else {
            query.push(" WHERE ");
            has_filter = true;
        }
    };

    if !filter.trim().is_empty() {
        let filter_like = format!("%{}%", filter.trim());
        push_conjunction(query);
        query
            .push("(")
            .push("p.product_name ILIKE ")
            .push_bind(filter_like.clone())
            .push(" OR p.internal_code ILIKE ")
            .push_bind(filter_like.clone())
            .push(" OR p.supplier_code ILIKE ")
            .push_bind(filter_like)
            .push(")");
    }

    if let Some(group) = group.filter(|value| !value.trim().is_empty()) {
        push_conjunction(query);
        query.push("p.product_group = ").push_bind(group.trim().to_string());
    }

    if only_new {
        push_conjunction(query);
        query.push("p.created_at >= now() - interval '30 days'");
    }
}

pub async fn create(pool: &sqlx::PgPool, input: &ProductInput) -> Result<Product, AppError> {
    let mut tx = pool.begin().await.map_err(|e| AppError::internal(e.to_string()))?;

    let product = sqlx::query_as::<_, Product>(&format!(
        "INSERT INTO products (product_name, internal_code, supplier_code, supplier_id,
                product_group, description, photos, ncm, material_origin, stock,
                supplier_stock, moves_stock, enabled_for_invoice, cost_price,
                selling_price, kit_type, is_composition, color)
         VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18)
         RETURNING {RETURN_COLS}"
    ))
    .bind(&input.product_name)
    .bind(&input.internal_code)
    .bind(&input.supplier_code)
    .bind(input.supplier_id)
    .bind(&input.product_group)
    .bind(&input.description)
    .bind(&input.photos)
    .bind(&input.ncm)
    .bind(&input.material_origin)
    .bind(input.stock.unwrap_or(0))
    .bind(input.supplier_stock.unwrap_or(0))
    .bind(input.moves_stock.unwrap_or(false))
    .bind(input.enabled_for_invoice.unwrap_or(true))
    .bind(input.cost_price.unwrap_or(0.0))
    .bind(input.selling_price.unwrap_or(0.0))
    .bind(input.kit_type.as_deref().unwrap_or("none"))
    .bind(input.is_composition.unwrap_or(false))
    .bind(&input.color)
    .fetch_one(&mut *tx)
    .await
    .map_err(|e| AppError::internal(e.to_string()))?;

    for item in &input.items {
        sqlx::query(
            "INSERT INTO product_items (product_parent_id, product_id, quantity)
             VALUES ($1, $2, $3)",
        )
        .bind(product.id)
        .bind(item.product_id)
        .bind(item.quantity)
        .execute(&mut *tx)
        .await
        .map_err(|e| AppError::internal(e.to_string()))?;
    }

    tx.commit().await.map_err(|e| AppError::internal(e.to_string()))?;

    let mut p = product;
    p.items = get_items(pool, p.id).await?;
    Ok(p)
}

pub async fn update(pool: &sqlx::PgPool, id: i32, input: &ProductInput) -> Result<Product, AppError> {
    sqlx::query(
        "UPDATE products SET product_name=$1, internal_code=$2, supplier_code=$3, supplier_id=$4,
                product_group=$5, description=$6, photos=$7, ncm=$8, material_origin=$9,
                stock=$10, supplier_stock=$11, moves_stock=$12, enabled_for_invoice=$13,
                cost_price=$14, selling_price=$15, kit_type=$16, is_composition=$17,
                color=$18, updated_at=NOW()
         WHERE id=$19",
    )
    .bind(&input.product_name)
    .bind(&input.internal_code)
    .bind(&input.supplier_code)
    .bind(input.supplier_id)
    .bind(&input.product_group)
    .bind(&input.description)
    .bind(&input.photos)
    .bind(&input.ncm)
    .bind(&input.material_origin)
    .bind(input.stock.unwrap_or(0))
    .bind(input.supplier_stock.unwrap_or(0))
    .bind(input.moves_stock.unwrap_or(false))
    .bind(input.enabled_for_invoice.unwrap_or(true))
    .bind(input.cost_price.unwrap_or(0.0))
    .bind(input.selling_price.unwrap_or(0.0))
    .bind(input.kit_type.as_deref().unwrap_or("none"))
    .bind(input.is_composition.unwrap_or(false))
    .bind(&input.color)
    .bind(id)
    .execute(pool)
    .await
    .map_err(|e| AppError::internal(e.to_string()))?;

    get_by_id(pool, id).await
}

pub async fn delete(pool: &sqlx::PgPool, id: i32) -> Result<(), AppError> {
    sqlx::query("DELETE FROM products WHERE id = $1")
        .bind(id)
        .execute(pool)
        .await
        .map_err(|e| AppError::internal(e.to_string()))?;
    Ok(())
}

pub async fn get_groups(pool: &sqlx::PgPool) -> Result<Vec<String>, AppError> {
    sqlx::query_scalar("SELECT DISTINCT product_group FROM products WHERE product_group IS NOT NULL ORDER BY product_group")
        .fetch_all(pool)
        .await
        .map_err(|e| AppError::internal(e.to_string()))
}

pub async fn get_pending_approval(pool: &sqlx::PgPool) -> Result<Vec<Product>, AppError> {
    sqlx::query_as::<_, Product>(&format!(
        "{PRODUCT_SELECT} WHERE p.pending_approval = true ORDER BY p.id"
    ))
    .fetch_all(pool)
    .await
    .map_err(|e| AppError::internal(e.to_string()))
}

pub async fn approve(pool: &sqlx::PgPool, id: i32) -> Result<(), AppError> {
    sqlx::query("UPDATE products SET pending_approval = false WHERE id = $1")
        .bind(id)
        .execute(pool)
        .await
        .map_err(|e| AppError::internal(e.to_string()))?;
    Ok(())
}

pub async fn get_financial_report(pool: &sqlx::PgPool) -> Result<Vec<FinancialReportItem>, AppError> {
    sqlx::query_as::<_, FinancialReportItem>(
        "SELECT id as product_id, product_name, internal_code, kit_type, cost_price::float8,
                selling_price::float8, (selling_price::float8 - cost_price::float8) as margin_value,
                CASE WHEN cost_price::float8 > 0 THEN ((selling_price::float8 - cost_price::float8) / cost_price::float8 * 100.0) ELSE 0 END as margin_percent,
                moves_stock, enabled_for_invoice
         FROM products ORDER BY product_name",
    )
    .fetch_all(pool)
    .await
    .map_err(|e| AppError::internal(e.to_string()))
}

pub async fn update_last_cost(
    pool: &sqlx::PgPool,
    id: i32,
    cost: f64,
    date: chrono::NaiveDateTime,
    qty1: i32,
    qty2: i32,
    qty3: i32,
    val1: f64,
    val2: f64,
    val3: f64,
    user: &str,
) -> Result<(), AppError> {
    sqlx::query(
        "UPDATE products SET last_cost=$1, last_cost_date=$2, last_cost_qty1=$3,
                last_cost_qty2=$4, last_cost_qty3=$5, last_cost_val1=$6,
                last_cost_val2=$7, last_cost_val3=$8, last_cost_user=$9
         WHERE id=$10",
    )
    .bind(cost)
    .bind(date)
    .bind(qty1)
    .bind(qty2)
    .bind(qty3)
    .bind(val1)
    .bind(val2)
    .bind(val3)
    .bind(user)
    .bind(id)
    .execute(pool)
    .await
    .map_err(|e| AppError::internal(e.to_string()))?;
    Ok(())
}

const PRODUCT_SELECT: &str = "
    SELECT p.id, p.product_name, p.internal_code, p.supplier_code, p.supplier_id,
           p.product_group, p.description, p.photos, p.ncm, p.material_origin,
           p.stock, p.supplier_stock, p.moves_stock, p.enabled_for_invoice,
           p.cost_price::float8, p.selling_price::float8, p.kit_type, p.is_composition,
           p.source, p.imported_at, p.last_synced_at, p.color, p.origin,
           p.pending_approval, p.last_cost::float8, p.last_cost_date,
           p.last_cost_qty1, p.last_cost_qty2, p.last_cost_qty3,
           p.last_cost_val1::float8, p.last_cost_val2::float8, p.last_cost_val3::float8, p.last_cost_user,
           p.created_at, p.updated_at
    FROM products p";

const RETURN_COLS: &str = "
    p.id, p.product_name, p.internal_code, p.supplier_code, p.supplier_id,
    p.product_group, p.description, p.photos, p.ncm, p.material_origin,
    p.stock, p.supplier_stock, p.moves_stock, p.enabled_for_invoice,
    p.cost_price::float8, p.selling_price::float8, p.kit_type, p.is_composition,
    p.source, p.imported_at, p.last_synced_at, p.color, p.origin,
    p.pending_approval, p.last_cost::float8, p.last_cost_date,
    p.last_cost_qty1, p.last_cost_qty2, p.last_cost_qty3,
    p.last_cost_val1::float8, p.last_cost_val2::float8, p.last_cost_val3::float8, p.last_cost_user,
    p.created_at, p.updated_at";

async fn get_items_bulk(pool: &sqlx::PgPool) -> Result<std::collections::HashMap<i32, Vec<ProductItem>>, AppError> {
    let rows = sqlx::query_as::<_, ProductItem>(
        "SELECT id, product_parent_id, product_id, quantity, created_at, updated_at FROM product_items",
    )
    .fetch_all(pool)
    .await
    .map_err(|e| AppError::internal(e.to_string()))?;
    let mut map = std::collections::HashMap::new();
    for r in rows {
        map.entry(r.product_parent_id).or_insert_with(Vec::new).push(r);
    }
    Ok(map)
}

pub async fn search(pool: &sqlx::PgPool, q: &str, limit: i32) -> Result<Vec<Product>, AppError> {
    let products = if q.is_empty() {
        sqlx::query_as::<_, Product>(&format!("{PRODUCT_SELECT} ORDER BY id LIMIT $1"))
            .bind(limit as i64)
            .fetch_all(pool).await
    } else {
        sqlx::query_as::<_, Product>(&format!(
            "{PRODUCT_SELECT} WHERE p.product_name ILIKE $1 OR p.internal_code ILIKE $1 ORDER BY id LIMIT $2"
        ))
        .bind(format!("%{q}%")).bind(limit as i64)
        .fetch_all(pool).await
    }
    .map_err(|e| AppError::internal(e.to_string()))?;

    let items = get_items_bulk(pool).await?;
    Ok(attach_items(products, items))
}

pub async fn get_items(pool: &sqlx::PgPool, parent_id: i32) -> Result<Vec<ProductItem>, AppError> {
    sqlx::query_as::<_, ProductItem>(
        "SELECT id, product_parent_id, product_id, quantity, created_at, updated_at
         FROM product_items WHERE product_parent_id = $1 ORDER BY id"
    )
    .bind(parent_id).fetch_all(pool).await
    .map_err(|e| AppError::internal(e.to_string()))
}

pub async fn create_item(
    pool: &sqlx::PgPool, parent_id: i32, input: &crate::models::ProductItemInput,
) -> Result<ProductItem, AppError> {
    sqlx::query_as::<_, ProductItem>(
        "INSERT INTO product_items (product_parent_id, product_id, quantity)
         VALUES ($1,$2,$3)
         RETURNING id, product_parent_id, product_id, quantity, created_at, updated_at"
    )
    .bind(parent_id).bind(input.product_id).bind(input.quantity)
    .fetch_one(pool).await
    .map_err(|e| AppError::internal(e.to_string()))
}

pub async fn delete_item(pool: &sqlx::PgPool, item_id: i32) -> Result<(), AppError> {
    sqlx::query("DELETE FROM product_items WHERE id = $1").bind(item_id)
        .execute(pool).await
        .map_err(|e| AppError::internal(e.to_string()))?;
    Ok(())
}

pub async fn get_by_group(pool: &sqlx::PgPool, group: &str) -> Result<Vec<Product>, AppError> {
    let products = sqlx::query_as::<_, Product>(
        &format!("{PRODUCT_SELECT} WHERE p.product_group = $1 ORDER BY p.product_name")
    )
    .bind(group)
    .fetch_all(pool).await
    .map_err(|e| AppError::internal(e.to_string()))?;
    let items = get_items_bulk(pool).await?;
    Ok(attach_items(products, items))
}

pub async fn bulk_approve(pool: &sqlx::PgPool, ids: &[i32]) -> Result<u64, AppError> {
    let result = sqlx::query(
        "UPDATE products SET pending_approval = false, updated_at = NOW() WHERE id = ANY($1)"
    )
    .bind(ids)
    .execute(pool).await
    .map_err(|e| AppError::internal(e.to_string()))?;
    Ok(result.rows_affected())
}

fn attach_items(mut products: Vec<Product>, items: std::collections::HashMap<i32, Vec<ProductItem>>) -> Vec<Product> {
    for p in &mut products {
        p.items = items.get(&p.id).cloned().unwrap_or_default();
    }
    products
}
