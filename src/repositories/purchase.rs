use serde::Serialize;
use serde_json::Value;

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
