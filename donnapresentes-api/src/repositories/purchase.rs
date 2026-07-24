use serde::Serialize;

use crate::error::AppError;
use crate::models::{
    Notification, PurchaseActionInput, PurchaseAttachment, PurchaseEmail, PurchaseHistory,
    PurchaseIssue, PurchaseIssueInput, PurchaseIssueUpdateInput, PurchaseOrder,
    PurchasePayment, PurchasePaymentInput, PurchaseUpdateInput,
};

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

pub async fn release_sale_to_purchases(pool: &sqlx::PgPool, sale_id: i32) -> Result<PurchaseOrder, AppError> {
    let mut tx = pool.begin().await.map_err(|e| AppError::internal(e.to_string()))?;

    let sale = sqlx::query_as::<_, (i32, String)>(
        "SELECT s.id, COALESCE(c.name, '') FROM sales s
         LEFT JOIN customers c ON c.id = s.customer_id
         WHERE s.id = $1",
    )
    .bind(sale_id)
    .fetch_optional(&mut *tx)
    .await
    .map_err(|e| AppError::internal(e.to_string()))?
    .ok_or_else(|| AppError::not_found("Venda"))?;

    let general_number = format!("{:06}-{}", sale.0, sale.1.chars().take(20).collect::<String>().replace(' ', "_"));

    let po = sqlx::query_as::<_, PurchaseOrder>(
        "INSERT INTO purchase_orders (sale_id, general_number, status)
         VALUES ($1, $2, 'Pendente de Compra')
         ON CONFLICT (sale_id) DO UPDATE SET updated_at = NOW()
         RETURNING id, sale_id, general_number, status, status_updated_at,
                   buyer_id, material_supplier_id, engraving_supplier_id,
                   is_sample, sample_has_engraving, has_engraving,
                   corel_required, corel_requested_at, corel_requested_by,
                   corel_attached_at, corel_attached_by,
                   material_unit_cost, material_total_cost, engraving_cost,
                   freight_cost, other_cost, buyer_discount,
                   negotiation_contact, negotiation_notes,
                   material_deadline, engraving_deadline, payment_method,
                   requires_advance_payment, material_accepted, material_accepted_at,
                   engraving_accepted, engraving_accepted_at,
                   first_piece_required, first_piece_status, first_piece_url,
                   commercial_notes, purchase_notes, production_released_at,
                   created_at, updated_at",
    )
    .bind(sale.0)
    .bind(&general_number)
    .fetch_one(&mut *tx)
    .await
    .map_err(|e| AppError::internal(e.to_string()))?;

    tx.commit().await.map_err(|e| AppError::internal(e.to_string()))?;
    Ok(po)
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
        "send_material_email" => ("E-mail Material Enviado", true),
        "send_engraving_email" => ("E-mail Gravação Enviado", true),
        "release_to_production" => ("Liberado para Produção", true),
        "start_review" => ("Em Conferência de Compras", true),
        "accept_material" => (current.status.as_str(), true),
        "accept_engraving" => (current.status.as_str(), true),
        _ => return Err(AppError::bad_request(format!("Ação desconhecida: {action}"))),
    };

    if new_status != current.status {
        sqlx::query("UPDATE purchase_orders SET status = $1, status_updated_at = NOW() WHERE id = $2")
            .bind(new_status).bind(purchase_id)
            .execute(&mut *tx).await
            .map_err(|e| AppError::internal(e.to_string()))?;
    }

    if action == "release_to_production" {
        sqlx::query("UPDATE purchase_orders SET production_released_at = NOW() WHERE id = $1")
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
        sqlx::query(
            "INSERT INTO purchase_history (purchase_id, action, from_status, to_status, user_id, user_name)
             VALUES ($1, $2, $3, $4, $5, 'system')",
        )
        .bind(purchase_id).bind(action)
        .bind(&current.status).bind(new_status).bind(user_id)
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
    SELECT po.id, po.sale_id, po.general_number, po.status, po.status_updated_at,
           po.buyer_id, po.material_supplier_id, po.engraving_supplier_id,
           po.is_sample, po.sample_has_engraving, po.has_engraving,
           po.corel_required, po.corel_requested_at, po.corel_requested_by,
           po.corel_attached_at, po.corel_attached_by,
           po.material_unit_cost, po.material_total_cost, po.engraving_cost,
           po.freight_cost, po.other_cost, po.buyer_discount,
           po.negotiation_contact, po.negotiation_notes,
           po.material_deadline, po.engraving_deadline, po.payment_method,
           po.requires_advance_payment, po.material_accepted, po.material_accepted_at,
           po.engraving_accepted, po.engraving_accepted_at,
           po.first_piece_required, po.first_piece_status, po.first_piece_url,
           po.commercial_notes, po.purchase_notes, po.production_released_at,
           po.created_at, po.updated_at
    FROM purchase_orders po";

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
