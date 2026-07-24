use crate::error::AppError;
use crate::models::{Quote, QuoteInput, QuoteItem};

pub async fn get_all(pool: &sqlx::PgPool) -> Result<Vec<Quote>, AppError> {
    let quotes = sqlx::query_as::<_, Quote>(QUOTE_SELECT_ALL)
        .fetch_all(pool)
        .await
        .map_err(|e| AppError::internal(e.to_string()))?;
    attach_items_bulk(pool, quotes).await
}

pub async fn get_by_id(pool: &sqlx::PgPool, id: i32) -> Result<Quote, AppError> {
    let mut quote = sqlx::query_as::<_, Quote>(&format!("{QUOTE_SELECT_ALL} WHERE q.id = $1"))
        .bind(id)
        .fetch_optional(pool)
        .await
        .map_err(|e| AppError::internal(e.to_string()))?
        .ok_or_else(|| AppError::not_found("Orçamento"))?;
    quote.items = sqlx::query_as::<_, QuoteItem>(
        "SELECT id, quote_id, product_id, quantity, unit_price, total_price,
                personalization_type, dn_code, description_summary, is_kit,
                base_cost_unit, labor_cost, extra_unit_cost1, extra_unit_cost2,
                engraving_cost, urgency_fee, logistics_cost, freight_cost,
                tax_percent, st_percent, loss_index_percent, imported_labor_percent,
                mgmt_commission_percent, seller_commission_percent,
                agency_commission_percent, publicity_percent, scrap_index,
                over_percent, financial_factor, financial_percent, sale_unit_value,
                transport_apart, production_cost_calc, transport_cost_calc,
                additional_costs_calc, profit_calc, margin_percent_calc,
                cost_unit_calc, engravings, has_price_formation, discount,
                created_at, updated_at
         FROM quote_items WHERE quote_id = $1 ORDER BY id",
    )
    .bind(id)
    .fetch_all(pool)
    .await
    .map_err(|e| AppError::internal(e.to_string()))?;
    Ok(quote)
}

pub async fn create(pool: &sqlx::PgPool, input: &QuoteInput) -> Result<Quote, AppError> {
    let mut tx = pool.begin().await.map_err(|e| AppError::internal(e.to_string()))?;

    let total: f64 = input.items.iter().map(|i| i.unit_price * i.quantity as f64).sum();

    let quote = sqlx::query_as::<_, Quote>(
        "INSERT INTO quotes (quote_number, seller_id, customer_id, responsible_name,
                quote_valid_until, production_lead_time, total_value, quote_date,
                care_of, sales_channel, observations, feedback_datetime,
                feedback_observation, payment_method, installments,
                installment_dates, carrier_id, freight_value,
                freight_tax_id_sender, freight_tax_id_origin, freight_tax_id_dest,
                freight_tax_id_payer, freight_tipo_transporte, freight_contato,
                freight_cidade_origem, freight_cidade_destino, freight_material,
                freight_tipo_frete, freight_produto, freight_tipo_embalagem,
                freight_quantidade, freight_volumes, freight_valor_nota,
                freight_peso_real)
         VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,
                 $19,$20,$21,$22,$23,$24,$25,$26,$27,$28,$29,$30,$31,$32,$33,$34)
         RETURNING id, quote_number, seller_id, customer_id, responsible_name,
                   quote_valid_until, production_lead_time, total_value,
                   quote_date, care_of, sales_channel, observations,
                   feedback_datetime, feedback_observation, payment_method,
                   installments, installment_dates, carrier_id, freight_value,
                   freight_tax_id_sender, freight_tax_id_origin, freight_tax_id_dest,
                   freight_tax_id_payer, freight_tipo_transporte, freight_contato,
                   freight_cidade_origem, freight_cidade_destino, freight_material,
                   freight_tipo_frete, freight_produto, freight_tipo_embalagem,
                   freight_quantidade, freight_volumes, freight_valor_nota,
                   freight_peso_real, created_at, updated_at",
    )
    .bind(&input.quote_number)
    .bind(input.seller_id)
    .bind(input.customer_id)
    .bind(&input.responsible_name)
    .bind(input.quote_valid_until)
    .bind(&input.production_lead_time)
    .bind(total)
    .bind(input.quote_date)
    .bind(&input.care_of)
    .bind(&input.sales_channel)
    .bind(&input.observations)
    .bind(input.feedback_datetime)
    .bind(&input.feedback_observation)
    .bind(&input.payment_method)
    .bind(input.installments)
    .bind(&input.installment_dates)
    .bind(input.carrier_id)
    .bind(input.freight_value)
    .bind(&input.freight_tax_id_sender)
    .bind(&input.freight_tax_id_origin)
    .bind(&input.freight_tax_id_dest)
    .bind(&input.freight_tax_id_payer)
    .bind(&input.freight_tipo_transporte)
    .bind(&input.freight_contato)
    .bind(&input.freight_cidade_origem)
    .bind(&input.freight_cidade_destino)
    .bind(&input.freight_material)
    .bind(&input.freight_tipo_frete)
    .bind(&input.freight_produto)
    .bind(&input.freight_tipo_embalagem)
    .bind(input.freight_quantidade)
    .bind(&input.freight_volumes)
    .bind(input.freight_valor_nota)
    .bind(input.freight_peso_real)
    .fetch_one(&mut *tx)
    .await
    .map_err(|e| AppError::internal(e.to_string()))?;

    for item in &input.items {
        let iprice = item.unit_price * item.quantity as f64;
        sqlx::query(
            "INSERT INTO quote_items (quote_id, product_id, quantity, unit_price,
                    total_price, personalization_type, dn_code, description_summary,
                    is_kit, base_cost_unit, labor_cost, extra_unit_cost1,
                    extra_unit_cost2, engraving_cost, urgency_fee, logistics_cost,
                    freight_cost, tax_percent, st_percent, loss_index_percent,
                    imported_labor_percent, mgmt_commission_percent,
                    seller_commission_percent, agency_commission_percent,
                    publicity_percent, scrap_index, over_percent, financial_factor,
                    financial_percent, sale_unit_value, transport_apart,
                    production_cost_calc, transport_cost_calc, additional_costs_calc,
                    profit_calc, margin_percent_calc, cost_unit_calc, engravings,
                    has_price_formation, discount)
             VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,
                     $17,$18,$19,$20,$21,$22,$23,$24,$25,$26,$27,$28,$29,$30,
                     $31,$32,$33,$34,$35,$36,$37,$38,$39,$40)",
        )
        .bind(quote.id)
        .bind(item.product_id)
        .bind(item.quantity)
        .bind(item.unit_price)
        .bind(iprice)
        .bind(&item.personalization_type)
        .bind(&item.dn_code)
        .bind(&item.description_summary)
        .bind(item.is_kit.unwrap_or(false))
        .bind(item.base_cost_unit.unwrap_or(0.0))
        .bind(item.labor_cost.unwrap_or(0.0))
        .bind(item.extra_unit_cost1.unwrap_or(0.0))
        .bind(item.extra_unit_cost2.unwrap_or(0.0))
        .bind(item.engraving_cost.unwrap_or(0.0))
        .bind(item.urgency_fee.unwrap_or(0.0))
        .bind(item.logistics_cost.unwrap_or(0.0))
        .bind(item.freight_cost.unwrap_or(0.0))
        .bind(item.tax_percent.unwrap_or(0.0))
        .bind(item.st_percent.unwrap_or(0.0))
        .bind(item.loss_index_percent.unwrap_or(0.0))
        .bind(item.imported_labor_percent.unwrap_or(0.0))
        .bind(item.mgmt_commission_percent.unwrap_or(0.0))
        .bind(item.seller_commission_percent.unwrap_or(0.0))
        .bind(item.agency_commission_percent.unwrap_or(0.0))
        .bind(item.publicity_percent.unwrap_or(0.0))
        .bind(item.scrap_index.unwrap_or(0.0))
        .bind(item.over_percent.unwrap_or(0.0))
        .bind(item.financial_factor.unwrap_or(0.0))
        .bind(item.financial_percent.unwrap_or(0.0))
        .bind(item.sale_unit_value.unwrap_or(0.0))
        .bind(item.transport_apart.unwrap_or(0.0))
        .bind(item.production_cost_calc.unwrap_or(0.0))
        .bind(item.transport_cost_calc.unwrap_or(0.0))
        .bind(item.additional_costs_calc.unwrap_or(0.0))
        .bind(item.profit_calc.unwrap_or(0.0))
        .bind(item.margin_percent_calc.unwrap_or(0.0))
        .bind(item.cost_unit_calc.unwrap_or(0.0))
        .bind(&item.engravings)
        .bind(item.has_price_formation.unwrap_or(false))
        .bind(item.discount.unwrap_or(0.0))
        .execute(&mut *tx)
        .await
        .map_err(|e| AppError::internal(e.to_string()))?;
    }

    tx.commit().await.map_err(|e| AppError::internal(e.to_string()))?;
    get_by_id(pool, quote.id).await
}

pub async fn update(pool: &sqlx::PgPool, id: i32, input: &QuoteInput) -> Result<Quote, AppError> {
    let mut tx = pool.begin().await.map_err(|e| AppError::internal(e.to_string()))?;

    let total: f64 = input.items.iter().map(|i| i.unit_price * i.quantity as f64).sum();

    sqlx::query(
        "UPDATE quotes SET seller_id=$1, customer_id=$2, responsible_name=$3,
                quote_valid_until=$4, production_lead_time=$5, total_value=$6,
                quote_date=$7, care_of=$8, sales_channel=$9, observations=$10,
                feedback_datetime=$11, feedback_observation=$12, payment_method=$13,
                installments=$14, installment_dates=$15, carrier_id=$16,
                freight_value=$17, freight_tax_id_sender=$18,
                freight_tax_id_origin=$19, freight_tax_id_dest=$20,
                freight_tax_id_payer=$21, freight_tipo_transporte=$22,
                freight_contato=$23, freight_cidade_origem=$24,
                freight_cidade_destino=$25, freight_material=$26,
                freight_tipo_frete=$27, freight_produto=$28,
                freight_tipo_embalagem=$29, freight_quantidade=$30,
                freight_volumes=$31, freight_valor_nota=$32,
                freight_peso_real=$33, updated_at=NOW()
         WHERE id=$34",
    )
    .bind(input.seller_id)
    .bind(input.customer_id)
    .bind(&input.responsible_name)
    .bind(input.quote_valid_until)
    .bind(&input.production_lead_time)
    .bind(total)
    .bind(input.quote_date)
    .bind(&input.care_of)
    .bind(&input.sales_channel)
    .bind(&input.observations)
    .bind(input.feedback_datetime)
    .bind(&input.feedback_observation)
    .bind(&input.payment_method)
    .bind(input.installments)
    .bind(&input.installment_dates)
    .bind(input.carrier_id)
    .bind(input.freight_value)
    .bind(&input.freight_tax_id_sender)
    .bind(&input.freight_tax_id_origin)
    .bind(&input.freight_tax_id_dest)
    .bind(&input.freight_tax_id_payer)
    .bind(&input.freight_tipo_transporte)
    .bind(&input.freight_contato)
    .bind(&input.freight_cidade_origem)
    .bind(&input.freight_cidade_destino)
    .bind(&input.freight_material)
    .bind(&input.freight_tipo_frete)
    .bind(&input.freight_produto)
    .bind(&input.freight_tipo_embalagem)
    .bind(input.freight_quantidade)
    .bind(&input.freight_volumes)
    .bind(input.freight_valor_nota)
    .bind(input.freight_peso_real)
    .bind(id)
    .execute(&mut *tx)
    .await
    .map_err(|e| AppError::internal(e.to_string()))?;

    sqlx::query("DELETE FROM quote_items WHERE quote_id = $1").bind(id)
        .execute(&mut *tx).await.map_err(|e| AppError::internal(e.to_string()))?;

    for item in &input.items {
        let iprice = item.unit_price * item.quantity as f64;
        sqlx::query(
            "INSERT INTO quote_items (quote_id, product_id, quantity, unit_price,
                    total_price, personalization_type)
             VALUES ($1,$2,$3,$4,$5,$6)",
        )
        .bind(id).bind(item.product_id).bind(item.quantity)
        .bind(item.unit_price).bind(iprice).bind(&item.personalization_type)
        .execute(&mut *tx).await
        .map_err(|e| AppError::internal(e.to_string()))?;
    }

    tx.commit().await.map_err(|e| AppError::internal(e.to_string()))?;
    get_by_id(pool, id).await
}

pub async fn delete(pool: &sqlx::PgPool, id: i32) -> Result<(), AppError> {
    sqlx::query("DELETE FROM quotes WHERE id = $1").bind(id).execute(pool).await
        .map_err(|e| AppError::internal(e.to_string()))?;
    Ok(())
}

const QUOTE_SELECT_ALL: &str = "
    SELECT q.id, q.quote_number, q.seller_id, q.customer_id, q.responsible_name,
           q.quote_valid_until, q.production_lead_time, q.total_value,
           q.quote_date, q.care_of, q.sales_channel, q.observations,
           q.feedback_datetime, q.feedback_observation, q.payment_method,
           q.installments, q.installment_dates, q.carrier_id, q.freight_value,
           q.freight_tax_id_sender, q.freight_tax_id_origin, q.freight_tax_id_dest,
           q.freight_tax_id_payer, q.freight_tipo_transporte, q.freight_contato,
           q.freight_cidade_origem, q.freight_cidade_destino, q.freight_material,
           q.freight_tipo_frete, q.freight_produto, q.freight_tipo_embalagem,
           q.freight_quantidade, q.freight_volumes, q.freight_valor_nota,
           q.freight_peso_real, q.created_at, q.updated_at
    FROM quotes q";

async fn attach_items_bulk(pool: &sqlx::PgPool, mut quotes: Vec<Quote>) -> Result<Vec<Quote>, AppError> {
    let ids: Vec<i32> = quotes.iter().map(|q| q.id).collect();
    if ids.is_empty() {
        return Ok(quotes);
    }
    let items = sqlx::query_as::<_, QuoteItem>(
        "SELECT id, quote_id, product_id, quantity, unit_price, total_price,
                personalization_type, dn_code, description_summary, is_kit,
                base_cost_unit, labor_cost, extra_unit_cost1, extra_unit_cost2,
                engraving_cost, urgency_fee, logistics_cost, freight_cost,
                tax_percent, st_percent, loss_index_percent, imported_labor_percent,
                mgmt_commission_percent, seller_commission_percent,
                agency_commission_percent, publicity_percent, scrap_index,
                over_percent, financial_factor, financial_percent, sale_unit_value,
                transport_apart, production_cost_calc, transport_cost_calc,
                additional_costs_calc, profit_calc, margin_percent_calc,
                cost_unit_calc, engravings, has_price_formation, discount,
                created_at, updated_at
         FROM quote_items WHERE quote_id = ANY($1) ORDER BY id",
    )
    .bind(&ids)
    .fetch_all(pool)
    .await
    .map_err(|e| AppError::internal(e.to_string()))?;

    let mut map = std::collections::HashMap::new();
    for item in items {
        map.entry(item.quote_id).or_insert_with(Vec::new).push(item);
    }
    for q in &mut quotes {
        q.items = map.remove(&q.id).unwrap_or_default();
    }
    Ok(quotes)
}
