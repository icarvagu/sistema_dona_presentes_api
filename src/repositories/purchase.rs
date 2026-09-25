use serde::Serialize;
use serde_json::Value;
use sqlx::types::Json;

use crate::error::AppError;
use crate::models::{
    Notification, PurchaseActionInput, PurchaseAttachment, PurchaseEmail, PurchaseHistory,
    PurchaseIssue, PurchaseIssueInput, PurchaseIssueUpdateInput, PurchaseOrder,
    PurchasePayment, PurchasePaymentInput, PurchaseReleaseInput, PurchaseUpdateInput,
    PurchaseRequest, PurchaseRequestInput,
    PurchaseRequestUpdateInput, PurchaseBatchInput, PurchaseBatchResult,
};

pub async fn create_purchase_batches(
    pool: &sqlx::PgPool,
    input: &PurchaseBatchInput,
    user_id: i32,
) -> Result<Vec<PurchaseBatchResult>, AppError> {
    let ids: Vec<i32> = input.purchase_ids.iter().copied().collect();
    if ids.is_empty() {
        return Err(AppError::bad_request("Selecione ao menos um pedido para comprar"));
    }
    let mut tx = pool.begin().await.map_err(|e| AppError::internal(e.to_string()))?;
    let rows = sqlx::query_as::<_, (i32, Option<i32>, Value)>(
        "SELECT po.id, po.material_supplier_id,
                COALESCE(sale_items.rows, '[]'::jsonb)
         FROM purchase_orders po
         LEFT JOIN LATERAL (
           SELECT jsonb_agg(jsonb_build_object(
             'product_id', si.product_id,
             'product_name', COALESCE(p.product_name, 'Item sem cadastro'),
             'quantity', si.quantity,
             'sale_id', si.sale_id
           ) ORDER BY si.id) AS rows
           FROM sale_items si LEFT JOIN products p ON p.id = si.product_id
           WHERE si.sale_id = po.sale_id
         ) sale_items ON TRUE
         WHERE po.id = ANY($1) AND po.status NOT IN ('Pedido Comprado', 'Liberado para Produção')
         ORDER BY po.id"
    ).bind(&ids).fetch_all(&mut *tx).await.map_err(|e| AppError::internal(e.to_string()))?;
    if rows.is_empty() {
        return Err(AppError::bad_request("Nenhum dos pedidos selecionados está disponível para compra"));
    }

    let mut by_supplier: std::collections::BTreeMap<Option<i32>, Vec<(i32, Value)>> = std::collections::BTreeMap::new();
    for (purchase_id, supplier_id, items) in rows {
        by_supplier.entry(supplier_id).or_default().push((purchase_id, items));
    }
    let mut results = Vec::new();
    for (supplier_id, orders) in by_supplier {
        let batch_id: i32 = sqlx::query_scalar(
            "INSERT INTO purchase_batches (supplier_id, created_by) VALUES ($1, $2) RETURNING id"
        ).bind(supplier_id).bind(user_id).fetch_one(&mut *tx).await.map_err(|e| AppError::internal(e.to_string()))?;
        let mut aggregate: std::collections::BTreeMap<(Option<i32>, String), (f64, Vec<Value>)> = std::collections::BTreeMap::new();
        let mut purchase_ids = Vec::new();
        for (purchase_id, items) in orders {
            purchase_ids.push(purchase_id);
            sqlx::query("INSERT INTO purchase_batch_orders (batch_id, purchase_id) VALUES ($1, $2)")
                .bind(batch_id).bind(purchase_id).execute(&mut *tx).await.map_err(|e| AppError::internal(e.to_string()))?;
            if let Some(items) = items.as_array() {
                for item in items {
                    let product_id = item.get("product_id").and_then(Value::as_i64).map(|v| v as i32);
                    let name = item.get("product_name").and_then(Value::as_str).unwrap_or("Item sem cadastro").to_string();
                    let quantity = item.get("quantity").and_then(Value::as_f64).unwrap_or(0.0);
                    let allocation = serde_json::json!({ "purchase_id": purchase_id, "sale_id": item.get("sale_id"), "quantity": quantity });
                    let entry = aggregate.entry((product_id, name)).or_insert((0.0, Vec::new()));
                    entry.0 += quantity;
                    entry.1.push(allocation);
                }
            }
            sqlx::query("UPDATE purchase_orders SET status = 'E-mail Material Enviado', status_updated_at = NOW(), updated_at = NOW() WHERE id = $1")
                .bind(purchase_id).execute(&mut *tx).await.map_err(|e| AppError::internal(e.to_string()))?;
            sqlx::query("INSERT INTO purchase_history (purchase_id, action, from_status, to_status, user_id, user_name, details) SELECT $1, 'compra agrupada', '', 'E-mail Material Enviado', $2, COALESCE(full_name, 'Compras'), $3 FROM users WHERE id = $2")
                .bind(purchase_id).bind(user_id).bind(format!("Lote de compra #{}; CNPJ permanece separado na entrega.", batch_id))
                .execute(&mut *tx).await.map_err(|e| AppError::internal(e.to_string()))?;
        }
        let mut output_items = Vec::new();
        for ((product_id, product_name), (total_quantity, allocations)) in aggregate {
            sqlx::query("INSERT INTO purchase_batch_items (batch_id, product_id, product_name, total_quantity, allocations) VALUES ($1, $2, $3, $4, $5)")
                .bind(batch_id).bind(product_id).bind(&product_name).bind(total_quantity).bind(sqlx::types::Json(allocations.clone()))
                .execute(&mut *tx).await.map_err(|e| AppError::internal(e.to_string()))?;
            output_items.push(serde_json::json!({ "product_id": product_id, "product_name": product_name, "total_quantity": total_quantity, "allocations": allocations }));
        }
        results.push(PurchaseBatchResult { batch_id, supplier_id, purchase_ids, items: Value::Array(output_items) });
    }
    tx.commit().await.map_err(|e| AppError::internal(e.to_string()))?;
    Ok(results)
}

const PURCHASE_REQUEST_SELECT: &str =
    "SELECT r.id, r.request_type, r.product, r.product_link, r.supplier, r.quantity,
            r.description, r.attachments, r.status, r.requested_by,
            u.full_name AS requester_name, r.created_at, r.updated_at,
            r.items, r.total_value, r.freight_value, r.payment_method,
            r.delivery_date, r.receiver_name, r.receiver_phone, r.buyer_message
     FROM purchase_requests r LEFT JOIN users u ON u.id = r.requested_by";

pub async fn get_requests(pool: &sqlx::PgPool) -> Result<Vec<PurchaseRequest>, AppError> {
    sqlx::query_as::<_, PurchaseRequest>(&format!("{PURCHASE_REQUEST_SELECT} ORDER BY r.id DESC"))
        .fetch_all(pool).await.map_err(|e| AppError::internal(e.to_string()))
}

pub async fn create_request(pool: &sqlx::PgPool, input: &PurchaseRequestInput, user_id: i32) -> Result<PurchaseRequest, AppError> {
    if !["supply", "internal"].contains(&input.request_type.as_str()) {
        return Err(AppError::bad_request("Tipo de solicitação inválido"));
    }
    if input.product.trim().is_empty() {
        return Err(AppError::bad_request("Informe o produto que deseja"));
    }
    if input.request_type == "internal" && input.product_link.as_deref().unwrap_or("").trim().is_empty() {
        return Err(AppError::bad_request("O link do produto é obrigatório para compra interna"));
    }
    if input.request_type == "internal" && input.supplier.as_deref().unwrap_or("").trim().is_empty() {
        return Err(AppError::bad_request("Informe o fornecedor do produto"));
    }
    if input.request_type == "internal" && input.quantity.unwrap_or(0.0) <= 0.0 {
        return Err(AppError::bad_request("Informe uma quantidade válida"));
    }
    if input.request_type == "internal" {
        let used: Option<i32> = sqlx::query_scalar(
            "SELECT id FROM purchase_requests
             WHERE request_type = 'internal' AND requested_by = $1
               AND created_at >= date_trunc('month', NOW())
               AND created_at < date_trunc('month', NOW()) + INTERVAL '1 month'
             ORDER BY id DESC LIMIT 1"
        ).bind(user_id).fetch_optional(pool).await
            .map_err(|e| AppError::internal(e.to_string()))?;
        if used.is_some() {
            return Err(AppError::bad_request("A compra interna deste mês já foi solicitada. Ela ficará disponível novamente no próximo mês."));
        }
    }
    sqlx::query_as::<_, PurchaseRequest>(
        "INSERT INTO purchase_requests
          (request_type, product, product_link, supplier, quantity, description, attachments, requested_by)
         VALUES ($1,$2,$3,$4,$5,$6,$7,$8)
         RETURNING id, request_type, product, product_link, supplier, quantity, description,
                   attachments, status, requested_by, NULL::TEXT AS requester_name, created_at, updated_at,
                   items, total_value, freight_value, payment_method, delivery_date, receiver_name, receiver_phone, buyer_message"
    ).bind(&input.request_type).bind(input.product.trim())
     .bind(input.product_link.as_deref().unwrap_or(""))
     .bind(input.supplier.as_deref().unwrap_or(""))
     .bind(input.quantity).bind(input.description.as_deref().unwrap_or(""))
     .bind(sqlx::types::Json(input.attachments.clone().unwrap_or_default())).bind(user_id)
     .fetch_one(pool).await.map_err(|e| AppError::internal(e.to_string()))
}

pub async fn update_request(pool: &sqlx::PgPool, id: i32, input: &PurchaseRequestUpdateInput) -> Result<PurchaseRequest, AppError> {
    let date = input.delivery_date.as_deref().filter(|value| !value.is_empty());
    let request = sqlx::query_as::<_, PurchaseRequest>(&format!("{PURCHASE_REQUEST_SELECT} WHERE r.id = $1"))
    .bind(id)
    .fetch_optional(pool).await.map_err(|e| AppError::internal(e.to_string()))?
    .ok_or_else(|| AppError::not_found("Solicitação de compra"))?;
    sqlx::query("UPDATE purchase_requests SET product=$1, supplier=$2, quantity=$3, description=$4, items=$5, total_value=$6, freight_value=$7, payment_method=$8, delivery_date=$9::date, receiver_name=$10, receiver_phone=$11, status=COALESCE($12,status), buyer_message=$13, updated_at=NOW() WHERE id=$14")
        .bind(input.product.as_deref().unwrap_or(&request.product))
        .bind(input.supplier.as_deref().unwrap_or(&request.supplier))
        .bind(input.quantity.or(request.quantity))
        .bind(input.description.as_deref().unwrap_or(&request.description))
        .bind(sqlx::types::Json(input.items.clone().unwrap_or_else(|| request.items.clone().unwrap_or_else(|| serde_json::json!([])))))
        .bind(input.total_value.or(request.total_value))
        .bind(input.freight_value.or(request.freight_value))
        .bind(input.payment_method.as_deref().unwrap_or(&request.payment_method))
        .bind(date)
        .bind(input.receiver_name.as_deref().unwrap_or(&request.receiver_name))
        .bind(input.receiver_phone.as_deref().unwrap_or(&request.receiver_phone))
        .bind(input.status.as_deref())
        .bind(input.buyer_message.as_deref().unwrap_or(&request.buyer_message))
        .bind(id).execute(pool).await.map_err(|e| AppError::internal(e.to_string()))?;
    if input.status.as_deref() == Some("Concluída") {
        if let Some(recipient) = request.requested_by {
            let total = input.total_value.or(request.total_value).unwrap_or(0.0);
            sqlx::query(
                "INSERT INTO notifications (purchase_id, recipient_user_id, recipient_permission, notification_type, message)
                 SELECT NULL, $1, '', 'internal_purchase_completed', $2
                 WHERE NOT EXISTS (
                   SELECT 1 FROM notifications
                   WHERE recipient_user_id = $1 AND notification_type = 'internal_purchase_completed'
                     AND message LIKE $3
                 )"
            ).bind(recipient)
             .bind(format!("Compra interna CIN-{:05} concluída: R$ {:.2} para desconto em folha.", id, total).replace('.', ","))
             .bind(format!("Compra interna CIN-{:05} concluída:%", id))
             .execute(pool).await.map_err(|e| AppError::internal(e.to_string()))?;
        }
    }
    sqlx::query_as::<_, PurchaseRequest>(&format!("{PURCHASE_REQUEST_SELECT} WHERE r.id = $1"))
        .bind(id).fetch_one(pool).await.map_err(|e| AppError::internal(e.to_string()))
}

pub async fn get_all(pool: &sqlx::PgPool) -> Result<Vec<PurchaseOrder>, AppError> {
    sqlx::query_as::<_, PurchaseOrder>(SELECT_PURCHASE_ALL)
        .fetch_all(pool)
        .await
        .map_err(|e| AppError::internal(e.to_string()))
}

pub async fn get_by_id(pool: &sqlx::PgPool, id: i32) -> Result<PurchaseOrder, AppError> {
    let mut po = sqlx::query_as::<_, PurchaseOrder>(&format!("{SELECT_PURCHASE_ALL} WHERE po.id = $1"))
        .bind(id)
        .fetch_optional(pool)
        .await
        .map_err(|e| AppError::internal(e.to_string()))?
        .ok_or_else(|| AppError::not_found("Ordem de compra"))?;
    po.attachments = get_attachments(pool, id).await?;
    po.emails = get_emails(pool, id).await?;
    po.payments = get_payments(pool, id).await?;
    po.issues = get_issues(pool, id).await?;
    po.history = get_history(pool, id).await?;
    Ok(po)
}

#[derive(Debug, Serialize, sqlx::FromRow)]
pub struct PurchaseFinancialRow {
    pub purchase_id: i32,
    pub general_number: String,
    pub customer_name: String,
    pub cost_type: String,
    pub supplier_name: Option<String>,
    pub amount: f64,
    pub method: String,
    pub status: String,
    pub receipt_url: Option<String>,
}

pub async fn get_financial_summary(pool: &sqlx::PgPool) -> Result<Vec<PurchaseFinancialRow>, AppError> {
    sqlx::query_as::<_, PurchaseFinancialRow>(
        "SELECT po.id as purchase_id, po.general_number,
                COALESCE(c.name, '') as customer_name,
                pp.cost_type,
                COALESCE(s.name, '') as supplier_name,
                pp.amount, pp.method, pp.status, pp.receipt_url
         FROM purchase_payments pp
         JOIN purchase_orders po ON po.id = pp.purchase_id
         JOIN sales sl ON sl.id = po.sale_id
         LEFT JOIN customers c ON c.id = sl.customer_id
         LEFT JOIN suppliers s ON s.id = pp.supplier_id
         ORDER BY po.id",
    )
    .fetch_all(pool)
    .await
    .map_err(|e| AppError::internal(e.to_string()))
}

pub async fn get_financial_overview(pool: &sqlx::PgPool) -> Result<Value, AppError> {
    let overview = sqlx::query_scalar::<_, Json<Value>>(
        r#"
        WITH purchase_payment_costs AS (
            SELECT
                pp.id,
                pp.purchase_id,
                pp.cost_type,
                pp.amount::float8 AS amount,
                pp.method,
                pp.status,
                pp.requested_at,
                pp.approved_at,
                po.general_number,
                COALESCE(sup.name, '') AS supplier_name,
                COALESCE(c.name, '') AS customer_name,
                CASE
                    WHEN pp.status IN ('Pago', 'Pagamento Registrado', 'Aprovado') THEN TRUE
                    ELSE FALSE
                END AS paid,
                'Pagamento solicitado'::text AS source_note
            FROM purchase_payments pp
            JOIN purchase_orders po ON po.id = pp.purchase_id
            JOIN sales sl ON sl.id = po.sale_id
            LEFT JOIN customers c ON c.id = sl.customer_id
            LEFT JOIN suppliers sup ON sup.id = pp.supplier_id
        ),
        purchase_order_costs AS (
            SELECT
                po.id AS purchase_id,
                costs.cost_key,
                costs.cost_type,
                costs.group_key,
                costs.amount::float8 AS amount,
                COALESCE(NULLIF(po.payment_method, ''), 'sem forma definida') AS method,
                po.created_at,
                COALESCE(costs.deadline::timestamptz, po.created_at + INTERVAL '30 days') AS due_at,
                po.general_number,
                COALESCE(sup.name, '') AS supplier_name,
                COALESCE(c.name, '') AS customer_name
            FROM purchase_orders po
            JOIN sales sl ON sl.id = po.sale_id
            LEFT JOIN customers c ON c.id = sl.customer_id
            CROSS JOIN LATERAL (
                VALUES
                    ('material', 'Material previsto da compra', 'cmv', po.material_total_cost, po.material_supplier_id, po.material_deadline),
                    ('gravacao', 'Gravação prevista da compra', 'gravacoes', po.engraving_cost, po.engraving_supplier_id, po.engraving_deadline),
                    ('frete', 'Frete do fornecedor previsto', 'transporte', po.freight_cost, po.material_supplier_id, po.material_deadline),
                    ('outros', 'Outros custos previstos', 'variaveis', po.other_cost, po.material_supplier_id, po.material_deadline)
            ) AS costs(cost_key, cost_type, group_key, amount, supplier_id, deadline)
            LEFT JOIN suppliers sup ON sup.id = costs.supplier_id
            WHERE COALESCE(costs.amount, 0) > 0
              AND NOT EXISTS (
                  SELECT 1 FROM purchase_payments pp
                  WHERE pp.purchase_id = po.id
                    AND (
                        (costs.cost_key = 'material' AND lower(pp.cost_type) NOT LIKE '%grav%' AND lower(pp.cost_type) NOT LIKE '%frete%' AND lower(pp.cost_type) NOT LIKE '%transport%' AND lower(pp.cost_type) NOT LIKE '%outro%')
                        OR (costs.cost_key = 'gravacao' AND lower(pp.cost_type) LIKE '%grav%')
                        OR (costs.cost_key = 'frete' AND (lower(pp.cost_type) LIKE '%frete%' OR lower(pp.cost_type) LIKE '%transport%'))
                        OR (costs.cost_key = 'outros' AND lower(pp.cost_type) LIKE '%outro%')
                    )
              )
        ),
        sales_costs AS (
            SELECT
                s.id AS sale_id,
                (
                    COALESCE(SUM(pp.amount), 0)
                    + COALESCE(MAX(po.material_total_cost), 0)
                    + COALESCE(MAX(po.engraving_cost), 0)
                    + COALESCE(MAX(po.freight_cost), 0)
                    + COALESCE(MAX(po.other_cost), 0)
                )::float8 AS expected_costs
            FROM sales s
            LEFT JOIN purchase_orders po ON po.sale_id = s.id
            LEFT JOIN purchase_payments pp ON pp.purchase_id = po.id
            GROUP BY s.id
        ),
        sale_receipt_status AS (
            SELECT
                s.id AS sale_id,
                COUNT(spr.id)::int AS receipt_count,
                COUNT(*) FILTER (WHERE spr.status = 'validated')::int AS validated_count,
                COUNT(*) FILTER (WHERE spr.status = 'pending')::int AS pending_count,
                COUNT(*) FILTER (WHERE spr.status = 'rejected')::int AS rejected_count,
                MAX(spr.validated_at) AS validated_at,
                MAX(spr.created_at) AS last_receipt_at
            FROM sales s
            LEFT JOIN sale_payment_receipts spr ON spr.sale_id = s.id
            GROUP BY s.id
        ),
        movements AS (
            SELECT
                ('purchase-payment-' || p.id)::text AS id,
                p.id::bigint AS sort_id,
                ('LC-' || LPAD(p.id::text, 5, '0'))::text AS num,
                CASE
                    WHEN p.paid = FALSE THEN 'fornecedores-prazo'
                    WHEN lower(p.cost_type) LIKE '%grav%' OR lower(p.cost_type) LIKE '%embal%' THEN 'gravacoes'
                    WHEN lower(p.cost_type) LIKE '%frete%' OR lower(p.cost_type) LIKE '%transport%' THEN 'transporte'
                    WHEN lower(p.cost_type) LIKE '%comiss%' THEN 'comissoes'
                    ELSE 'cmv'
                END AS grupo,
                'saida'::text AS tipo,
                to_char(date_trunc('month', COALESCE(p.approved_at, p.requested_at)), 'YYYY-MM') AS mes,
                to_char(COALESCE(p.approved_at, p.requested_at) AT TIME ZONE 'America/Sao_Paulo', 'DD/MM/YYYY') AS data,
                to_char(COALESCE(p.approved_at, p.requested_at) AT TIME ZONE 'America/Sao_Paulo', 'HH24:MI') AS hora,
                (p.cost_type || ' · ' || p.general_number)::text AS descricao,
                COALESCE(NULLIF(p.supplier_name, ''), p.customer_name, 'Fornecedor não informado') AS parte,
                p.general_number AS pedido,
                p.method AS forma,
                to_char((p.requested_at + INTERVAL '30 days') AT TIME ZONE 'America/Sao_Paulo', 'DD/MM/YYYY') AS vencimento,
                NULL::text AS quantidade,
                p.amount AS valor,
                p.paid AS pago,
                'Financeiro'::text AS lancado_por,
                'purchase_payment'::text AS origem_tipo,
                p.id::bigint AS origem_id,
                (NOT p.paid)::boolean AS acionavel,
                p.source_note AS informe,
                jsonb_build_object() AS venda
            FROM purchase_payment_costs p

            UNION ALL

            SELECT
                ('purchase-cost-' || p.purchase_id || '-' || p.cost_key)::text AS id,
                (p.purchase_id::bigint * 10)::bigint AS sort_id,
                ('PC-' || LPAD(p.purchase_id::text, 5, '0'))::text AS num,
                p.group_key AS grupo,
                'saida'::text AS tipo,
                to_char(date_trunc('month', p.due_at), 'YYYY-MM') AS mes,
                to_char(p.created_at AT TIME ZONE 'America/Sao_Paulo', 'DD/MM/YYYY') AS data,
                to_char(p.created_at AT TIME ZONE 'America/Sao_Paulo', 'HH24:MI') AS hora,
                (p.cost_type || ' · ' || p.general_number)::text AS descricao,
                COALESCE(NULLIF(p.supplier_name, ''), p.customer_name, 'Fornecedor não informado') AS parte,
                p.general_number AS pedido,
                p.method AS forma,
                to_char(p.due_at AT TIME ZONE 'America/Sao_Paulo', 'DD/MM/YYYY') AS vencimento,
                NULL::text AS quantidade,
                p.amount AS valor,
                FALSE AS pago,
                'Compras'::text AS lancado_por,
                'purchase_order_cost'::text AS origem_tipo,
                p.purchase_id::bigint AS origem_id,
                FALSE AS acionavel,
                'Custo existe na ordem de compra, mas ainda não há solicitação de pagamento lançada.'::text AS informe,
                jsonb_build_object() AS venda
            FROM purchase_order_costs p

            UNION ALL

            SELECT
                ('sale-' || s.id)::text AS id,
                s.id::bigint AS sort_id,
                ('RV-' || LPAD(s.id::text, 5, '0'))::text AS num,
                CASE
                    WHEN EXISTS (
                        SELECT 1 FROM sale_payment_receipts spr
                        WHERE spr.sale_id = s.id AND spr.status = 'validated'
                    ) THEN 'venda-mes'
                    ELSE 'venda-prox'
                END AS grupo,
                'entrada'::text AS tipo,
                to_char(date_trunc('month', s.created_at), 'YYYY-MM') AS mes,
                to_char(s.created_at AT TIME ZONE 'America/Sao_Paulo', 'DD/MM/YYYY') AS data,
                to_char(s.created_at AT TIME ZONE 'America/Sao_Paulo', 'HH24:MI') AS hora,
                ('Venda ' || COALESCE(po.general_number, s.id::text))::text AS descricao,
                COALESCE(c.name, 'Cliente não informado') AS parte,
                COALESCE(po.general_number, s.id::text) AS pedido,
                COALESCE(s.payment_method, '') AS forma,
                CASE
                    WHEN s.first_installment_start IS NULL THEN ''
                    ELSE to_char(s.first_installment_start AT TIME ZONE 'America/Sao_Paulo', 'DD/MM/YYYY')
                END AS vencimento,
                NULL::text AS quantidade,
                COALESCE(s.total_value, 0)::float8 AS valor,
                EXISTS (
                    SELECT 1 FROM sale_payment_receipts spr
                    WHERE spr.sale_id = s.id AND spr.status = 'validated'
                ) AS pago,
                'Sistema'::text AS lancado_por,
                'sale'::text AS origem_tipo,
                s.id::bigint AS origem_id,
                FALSE AS acionavel,
                CASE
                    WHEN COALESCE(rs.validated_count, 0) > 0 THEN 'Recebimento validado pelo financeiro.'
                    WHEN COALESCE(rs.pending_count, 0) > 0 THEN 'Comprovante enviado pelo cliente, aguardando validação.'
                    WHEN COALESCE(rs.rejected_count, 0) > 0 THEN 'Último comprovante rejeitado; venda segue a receber.'
                    ELSE 'Venda sem comprovante validado; permanece em a receber.'
                END AS informe,
                jsonb_build_object(
                    'client', COALESCE(c.name, ''),
                    'seller', COALESCE(u.full_name, ''),
                    'profit', COALESCE(s.total_value, 0)::float8 - COALESCE(sc.expected_costs, 0),
                    'margin', CASE WHEN COALESCE(s.total_value, 0) > 0 THEN (COALESCE(s.total_value, 0)::float8 - COALESCE(sc.expected_costs, 0)) / COALESCE(s.total_value, 0)::float8 ELSE 0 END,
                    'freight_paid', COALESCE(po.freight_cost, 0)::float8,
                    'freight_charged', 0,
                    'logistics', 0,
                    'loss', 0,
                    'advertising', 0,
                    'labor', 0,
                    'receipt_count', COALESCE(rs.receipt_count, 0),
                    'validated_receipts', COALESCE(rs.validated_count, 0),
                    'pending_receipts', COALESCE(rs.pending_count, 0),
                    'rejected_receipts', COALESCE(rs.rejected_count, 0),
                    'financial_status', COALESCE(fa.status, '')
                ) AS venda
            FROM sales s
            LEFT JOIN customers c ON c.id = s.customer_id
            LEFT JOIN users u ON u.id = s.seller_id
            LEFT JOIN purchase_orders po ON po.sale_id = s.id
            LEFT JOIN sales_costs sc ON sc.sale_id = s.id
            LEFT JOIN sale_receipt_status rs ON rs.sale_id = s.id
            LEFT JOIN financial_analyses fa ON fa.sale_id = s.id
            WHERE COALESCE(s.total_value, 0) > 0
        ),
        month_list AS (
            SELECT DISTINCT mes FROM movements
        ),
        insights AS (
            SELECT jsonb_build_object(
                'open_payables_count', COUNT(*) FILTER (WHERE tipo = 'saida' AND pago = FALSE),
                'open_payables_value', COALESCE(SUM(valor) FILTER (WHERE tipo = 'saida' AND pago = FALSE), 0),
                'open_receivables_count', COUNT(*) FILTER (WHERE tipo = 'entrada' AND pago = FALSE),
                'open_receivables_value', COALESCE(SUM(valor) FILTER (WHERE tipo = 'entrada' AND pago = FALSE), 0),
                'paid_movements_count', COUNT(*) FILTER (WHERE pago = TRUE),
                'paid_movements_value', COALESCE(SUM(valor) FILTER (WHERE pago = TRUE), 0),
                'costs_without_payment_request_count', COUNT(*) FILTER (WHERE origem_tipo = 'purchase_order_cost'),
                'costs_without_payment_request_value', COALESCE(SUM(valor) FILTER (WHERE origem_tipo = 'purchase_order_cost'), 0),
                'sales_waiting_receipt_count', COUNT(*) FILTER (WHERE origem_tipo = 'sale' AND pago = FALSE),
                'sales_waiting_receipt_value', COALESCE(SUM(valor) FILTER (WHERE origem_tipo = 'sale' AND pago = FALSE), 0)
            ) AS data
            FROM movements
        ),
        financial_orders AS (
            SELECT
                s.id AS sale_id,
                po.id AS purchase_id,
                COALESCE(po.general_number, s.id::text) AS order_number,
                COALESCE(c.name, 'Cliente não informado') AS customer_name,
                COALESCE(seller.full_name, 'Vendedor não informado') AS seller_name,
                COALESCE(buyer.full_name, '') AS buyer_name,
                COALESCE(s.status, '') AS sale_status,
                COALESCE(po.status, 'Sem compra liberada') AS purchase_status,
                to_char(s.created_at AT TIME ZONE 'America/Sao_Paulo', 'DD/MM/YYYY HH24:MI') AS created_at_br,
                to_char(po.production_released_at AT TIME ZONE 'America/Sao_Paulo', 'DD/MM/YYYY HH24:MI') AS production_released_at_br,
                COALESCE(s.total_value, 0)::float8 AS sale_value,
                CASE
                    WHEN COALESCE(payments.total, 0) > 0 THEN COALESCE(payments.total, 0)::float8
                    ELSE (
                        COALESCE(po.material_total_cost, 0)
                        + COALESCE(po.engraving_cost, 0)
                        + COALESCE(po.freight_cost, 0)
                        + COALESCE(po.other_cost, 0)
                    )::float8
                END AS expected_cost,
                COALESCE(payments.paid_total, 0)::float8 AS paid_cost,
                COALESCE(payments.pending_total, 0)::float8 AS pending_cost,
                COALESCE(payments.total_count, 0)::int AS payments_count,
                COALESCE(payments.pending_count, 0)::int AS pending_payments_count,
                COALESCE(receipts.receipt_count, 0)::int AS receipt_count,
                COALESCE(receipts.validated_count, 0)::int AS validated_receipts,
                COALESCE(receipts.pending_count, 0)::int AS pending_receipts,
                COALESCE(items.items_count, 0)::int AS items_count,
                COALESCE(items.items_json, '[]'::jsonb) AS items_json,
                COALESCE(payments.payments_json, '[]'::jsonb) AS payments_json,
                CASE
                    WHEN po.id IS NULL THEN 'Venda registrada, mas ainda não foi liberada para Compras.'
                    WHEN COALESCE(payments.total_count, 0) = 0
                         AND (
                            COALESCE(po.material_total_cost, 0)
                            + COALESCE(po.engraving_cost, 0)
                            + COALESCE(po.freight_cost, 0)
                            + COALESCE(po.other_cost, 0)
                         ) > 0 THEN 'Compra possui custo previsto, mas não há solicitação de pagamento.'
                    WHEN COALESCE(payments.pending_count, 0) > 0 THEN 'Há pagamento de fornecedor pendente de baixa.'
                    WHEN COALESCE(receipts.validated_count, 0) = 0 THEN 'Venda ainda sem recebimento validado.'
                    ELSE 'Pedido com informações financeiras conciliadas pelo backend.'
                END AS report
            FROM sales s
            LEFT JOIN purchase_orders po ON po.sale_id = s.id
            LEFT JOIN customers c ON c.id = s.customer_id
            LEFT JOIN users seller ON seller.id = s.seller_id
            LEFT JOIN users buyer ON buyer.id = po.buyer_id
            LEFT JOIN LATERAL (
                SELECT
                    COUNT(*) AS total_count,
                    COUNT(*) FILTER (WHERE pp.status NOT IN ('Pago', 'Pagamento Registrado', 'Aprovado')) AS pending_count,
                    COALESCE(SUM(pp.amount), 0) AS total,
                    COALESCE(SUM(pp.amount) FILTER (WHERE pp.status IN ('Pago', 'Pagamento Registrado', 'Aprovado')), 0) AS paid_total,
                    COALESCE(SUM(pp.amount) FILTER (WHERE pp.status NOT IN ('Pago', 'Pagamento Registrado', 'Aprovado')), 0) AS pending_total,
                    jsonb_agg(jsonb_build_object(
                        'id', pp.id,
                        'cost_type', pp.cost_type,
                        'supplier', COALESCE(sup.name, ''),
                        'amount', pp.amount::float8,
                        'method', pp.method,
                        'status', pp.status,
                        'requested_at', to_char(pp.requested_at AT TIME ZONE 'America/Sao_Paulo', 'DD/MM/YYYY HH24:MI'),
                        'approved_at', CASE WHEN pp.approved_at IS NULL THEN '' ELSE to_char(pp.approved_at AT TIME ZONE 'America/Sao_Paulo', 'DD/MM/YYYY HH24:MI') END
                    ) ORDER BY pp.id) AS payments_json
                FROM purchase_payments pp
                LEFT JOIN suppliers sup ON sup.id = pp.supplier_id
                WHERE pp.purchase_id = po.id
            ) payments ON TRUE
            LEFT JOIN LATERAL (
                SELECT
                    COUNT(*) AS receipt_count,
                    COUNT(*) FILTER (WHERE spr.status = 'validated') AS validated_count,
                    COUNT(*) FILTER (WHERE spr.status = 'pending') AS pending_count
                FROM sale_payment_receipts spr
                WHERE spr.sale_id = s.id
            ) receipts ON TRUE
            LEFT JOIN LATERAL (
                SELECT
                    COUNT(*) AS items_count,
                    jsonb_agg(jsonb_build_object(
                        'product', COALESCE(p.product_name, 'Item sem cadastro'),
                        'quantity', si.quantity,
                        'unit_price', si.unit_price::float8,
                        'total_price', si.total_price::float8,
                        'estimated_cost', (si.quantity * COALESCE(p.cost_price, 0))::float8,
                        'supplier_id', p.supplier_id
                    ) ORDER BY si.id) AS items_json
                FROM sale_items si
                LEFT JOIN products p ON p.id = si.product_id
                WHERE si.sale_id = s.id
            ) items ON TRUE
            WHERE COALESCE(s.total_value, 0) > 0
        )
        SELECT jsonb_build_object(
            'generated_at', NOW(),
            'opening_balances', jsonb_build_object(),
            'insights', COALESCE((SELECT data FROM insights), jsonb_build_object()),
            'orders', COALESCE((
                SELECT jsonb_agg(jsonb_build_object(
                    'sale_id', sale_id,
                    'purchase_id', purchase_id,
                    'order_number', order_number,
                    'customer', customer_name,
                    'seller', seller_name,
                    'buyer', buyer_name,
                    'sale_status', sale_status,
                    'purchase_status', purchase_status,
                    'created_at', created_at_br,
                    'production_released_at', COALESCE(production_released_at_br, ''),
                    'sale_value', sale_value,
                    'expected_cost', expected_cost,
                    'paid_cost', paid_cost,
                    'pending_cost', pending_cost,
                    'open_cost', GREATEST(expected_cost - paid_cost, 0),
                    'profit', sale_value - expected_cost,
                    'margin', CASE WHEN sale_value > 0 THEN (sale_value - expected_cost) / sale_value ELSE 0 END,
                    'payments_count', payments_count,
                    'pending_payments_count', pending_payments_count,
                    'receipt_count', receipt_count,
                    'validated_receipts', validated_receipts,
                    'pending_receipts', pending_receipts,
                    'items_count', items_count,
                    'items', items_json,
                    'payments', payments_json,
                    'report', report
                ) ORDER BY sale_id DESC)
                FROM financial_orders
            ), '[]'::jsonb),
            'months', COALESCE((SELECT jsonb_agg(mes ORDER BY mes DESC) FROM month_list), '[]'::jsonb),
            'movements', COALESCE((
                SELECT jsonb_agg(jsonb_build_object(
                    'id', id,
                    'num', num,
                    'group', grupo,
                    'type', tipo,
                    'month', mes,
                    'date', data,
                    'time', hora,
                    'description', descricao,
                    'party', parte,
                    'order', pedido,
                    'form', forma,
                    'due_date', vencimento,
                    'quantity', quantidade,
                    'value', valor,
                    'paid', pago,
                    'launched_by', lancado_por,
                    'source_type', origem_tipo,
                    'source_id', origem_id,
                    'actionable', acionavel,
                    'report', informe,
                    'sale', venda
                ) ORDER BY mes DESC, data, hora, sort_id)
                FROM movements
            ), '[]'::jsonb)
        ) AS overview
        "#
    )
    .fetch_one(pool)
    .await
    .map_err(|e| AppError::internal(e.to_string()))?;

    Ok(overview.0)
}

pub async fn get_client_notes(pool: &sqlx::PgPool) -> Result<Value, AppError> {
    let data = sqlx::query_scalar::<_, Json<Value>>(
        r#"
        WITH eligible AS (
            SELECT s.id AS sale_id,
                   COALESCE(po.general_number, 'PED-' || LPAD(s.id::text, 6, '0')) AS order_number,
                   COALESCE(c.name, 'Cliente não informado') AS customer,
                   COALESCE(s.total_value, 0)::float8 AS sale_value,
                   COALESCE(s.invoice_email, NULLIF(c.email, ''), '') AS invoice_email,
                   COALESCE(s.financial_email, NULLIF(c.contact_financial_email, ''), NULLIF(c.email, ''), '') AS financial_email,
                   COALESCE(s.status, '') AS sale_status,
                   COALESCE(po.status, 'Sem compra liberada') AS purchase_status,
                   COALESCE(seller.full_name, 'Vendedor não informado') AS seller,
                   COALESCE(c.cnpj, c.cpf, '') AS document,
                   COALESCE(w.status, 'financeiro') AS workflow_status,
                   COALESCE(w.sent_email, '') AS sent_email,
                   w.sent_at,
                   COALESCE(fiscal.documents, '[]'::jsonb) AS documents
            FROM sales s
            LEFT JOIN purchase_orders po ON po.sale_id = s.id
            LEFT JOIN production_orders prod ON prod.sale_id = s.id
            LEFT JOIN customers c ON c.id = s.customer_id
            LEFT JOIN users seller ON seller.id = s.seller_id
            LEFT JOIN client_note_workflows w ON w.sale_id = s.id
            LEFT JOIN LATERAL (
                SELECT jsonb_agg(jsonb_build_object(
                    'id', pfd.id,
                    'type', pfd.document_type,
                    'number', pfd.document_number,
                    'access_key', pfd.access_key,
                    'file_url', pfd.file_url,
                    'issued_at', COALESCE(to_char(pfd.issued_at AT TIME ZONE 'America/Sao_Paulo', 'DD/MM/YYYY'), '')
                ) ORDER BY pfd.issued_at DESC NULLS LAST, pfd.id DESC) AS documents
                FROM production_fiscal_documents pfd
                WHERE pfd.production_order_id = prod.id
            ) fiscal ON TRUE
            WHERE COALESCE(s.total_value, 0) > 0
              AND (prod.id IS NOT NULL OR po.production_released_at IS NOT NULL)
        )
        SELECT jsonb_build_object(
            'generated_at', NOW(),
            'orders', COALESCE((SELECT jsonb_agg(jsonb_build_object(
                'sale_id', sale_id,
                'order_number', order_number,
                'customer', customer,
                'sale_value', sale_value,
                'invoice_email', invoice_email,
                'financial_email', financial_email,
                'sale_status', sale_status,
                'purchase_status', purchase_status,
                'seller', seller,
                'document', document,
                'status', workflow_status,
                'sent_email', sent_email,
                'sent_at', COALESCE(to_char(sent_at AT TIME ZONE 'America/Sao_Paulo', 'DD/MM/YYYY HH24:MI'), ''),
                'documents', documents,
                'report', CASE
                    WHEN jsonb_array_length(documents) > 0 THEN 'Documento fiscal disponível no backend.'
                    WHEN workflow_status = 'financeiro' THEN 'Pedido liberado para produção, aguardando anexação da nota fiscal.'
                    WHEN workflow_status = 'adm' THEN 'Conferido pelo Financeiro; aguardando envio ao cliente.'
                    ELSE 'Nota encaminhada ao cliente, sem documento fiscal anexado.'
                END
            ) ORDER BY sale_id DESC) FROM eligible), '[]'::jsonb)
        )"#
    ).fetch_one(pool).await.map_err(|e| AppError::internal(e.to_string()))?;
    Ok(data.0)
}

pub async fn update_client_note_status(
    pool: &sqlx::PgPool,
    sale_id: i32,
    status: &str,
    sent_email: &str,
    user_id: i32,
) -> Result<Value, AppError> {
    let row = sqlx::query_scalar::<_, Json<Value>>(
        "INSERT INTO client_note_workflows (sale_id, status, sent_email, sent_at, updated_by)
         VALUES ($1, $2, $3, CASE WHEN $2 = 'enviadas' THEN NOW() ELSE NULL END, $4)
         ON CONFLICT (sale_id) DO UPDATE SET status = EXCLUDED.status,
             sent_email = EXCLUDED.sent_email,
             sent_at = CASE WHEN EXCLUDED.status = 'enviadas' THEN NOW() ELSE NULL END,
             updated_by = EXCLUDED.updated_by, updated_at = NOW()
         RETURNING jsonb_build_object('sale_id', sale_id, 'status', status, 'sent_email', sent_email)"
    ).bind(sale_id).bind(status).bind(sent_email).bind(user_id)
    .fetch_optional(pool).await.map_err(|e| AppError::internal(e.to_string()))?
    .ok_or_else(|| AppError::not_found("Venda"))?;
    Ok(row.0)
}

pub async fn get_icms_credit(pool: &sqlx::PgPool, month: Option<&str>) -> Result<Value, AppError> {
    let month = month.filter(|value| value.len() == 7 && value.as_bytes().get(4) == Some(&b'-'));
    let data = sqlx::query_scalar::<_, Json<Value>>(
        r#"
        WITH entries AS (
            SELECT e.id, e.invoice_number, e.invoice_key,
                   to_char(e.issued_at AT TIME ZONE 'America/Sao_Paulo', 'YYYY-MM-DD') AS issued_date,
                   to_char(e.issued_at AT TIME ZONE 'America/Sao_Paulo', 'YYYY-MM') AS issued_month,
                   e.tax_base::float8 AS tax_base, e.aliquota::float8 AS aliquota,
                   e.icms_value::float8 AS icms_value, e.xml_url,
                   COALESCE(sup.name, 'Fornecedor não informado') AS supplier,
                   COALESCE(po.general_number, '') AS purchase_number,
                   COALESCE(pfd.document_number, e.invoice_number) AS fiscal_number,
                   p.name AS parameter_name,
                   p.credit_percent::float8 AS credit_percent,
                   CASE WHEN p.id IS NULL THEN 0::float8 ELSE (e.icms_value * p.credit_percent / 100)::float8 END AS credit_value,
                   CASE
                       WHEN p.id IS NULL THEN 'Alíquota sem parâmetro ativo; crédito precisa de revisão fiscal.'
                       WHEN p.credit_percent = 0 THEN 'Parâmetro ativo, mas esta operação não gera crédito.'
                       ELSE 'Crédito calculado com base no valor de ICMS informado na nota.'
                   END AS report
            FROM financial_icms_entries e
            LEFT JOIN suppliers sup ON sup.id = e.supplier_id
            LEFT JOIN purchase_orders po ON po.id = e.purchase_id
            LEFT JOIN financial_icms_parameters p ON p.active = TRUE AND abs(p.aliquota - e.aliquota) < 0.0001
            LEFT JOIN production_orders prod ON prod.id = e.production_order_id
            LEFT JOIN LATERAL (
                SELECT document_number
                FROM production_fiscal_documents
                WHERE production_order_id = prod.id
                  AND document_number = e.invoice_number
                ORDER BY id DESC LIMIT 1
            ) pfd ON TRUE
            WHERE ($1::text IS NULL OR to_char(e.issued_at AT TIME ZONE 'America/Sao_Paulo', 'YYYY-MM') = $1)
        ),
        months AS (
            SELECT DISTINCT to_char(issued_at AT TIME ZONE 'America/Sao_Paulo', 'YYYY-MM') AS issued_month
            FROM financial_icms_entries
            ORDER BY issued_month DESC
        )
        SELECT jsonb_build_object(
            'generated_at', NOW(),
            'selected_month', COALESCE($1, to_char(NOW() AT TIME ZONE 'America/Sao_Paulo', 'YYYY-MM')),
            'months', COALESCE((SELECT jsonb_agg(issued_month ORDER BY issued_month DESC) FROM months), '[]'::jsonb),
            'parameters', COALESCE((SELECT jsonb_agg(jsonb_build_object(
                'id', id, 'name', name, 'aliquota', aliquota::float8,
                'credit_percent', credit_percent::float8, 'description', description,
                'active', active
            ) ORDER BY aliquota DESC) FROM financial_icms_parameters WHERE active = TRUE), '[]'::jsonb),
            'summary', jsonb_build_object(
                'notes_count', (SELECT COUNT(*) FROM entries),
                'notes_with_credit_count', (SELECT COUNT(*) FROM entries WHERE credit_value > 0),
                'icms_total', COALESCE((SELECT SUM(icms_value) FROM entries), 0),
                'credit_total', COALESCE((SELECT SUM(credit_value) FROM entries), 0),
                'without_parameter_count', (SELECT COUNT(*) FROM entries WHERE parameter_name IS NULL)
            ),
            'entries', COALESCE((SELECT jsonb_agg(jsonb_build_object(
                'id', id, 'invoice_number', fiscal_number, 'issued_date', issued_date,
                'supplier', supplier, 'purchase_number', purchase_number,
                'tax_base', tax_base, 'aliquota', aliquota, 'icms_value', icms_value,
                'parameter_name', COALESCE(parameter_name, ''), 'credit_percent', COALESCE(credit_percent, 0),
                'credit_value', credit_value, 'xml_url', xml_url, 'report', report
            ) ORDER BY issued_date DESC, id DESC) FROM entries), '[]'::jsonb)
        )"#
    ).bind(month).fetch_one(pool).await.map_err(|e| AppError::internal(e.to_string()))?;
    Ok(data.0)
}

pub async fn create_icms_entry(
    pool: &sqlx::PgPool,
    input: &crate::handlers::purchase::IcmsEntryInput,
    user_id: i32,
) -> Result<Value, AppError> {
    let row = sqlx::query_scalar::<_, Json<Value>>(
        "INSERT INTO financial_icms_entries
            (purchase_id, production_order_id, supplier_id, invoice_number, invoice_key,
             issued_at, tax_base, aliquota, icms_value, xml_url, created_by)
         VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)
         RETURNING jsonb_build_object('id', id, 'invoice_number', invoice_number, 'issued_at', issued_at)"
    ).bind(input.purchase_id).bind(input.production_order_id).bind(input.supplier_id)
    .bind(&input.invoice_number).bind(&input.invoice_key).bind(input.issued_at)
    .bind(input.tax_base).bind(input.aliquota).bind(input.icms_value)
    .bind(&input.xml_url).bind(user_id)
    .fetch_one(pool).await.map_err(|e| AppError::internal(e.to_string()))?;
    Ok(row.0)
}

pub async fn release_sale_to_purchases(
    pool: &sqlx::PgPool,
    sale_id: i32,
    input: &PurchaseReleaseInput,
) -> Result<PurchaseOrder, AppError> {
    let mut tx = pool.begin().await.map_err(|e| AppError::internal(e.to_string()))?;

    let sale = sqlx::query_as::<_, (i32, String, Option<String>, Option<String>, bool, Option<i32>, f64, f64)>(
        "SELECT s.id, COALESCE(c.name, ''),
                s.payment_method,
                COALESCE(NULLIF(s.internal_notes, ''), NULLIF(s.external_notes, '')) AS notes,
                COALESCE(BOOL_OR(COALESCE(si.engravings, '[]'::jsonb) <> '[]'::jsonb OR COALESCE(si.personalization_type, '') <> ''), FALSE) AS has_engraving,
                MIN(p.supplier_id) AS material_supplier_id,
                COALESCE(MIN(p.cost_price), 0)::float8 AS material_unit_cost,
                COALESCE(SUM(si.quantity * p.cost_price), 0)::float8 AS material_total_cost
         FROM sales s
         LEFT JOIN customers c ON c.id = s.customer_id
         LEFT JOIN sale_items si ON si.sale_id = s.id
         LEFT JOIN products p ON p.id = si.product_id
         WHERE s.id = $1
         GROUP BY s.id, c.name, s.payment_method, s.internal_notes, s.external_notes",
    )
    .bind(sale_id)
    .fetch_optional(&mut *tx)
    .await
    .map_err(|e| AppError::internal(e.to_string()))?
    .ok_or_else(|| AppError::not_found("Venda"))?;

    let general_number = format!("{:06}-{}", sale.0, sale.1.chars().take(20).collect::<String>().replace(' ', "_"));
    let is_sample = input.is_sample.unwrap_or(false);
    let sample_has_engraving = input.sample_has_engraving.unwrap_or(false);
    let has_engraving = if is_sample { sample_has_engraving } else { sale.4 };

    let po_id = sqlx::query_as::<_, (i32,)>(
        "INSERT INTO purchase_orders (
                sale_id, general_number, status, material_supplier_id,
                is_sample, sample_has_engraving, has_engraving, corel_required,
                material_unit_cost, material_total_cost, payment_method, commercial_notes
         )
         VALUES ($1, $2, 'Pendente de Compra', $3, $4, $5, $6, $6, $7, $8, $9, $10)
         ON CONFLICT (sale_id) DO UPDATE SET
                status = 'Pendente de Compra',
                status_updated_at = NOW(),
                material_supplier_id = EXCLUDED.material_supplier_id,
                is_sample = EXCLUDED.is_sample,
                sample_has_engraving = EXCLUDED.sample_has_engraving,
                has_engraving = EXCLUDED.has_engraving,
                corel_required = EXCLUDED.corel_required,
                material_unit_cost = EXCLUDED.material_unit_cost,
                material_total_cost = EXCLUDED.material_total_cost,
                payment_method = EXCLUDED.payment_method,
                commercial_notes = EXCLUDED.commercial_notes,
                updated_at = NOW()
         RETURNING id",
    )
    .bind(sale.0)
    .bind(&general_number)
    .bind(sale.5)
    .bind(is_sample)
    .bind(sample_has_engraving)
    .bind(has_engraving)
    .bind(sale.6)
    .bind(sale.7)
    .bind(sale.2.as_deref().unwrap_or(""))
    .bind(sale.3.as_deref().unwrap_or(""))
    .fetch_one(&mut *tx)
    .await
    .map_err(|e| AppError::internal(e.to_string()))?;

    sqlx::query(
        "UPDATE sales
         SET status = 'Pendente de Compra',
             released_to_purchases_at = NOW(),
             updated_at = NOW()
         WHERE id = $1",
    )
    .bind(sale_id)
    .execute(&mut *tx)
    .await
    .map_err(|e| AppError::internal(e.to_string()))?;

    sqlx::query(
        "INSERT INTO purchase_history (purchase_id, action, from_status, to_status, user_name)
         SELECT $1, 'pedido liberado pelo comercial', '', 'Pendente de Compra', 'Sistema'
         WHERE NOT EXISTS (
             SELECT 1 FROM purchase_history
             WHERE purchase_id = $1 AND action = 'pedido liberado pelo comercial'
         )",
    )
    .bind(po_id.0)
    .execute(&mut *tx)
    .await
    .ok();

    tx.commit().await.map_err(|e| AppError::internal(e.to_string()))?;
    get_by_id(pool, po_id.0).await
}

pub async fn update(pool: &sqlx::PgPool, id: i32, input: &PurchaseUpdateInput) -> Result<PurchaseOrder, AppError> {
    sqlx::query(UPDATE_PURCHASE_SQL)
        .bind(input.buyer_id).bind(input.material_supplier_id).bind(input.engraving_supplier_id)
        .bind(input.is_sample.unwrap_or(false))
        .bind(input.sample_has_engraving.unwrap_or(false))
        .bind(input.has_engraving.unwrap_or(false))
        .bind(input.material_unit_cost.unwrap_or(0.0))
        .bind(input.material_total_cost.unwrap_or(0.0))
        .bind(input.engraving_cost.unwrap_or(0.0))
        .bind(input.freight_cost.unwrap_or(0.0))
        .bind(input.other_cost.unwrap_or(0.0))
        .bind(input.buyer_discount.unwrap_or(0.0))
        .bind(&input.negotiation_contact).bind(&input.negotiation_notes)
        .bind(input.material_deadline).bind(input.engraving_deadline)
        .bind(&input.payment_method)
        .bind(input.requires_advance_payment.unwrap_or(false))
        .bind(input.first_piece_required.unwrap_or(false))
        .bind(&input.commercial_notes).bind(&input.purchase_notes)
        .bind(id)
        .execute(pool).await
        .map_err(|e| AppError::internal(e.to_string()))?;
    get_by_id(pool, id).await
}

pub async fn execute_action(
    pool: &sqlx::PgPool, purchase_id: i32, input: &PurchaseActionInput,
    user_id: i32,
) -> Result<PurchaseOrder, AppError> {
    let mut tx = pool.begin().await.map_err(|e| AppError::internal(e.to_string()))?;

    let current = sqlx::query_as::<_, PurchaseOrder>(
        &format!("{SELECT_PURCHASE_ALL} WHERE po.id = $1"),
    )
    .bind(purchase_id)
    .fetch_one(&mut *tx)
    .await
    .map_err(|e| AppError::internal(e.to_string()))?;

    let action = input.action.as_str();
    let (new_status, record_history) = match action {
        "request_corel" => ("Aguardando Arquivo Corel", true),
        "attach_corel" => ("Arquivo Corel Recebido", true),
        "send_material_email" => ("E-mail Material Enviado", true),
        "send_engraving_email" => ("E-mail Gravação Enviado", true),
        "finalize_purchase" => ("Pedido Comprado", true),
        "release_to_production" => ("Liberado para Produção", true),
        "release_production" => ("Liberado para Produção", true),
        "start_review" => ("Em Conferência de Compras", true),
        "accept_material" => (current.status.as_str(), true),
        "accept_engraving" => (current.status.as_str(), true),
        "first_piece_received" => ("Aguardando Aprovação da Gravação", true),
        "approve_first_piece" => ("Primeira Peça Aprovada", true),
        _ => return Err(AppError::bad_request(format!("Ação desconhecida: {action}"))),
    };

    if new_status != current.status {
        sqlx::query("UPDATE purchase_orders SET status = $1, status_updated_at = NOW() WHERE id = $2")
            .bind(new_status).bind(purchase_id)
            .execute(&mut *tx).await
            .map_err(|e| AppError::internal(e.to_string()))?;
    }

    if action == "request_corel" {
        sqlx::query("UPDATE purchase_orders SET corel_required = true, corel_requested_at = NOW(), corel_requested_by = $1 WHERE id = $2")
            .bind(user_id)
            .bind(purchase_id)
            .execute(&mut *tx).await
            .map_err(|e| AppError::internal(e.to_string()))?;
    }

    if action == "attach_corel" {
        sqlx::query("UPDATE purchase_orders SET corel_required = true, corel_attached_at = NOW(), corel_attached_by = $1 WHERE id = $2")
            .bind(user_id)
            .bind(purchase_id)
            .execute(&mut *tx).await
            .map_err(|e| AppError::internal(e.to_string()))?;
        sqlx::query(
            "INSERT INTO purchase_attachments (purchase_id, category, file_name, url, uploaded_by)
             VALUES ($1, 'corel', $2, $3, $4)",
        )
        .bind(purchase_id)
        .bind(input.file_name.as_deref().unwrap_or("Corel"))
        .bind(input.url.as_deref().unwrap_or(""))
        .bind(user_id)
        .execute(&mut *tx).await.ok();
    }

    if action == "release_to_production" || action == "release_production" {
        sqlx::query("UPDATE purchase_orders SET production_released_at = NOW() WHERE id = $1")
            .bind(purchase_id)
            .execute(&mut *tx).await
            .map_err(|e| AppError::internal(e.to_string()))?;
    }

    if action == "first_piece_received" {
        sqlx::query("UPDATE purchase_orders SET first_piece_status = 'Aguardando aprovação', first_piece_url = $1 WHERE id = $2")
            .bind(input.url.as_deref().unwrap_or(""))
            .bind(purchase_id)
            .execute(&mut *tx).await
            .map_err(|e| AppError::internal(e.to_string()))?;
        sqlx::query(
            "INSERT INTO purchase_attachments (purchase_id, category, file_name, url, uploaded_by)
             VALUES ($1, 'primeira_peca', $2, $3, $4)",
        )
        .bind(purchase_id)
        .bind(input.file_name.as_deref().unwrap_or("Primeira peça"))
        .bind(input.url.as_deref().unwrap_or(""))
        .bind(user_id)
        .execute(&mut *tx).await.ok();
    }

    if action == "approve_first_piece" {
        sqlx::query("UPDATE purchase_orders SET first_piece_status = 'Aprovada' WHERE id = $1")
            .bind(purchase_id)
            .execute(&mut *tx).await
            .map_err(|e| AppError::internal(e.to_string()))?;
    }

    if action == "accept_material" {
        sqlx::query("UPDATE purchase_orders SET material_accepted = true, material_accepted_at = NOW() WHERE id = $1")
            .bind(purchase_id)
            .execute(&mut *tx).await.ok();
    }
    if action == "accept_engraving" {
        sqlx::query("UPDATE purchase_orders SET engraving_accepted = true, engraving_accepted_at = NOW() WHERE id = $1")
            .bind(purchase_id)
            .execute(&mut *tx).await.ok();
    }

    if action.starts_with("send_") {
        let kind = if action.contains("material") { "material" } else { "gravacao" };
        sqlx::query(
            "INSERT INTO purchase_emails (purchase_id, kind, recipient, subject, body, observation, attachments, sent_by)
             VALUES ($1, $2, $3, $4, $5, $6, $7, $8)",
        )
        .bind(purchase_id).bind(kind)
        .bind(input.recipient.as_deref().unwrap_or(""))
        .bind("Fornecedor").bind("Pedido de compra")
        .bind(input.observation.as_deref().unwrap_or(""))
        .bind(&input.attachments).bind(user_id)
        .execute(&mut *tx).await.ok();
    }

    if record_history {
        let user_name = sqlx::query_scalar::<_, String>(
            "SELECT COALESCE(full_name, username, 'Sistema') FROM users WHERE id = $1",
        )
        .bind(user_id)
        .fetch_optional(&mut *tx)
        .await
        .map_err(|e| AppError::internal(e.to_string()))?
        .unwrap_or_else(|| "Sistema".to_string());
        sqlx::query(
            "INSERT INTO purchase_history (purchase_id, action, from_status, to_status, user_id, user_name)
             VALUES ($1, $2, $3, $4, $5, $6)",
        )
        .bind(purchase_id).bind(action)
        .bind(&current.status).bind(new_status).bind(user_id).bind(user_name)
        .execute(&mut *tx).await.ok();
    }

    tx.commit().await.map_err(|e| AppError::internal(e.to_string()))?;
    get_by_id(pool, purchase_id).await
}

pub async fn add_attachment(
    pool: &sqlx::PgPool, purchase_id: i32, category: &str,
    file_name: &str, url: &str, user_id: Option<i32>,
) -> Result<PurchaseAttachment, AppError> {
    sqlx::query_as::<_, PurchaseAttachment>(
        "INSERT INTO purchase_attachments (purchase_id, category, file_name, url, uploaded_by)
         VALUES ($1,$2,$3,$4,$5) RETURNING id, purchase_id, category, file_name, url, uploaded_by, created_at"
    )
    .bind(purchase_id).bind(category).bind(file_name).bind(url).bind(user_id)
    .fetch_one(pool).await
    .map_err(|e| AppError::internal(e.to_string()))
}

pub async fn add_payment(
    pool: &sqlx::PgPool, purchase_id: i32, input: &PurchasePaymentInput, user_id: i32,
) -> Result<PurchasePayment, AppError> {
    sqlx::query_as::<_, PurchasePayment>(
        "INSERT INTO purchase_payments (purchase_id, cost_type, supplier_id, amount, method,
                justification, receipt_url, requested_by)
         VALUES ($1,$2,$3,$4,$5,$6,$7,$8)
         RETURNING id, purchase_id, cost_type, supplier_id, amount, method, status,
                   justification, receipt_url, requested_by, approved_by,
                   requested_at, approved_at, updated_at"
    )
    .bind(purchase_id).bind(&input.cost_type).bind(input.supplier_id)
    .bind(input.amount).bind(&input.method).bind(&input.justification)
    .bind(&input.receipt_url).bind(user_id)
    .fetch_one(pool).await
    .map_err(|e| AppError::internal(e.to_string()))
}

pub async fn approve_payment(
    pool: &sqlx::PgPool, payment_id: i32, user_id: i32,
) -> Result<PurchasePayment, AppError> {
    sqlx::query_as::<_, PurchasePayment>(
        "UPDATE purchase_payments SET status = 'Pago', approved_by = $1, approved_at = NOW(), updated_at = NOW()
         WHERE id = $2
         RETURNING id, purchase_id, cost_type, supplier_id, amount, method, status,
                   justification, receipt_url, requested_by, approved_by,
                   requested_at, approved_at, updated_at"
    )
    .bind(user_id).bind(payment_id)
    .fetch_one(pool).await
    .map_err(|e| AppError::internal(e.to_string()))
}

pub async fn add_issue(
    pool: &sqlx::PgPool, purchase_id: i32, input: &PurchaseIssueInput, user_id: i32,
) -> Result<PurchaseIssue, AppError> {
    sqlx::query_as::<_, PurchaseIssue>(
        "INSERT INTO purchase_issues (purchase_id, issue_type, description, attachments,
                supplier_id, solution, occurrence_date, resolution_deadline, priority,
                status, opened_by)
         VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,'Aberta',$10)
         RETURNING id, purchase_id, issue_type, description, attachments, supplier_id,
                   solution, occurrence_date, resolution_deadline, priority, status,
                   opened_by, resolved_by, created_at, updated_at, resolved_at"
    )
    .bind(purchase_id).bind(&input.issue_type).bind(&input.description)
    .bind(&input.attachments).bind(input.supplier_id).bind(&input.solution)
    .bind(input.occurrence_date).bind(input.resolution_deadline)
    .bind(input.priority).bind(user_id)
    .fetch_one(pool).await
    .map_err(|e| AppError::internal(e.to_string()))
}

pub async fn update_issue(
    pool: &sqlx::PgPool, issue_id: i32, input: &PurchaseIssueUpdateInput,
) -> Result<PurchaseIssue, AppError> {
    sqlx::query_as::<_, PurchaseIssue>(
        "UPDATE purchase_issues SET status = $1, solution = $2, updated_at = NOW()
         WHERE id = $3
         RETURNING id, purchase_id, issue_type, description, attachments, supplier_id,
                   solution, occurrence_date, resolution_deadline, priority, status,
                   opened_by, resolved_by, created_at, updated_at, resolved_at"
    )
    .bind(&input.status).bind(&input.solution).bind(issue_id)
    .fetch_one(pool).await
    .map_err(|e| AppError::internal(e.to_string()))
}

pub async fn get_notifications(
    pool: &sqlx::PgPool, user_id: i32,
) -> Result<Vec<Notification>, AppError> {
    sqlx::query_as::<_, Notification>(
        "SELECT n.id, n.purchase_id, n.recipient_user_id, n.recipient_permission,
                n.notification_type, n.message, n.read_at, n.created_at
         FROM notifications n
         WHERE n.recipient_user_id = $1 OR n.recipient_user_id IS NULL
         ORDER BY n.created_at DESC"
    )
    .bind(user_id)
    .fetch_all(pool)
    .await
    .map_err(|e| AppError::internal(e.to_string()))
}

pub async fn mark_notification_read(pool: &sqlx::PgPool, id: i64) -> Result<(), AppError> {
    sqlx::query("UPDATE notifications SET read_at = NOW() WHERE id = $1")
        .bind(id).execute(pool).await
        .map_err(|e| AppError::internal(e.to_string()))?;
    Ok(())
}

// --- SQL constants ---

const SELECT_PURCHASE_ALL: &str = "
    SELECT po.id, po.sale_id, po.general_number, po.status, po.status_updated_at::timestamp AS status_updated_at,
           po.buyer_id, po.material_supplier_id, po.engraving_supplier_id,
           po.is_sample, po.sample_has_engraving, po.has_engraving,
           po.corel_required, po.corel_requested_at::timestamp AS corel_requested_at, po.corel_requested_by,
           po.corel_attached_at::timestamp AS corel_attached_at, po.corel_attached_by,
           po.material_unit_cost::float8 AS material_unit_cost,
           po.material_total_cost::float8 AS material_total_cost,
           po.engraving_cost::float8 AS engraving_cost,
           po.freight_cost::float8 AS freight_cost,
           po.other_cost::float8 AS other_cost,
           po.buyer_discount::float8 AS buyer_discount,
           po.negotiation_contact, po.negotiation_notes,
           po.material_deadline::timestamp AS material_deadline, po.engraving_deadline::timestamp AS engraving_deadline, po.payment_method,
           po.requires_advance_payment, po.material_accepted, po.material_accepted_at::timestamp AS material_accepted_at,
           po.engraving_accepted, po.engraving_accepted_at::timestamp AS engraving_accepted_at,
           po.first_piece_required, po.first_piece_status, po.first_piece_url,
           po.commercial_notes, po.purchase_notes, po.production_released_at::timestamp AS production_released_at,
           po.created_at, po.updated_at,
           jsonb_build_object(
                'id', sl.id,
                'payment_method', sl.payment_method,
                'delivery_date', sl.delivery_date,
                'departure_date', sl.departure_date,
                'arrival_date', sl.arrival_date,
                'total_value', sl.total_value,
                'status', sl.status,
                'customer', CASE WHEN c.id IS NULL THEN NULL ELSE jsonb_build_object(
                    'id', c.id,
                    'name', c.name,
                    'document', COALESCE(NULLIF(c.cnpj, ''), NULLIF(c.cpf, ''))
                ) END,
                'seller', CASE WHEN u.id IS NULL THEN NULL ELSE jsonb_build_object(
                    'id', u.id,
                    'username', u.username,
                    'full_name', u.full_name
                ) END,
                'items', COALESCE(items.rows, '[]'::jsonb)
           ) AS sale
    FROM purchase_orders po
    JOIN sales sl ON sl.id = po.sale_id
    LEFT JOIN customers c ON c.id = sl.customer_id
    LEFT JOIN users u ON u.id = sl.seller_id
    LEFT JOIN LATERAL (
        SELECT jsonb_agg(jsonb_build_object(
            'id', si.id,
            'sale_id', si.sale_id,
            'product_id', si.product_id,
            'quantity', si.quantity,
            'unit_price', si.unit_price,
            'total_price', si.total_price,
            'engravings', COALESCE(si.engravings, '[]'::jsonb),
            'personalization_type', COALESCE(si.personalization_type, ''),
            'product', jsonb_build_object(
                'id', p.id,
                'product_name', p.product_name,
                'internal_code', p.internal_code,
                'supplier_id', p.supplier_id,
                'supplier_stock', COALESCE(p.supplier_stock, 0)
            )
        ) ORDER BY si.id) AS rows
        FROM sale_items si
        LEFT JOIN products p ON p.id = si.product_id
        WHERE si.sale_id = sl.id
    ) items ON TRUE";

const UPDATE_PURCHASE_SQL: &str = "
    UPDATE purchase_orders SET
        buyer_id=$1, material_supplier_id=$2, engraving_supplier_id=$3,
        is_sample=$4, sample_has_engraving=$5, has_engraving=$6,
        material_unit_cost=$7, material_total_cost=$8, engraving_cost=$9,
        freight_cost=$10, other_cost=$11, buyer_discount=$12,
        negotiation_contact=$13, negotiation_notes=$14,
        material_deadline=$15, engraving_deadline=$16, payment_method=$17,
        requires_advance_payment=$18, first_piece_required=$19,
        commercial_notes=$20, purchase_notes=$21, updated_at=NOW()
    WHERE id=$22";

// --- Sub-entity queries ---

async fn get_attachments(pool: &sqlx::PgPool, id: i32) -> Result<Vec<PurchaseAttachment>, AppError> {
    sqlx::query_as::<_, PurchaseAttachment>(
        "SELECT id, purchase_id, category, file_name, url, uploaded_by, created_at
         FROM purchase_attachments WHERE purchase_id = $1 ORDER BY id"
    ).bind(id).fetch_all(pool).await.map_err(|e| AppError::internal(e.to_string()))
}

async fn get_emails(pool: &sqlx::PgPool, id: i32) -> Result<Vec<PurchaseEmail>, AppError> {
    sqlx::query_as::<_, PurchaseEmail>(
        "SELECT id, purchase_id, kind, recipient, subject, body, observation,
                attachments, sent_by, sent_at
         FROM purchase_emails WHERE purchase_id = $1 ORDER BY id"
    ).bind(id).fetch_all(pool).await.map_err(|e| AppError::internal(e.to_string()))
}

async fn get_payments(pool: &sqlx::PgPool, id: i32) -> Result<Vec<PurchasePayment>, AppError> {
    sqlx::query_as::<_, PurchasePayment>(
        "SELECT id, purchase_id, cost_type, supplier_id, amount, method, status,
                justification, receipt_url, requested_by, approved_by,
                requested_at, approved_at, updated_at
         FROM purchase_payments WHERE purchase_id = $1 ORDER BY id"
    ).bind(id).fetch_all(pool).await.map_err(|e| AppError::internal(e.to_string()))
}

async fn get_issues(pool: &sqlx::PgPool, id: i32) -> Result<Vec<PurchaseIssue>, AppError> {
    sqlx::query_as::<_, PurchaseIssue>(
        "SELECT id, purchase_id, issue_type, description, attachments, supplier_id,
                solution, occurrence_date, resolution_deadline, priority, status,
                opened_by, resolved_by, created_at, updated_at, resolved_at
         FROM purchase_issues WHERE purchase_id = $1 ORDER BY id"
    ).bind(id).fetch_all(pool).await.map_err(|e| AppError::internal(e.to_string()))
}

async fn get_history(pool: &sqlx::PgPool, id: i32) -> Result<Vec<PurchaseHistory>, AppError> {
    sqlx::query_as::<_, PurchaseHistory>(
        "SELECT id, purchase_id, action, from_status, to_status, details, user_id, user_name, created_at
         FROM purchase_history WHERE purchase_id = $1 ORDER BY id DESC"
    ).bind(id).fetch_all(pool).await.map_err(|e| AppError::internal(e.to_string()))
}
