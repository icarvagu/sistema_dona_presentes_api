package repositories

import (
	"database/sql"
	"encoding/json"
	"fmt"

	"donapresentes/models"
)

type PurchaseRepository struct {
	db       *sql.DB
	saleRepo *SaleRepository
}

func NewPurchaseRepository(db *sql.DB) *PurchaseRepository {
	return &PurchaseRepository{db: db, saleRepo: NewSaleRepository(db)}
}

type scanner interface {
	Scan(dest ...interface{}) error
}

const purchaseColumns = `
	id, sale_id, general_number, status, status_updated_at, buyer_id, material_supplier_id,
	engraving_supplier_id, is_sample, sample_has_engraving, has_engraving,
	corel_required, corel_requested_at, corel_requested_by, corel_attached_at,
	corel_attached_by, material_unit_cost, material_total_cost, engraving_cost,
	freight_cost, other_cost, buyer_discount, negotiation_contact, negotiation_notes,
	material_deadline, engraving_deadline, payment_method,
	requires_advance_payment, material_accepted, material_accepted_at,
	engraving_accepted, engraving_accepted_at, first_piece_required,
	first_piece_status, first_piece_url, commercial_notes, purchase_notes,
	production_released_at, created_at, updated_at`

func scanPurchase(row scanner) (*models.PurchaseOrder, error) {
	var p models.PurchaseOrder
	err := row.Scan(
		&p.ID, &p.SaleID, &p.GeneralNumber, &p.Status, &p.StatusUpdatedAt, &p.BuyerID,
		&p.MaterialSupplierID, &p.EngravingSupplierID, &p.IsSample,
		&p.SampleHasEngraving, &p.HasEngraving, &p.CorelRequired,
		&p.CorelRequestedAt, &p.CorelRequestedBy, &p.CorelAttachedAt,
		&p.CorelAttachedBy, &p.MaterialUnitCost, &p.MaterialTotalCost,
		&p.EngravingCost, &p.FreightCost, &p.OtherCost,
		&p.BuyerDiscount, &p.NegotiationContact, &p.NegotiationNotes, &p.MaterialDeadline,
		&p.EngravingDeadline, &p.PaymentMethod, &p.RequiresAdvancePayment,
		&p.MaterialAccepted, &p.MaterialAcceptedAt, &p.EngravingAccepted,
		&p.EngravingAcceptedAt, &p.FirstPieceRequired, &p.FirstPieceStatus,
		&p.FirstPieceURL, &p.CommercialNotes, &p.PurchaseNotes,
		&p.ProductionReleasedAt, &p.CreatedAt, &p.UpdatedAt,
	)
	return &p, err
}

func (r *PurchaseRepository) List(isSample *bool) ([]models.PurchaseOrder, error) {
	query := `SELECT ` + purchaseColumns + ` FROM purchase_orders`
	args := []interface{}{}
	if isSample != nil {
		query += ` WHERE is_sample=$1`
		args = append(args, *isSample)
	}
	query += ` ORDER BY updated_at DESC`
	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := []models.PurchaseOrder{}
	for rows.Next() {
		p, err := scanPurchase(rows)
		if err != nil {
			return nil, err
		}
		if err := r.loadRelations(p, false); err != nil {
			return nil, err
		}
		result = append(result, *p)
	}
	return result, rows.Err()
}

func (r *PurchaseRepository) GetByID(id int) (*models.PurchaseOrder, error) {
	p, err := scanPurchase(r.db.QueryRow(`SELECT `+purchaseColumns+` FROM purchase_orders WHERE id=$1`, id))
	if err != nil {
		return nil, err
	}
	if err := r.loadRelations(p, true); err != nil {
		return nil, err
	}
	return p, nil
}

func (r *PurchaseRepository) GetBySaleID(saleID int) (*models.PurchaseOrder, error) {
	p, err := scanPurchase(r.db.QueryRow(`SELECT `+purchaseColumns+` FROM purchase_orders WHERE sale_id=$1`, saleID))
	if err != nil {
		return nil, err
	}
	if err := r.loadRelations(p, true); err != nil {
		return nil, err
	}
	return p, nil
}

func (r *PurchaseRepository) ReleaseSale(saleID, userID int, isSample, sampleHasEngraving bool) (*models.PurchaseOrder, error) {
	tx, err := r.db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	var purchaseID int
	created := false
	if _, err = tx.Exec(`SELECT pg_advisory_xact_lock($1)`, saleID); err != nil {
		return nil, err
	}
	var saleExists int
	if err = tx.QueryRow(`SELECT 1 FROM sales WHERE id=$1 FOR UPDATE`, saleID).Scan(&saleExists); err != nil {
		return nil, err
	}
	workflowRepo := &SalesWorkflowRepository{db: r.db}
	if err = workflowRepo.validateChecklist(tx, saleID, true); err != nil {
		return nil, err
	}
	err = tx.QueryRow(`SELECT id FROM purchase_orders WHERE sale_id=$1 FOR UPDATE`, saleID).Scan(&purchaseID)
	if err == sql.ErrNoRows {
		err = tx.QueryRow(`
			INSERT INTO purchase_orders (
				sale_id, general_number, status, material_supplier_id, is_sample,
				sample_has_engraving, has_engraving, corel_required, material_unit_cost,
				material_total_cost, payment_method, commercial_notes
			)
			SELECT s.id, s.id::text, $2, MIN(NULLIF(p.supplier_id,0)), $3,
			       $4, CASE WHEN $3 THEN $4 ELSE COALESCE(BOOL_OR(COALESCE(si.engravings,'[]'::jsonb) <> '[]'::jsonb OR COALESCE(si.personalization_type,'') <> ''),FALSE) END,
			       CASE WHEN $3 THEN $4 ELSE COALESCE(BOOL_OR(COALESCE(si.engravings,'[]'::jsonb) <> '[]'::jsonb OR COALESCE(si.personalization_type,'') <> ''),FALSE) END,
			       COALESCE(MIN(p.cost_price),0), COALESCE(SUM(si.quantity*p.cost_price),0),
			       COALESCE(s.payment_method,''), COALESCE(s.internal_notes,'')
			FROM sales s
			LEFT JOIN sale_items si ON si.sale_id=s.id
			LEFT JOIN products p ON p.id=si.product_id
			WHERE s.id=$1
			GROUP BY s.id
			RETURNING id`, saleID, models.PurchasePending, isSample, sampleHasEngraving).Scan(&purchaseID)
		created = true
	}
	if err != nil {
		return nil, err
	}
	if !created {
		if _, err = tx.Exec(`UPDATE purchase_orders SET is_sample=$1,sample_has_engraving=$2,general_number=sale_id::text,
			has_engraving=CASE WHEN $1 THEN $2 ELSE EXISTS(SELECT 1 FROM sale_items si WHERE si.sale_id=purchase_orders.sale_id AND (COALESCE(si.engravings,'[]'::jsonb)<>'[]'::jsonb OR COALESCE(si.personalization_type,'')<>'')) END,
			corel_required=CASE WHEN $1 THEN $2 ELSE EXISTS(SELECT 1 FROM sale_items si WHERE si.sale_id=purchase_orders.sale_id AND (COALESCE(si.engravings,'[]'::jsonb)<>'[]'::jsonb OR COALESCE(si.personalization_type,'')<>'')) END,updated_at=NOW() WHERE id=$3`, isSample, sampleHasEngraving, purchaseID); err != nil {
			return nil, err
		}
	}
	if _, err = tx.Exec(`UPDATE sales SET status='Liberado para Compras', released_to_purchases_at=COALESCE(released_to_purchases_at, NOW()), updated_at=NOW() WHERE id=$1`, saleID); err != nil {
		return nil, err
	}
	if created {
		if _, err = tx.Exec(`INSERT INTO purchase_history (purchase_id,action,to_status,details,user_id) VALUES ($1,'Pedido liberado pelo Comercial',$2,'Pedido enviado para Compras',$3)`, purchaseID, models.PurchasePending, userID); err != nil {
			return nil, err
		}
		if _, err = tx.Exec(`INSERT INTO notifications (purchase_id,recipient_permission,notification_type,message) VALUES ($1,'compras','pedido_compra',$2)`, purchaseID, fmt.Sprintf("Pedido %d pendente de compra", saleID)); err != nil {
			return nil, err
		}
		var internalStockProducts string
		if err = tx.QueryRow(`SELECT COALESCE(string_agg(p.product_name || ' (' || p.supplier_stock || ' disponível)', ', ' ORDER BY p.product_name),'')
			FROM sale_items si JOIN products p ON p.id=si.product_id
			WHERE si.sale_id=$1 AND p.supplier_stock > 0`, saleID).Scan(&internalStockProducts); err != nil {
			return nil, err
		}
		if internalStockProducts != "" {
			if _, err = tx.Exec(`INSERT INTO notifications (purchase_id,recipient_permission,notification_type,message) VALUES ($1,'compras','estoque_empresa',$2)`, purchaseID, "Atenção: há produto no estoque da empresa: "+internalStockProducts); err != nil {
				return nil, err
			}
		}
	}
	if err = tx.Commit(); err != nil {
		return nil, err
	}
	return r.GetByID(purchaseID)
}

func (r *PurchaseRepository) Update(id int, input *models.PurchaseUpdateInput, userID int) error {
	_, err := r.db.Exec(`UPDATE purchase_orders SET
		buyer_id=$1, material_supplier_id=$2, engraving_supplier_id=$3,
		is_sample=$4, sample_has_engraving=$5, has_engraving=$6,
		corel_required=$6, material_unit_cost=$7, material_total_cost=$8,
		engraving_cost=$9, freight_cost=$10, other_cost=$11,
		buyer_discount=$12, negotiation_contact=$13, negotiation_notes=$14,
		material_deadline=$15, engraving_deadline=$16, payment_method=$17,
		requires_advance_payment=$18, first_piece_required=$19,
		commercial_notes=$20, purchase_notes=$21, updated_at=NOW()
		WHERE id=$22`, input.BuyerID, input.MaterialSupplierID,
		input.EngravingSupplierID, input.IsSample, input.SampleHasEngraving,
		input.HasEngraving, input.MaterialUnitCost, input.MaterialTotalCost,
		input.EngravingCost, input.FreightCost, input.OtherCost,
		input.BuyerDiscount, input.NegotiationContact, input.NegotiationNotes,
		input.MaterialDeadline, input.EngravingDeadline, input.PaymentMethod,
		input.RequiresAdvancePayment, input.FirstPieceRequired,
		input.CommercialNotes, input.PurchaseNotes, id)
	if err == nil {
		err = r.AddHistory(id, "Purchase data updated", "", "", "Values, suppliers and deadlines checked", userID)
	}
	return err
}

func (r *PurchaseRepository) SetStatus(id int, status string, userID int, action, details string) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	var oldStatus string
	if err = tx.QueryRow(`SELECT status FROM purchase_orders WHERE id=$1 FOR UPDATE`, id).Scan(&oldStatus); err != nil {
		return err
	}
	productionReleased := interface{}(nil)
	if status == models.PurchaseReleased {
		productionReleased = "now"
	}
	if productionReleased == nil {
		_, err = tx.Exec(`UPDATE purchase_orders SET status=$1,status_updated_at=NOW(),updated_at=NOW() WHERE id=$2`, status, id)
	} else {
		_, err = tx.Exec(`UPDATE purchase_orders SET status=$1,status_updated_at=NOW(),production_released_at=NOW(),updated_at=NOW() WHERE id=$2`, status, id)
	}
	if err != nil {
		return err
	}
	if _, err = tx.Exec(`INSERT INTO purchase_history (purchase_id,action,from_status,to_status,details,user_id) VALUES ($1,$2,$3,$4,$5,$6)`, id, action, oldStatus, status, details, userID); err != nil {
		return err
	}
	if status == models.PurchaseReleased {
		if _, err = tx.Exec(`UPDATE sales SET status=$1,updated_at=NOW() WHERE id=(SELECT sale_id FROM purchase_orders WHERE id=$2)`, status, id); err != nil {
			return err
		}
		productionRepo := NewProductionRepository(r.db)
		productionID, ensureErr := productionRepo.EnsureForPurchaseTx(tx, id, userID)
		if ensureErr != nil {
			return ensureErr
		}
		if _, err = tx.Exec(`INSERT INTO notifications (purchase_id,production_order_id,recipient_permission,notification_type,message)
			VALUES ($1,$2,'producao','ordem_producao',$3)`, id, productionID, fmt.Sprintf("Pedido %d liberado para Produção", id)); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (r *PurchaseRepository) MarkCorelRequested(id, userID int) error {
	_, err := r.db.Exec(`UPDATE purchase_orders SET corel_requested_at=NOW(),corel_requested_by=$1,updated_at=NOW() WHERE id=$2`, userID, id)
	return err
}

func (r *PurchaseRepository) MarkCorelAttached(id, userID int) error {
	_, err := r.db.Exec(`UPDATE purchase_orders SET corel_attached_at=NOW(),corel_attached_by=$1,updated_at=NOW() WHERE id=$2`, userID, id)
	return err
}

func (r *PurchaseRepository) SetAcceptance(id int, engraving bool) error {
	if engraving {
		_, err := r.db.Exec(`UPDATE purchase_orders SET engraving_accepted=TRUE,engraving_accepted_at=NOW(),updated_at=NOW() WHERE id=$1`, id)
		return err
	}
	_, err := r.db.Exec(`UPDATE purchase_orders SET material_accepted=TRUE,material_accepted_at=NOW(),updated_at=NOW() WHERE id=$1`, id)
	return err
}

func (r *PurchaseRepository) SetFirstPiece(id int, status, url string) error {
	_, err := r.db.Exec(`UPDATE purchase_orders SET first_piece_status=$1,first_piece_url=CASE WHEN $2='' THEN first_piece_url ELSE $2 END,updated_at=NOW() WHERE id=$3`, status, url, id)
	return err
}

func (r *PurchaseRepository) AddAttachment(purchaseID int, category, fileName, url string, userID int) error {
	_, err := r.db.Exec(`INSERT INTO purchase_attachments (purchase_id,category,file_name,url,uploaded_by) VALUES ($1,$2,$3,$4,$5)`, purchaseID, category, fileName, url, userID)
	return err
}

func (r *PurchaseRepository) AddEmail(email *models.PurchaseEmail) error {
	attachments := email.Attachments
	if len(attachments) == 0 {
		attachments = json.RawMessage(`[]`)
	}
	return r.db.QueryRow(`INSERT INTO purchase_emails (purchase_id,kind,recipient,subject,body,observation,attachments,sent_by) VALUES ($1,$2,$3,$4,$5,$6,$7,$8) RETURNING id,sent_at`,
		email.PurchaseID, email.Kind, email.Recipient, email.Subject, email.Body,
		email.Observation, attachments, email.SentBy).Scan(&email.ID, &email.SentAt)
}

func (r *PurchaseRepository) AddPayment(purchaseID, userID int, input *models.PurchasePaymentInput) error {
	status := "Pagamento Registrado"
	if input.Method == "PIX" {
		status = "PIX Solicitado"
	}
	if input.Method == "Cheque" {
		status = "Cheque Solicitado"
	}
	_, err := r.db.Exec(`INSERT INTO purchase_payments (purchase_id,cost_type,supplier_id,amount,method,status,justification,receipt_url,requested_by) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)`,
		purchaseID, input.CostType, input.SupplierID, input.Amount, input.Method,
		status, input.Justification, input.ReceiptURL, userID)
	return err
}

func (r *PurchaseRepository) ApprovePayment(paymentID, userID int, receiptURL string) (int, error) {
	var purchaseID int
	err := r.db.QueryRow(`UPDATE purchase_payments SET status='Pagamento Registrado',receipt_url=CASE WHEN $1='' THEN receipt_url ELSE $1 END,approved_by=$2,approved_at=NOW(),updated_at=NOW() WHERE id=$3 RETURNING purchase_id`, receiptURL, userID, paymentID).Scan(&purchaseID)
	return purchaseID, err
}

func (r *PurchaseRepository) AddIssue(purchaseID, userID int, input *models.PurchaseIssueInput) error {
	attachments, _ := json.Marshal(input.Attachments)
	_, err := r.db.Exec(`INSERT INTO purchase_issues (purchase_id,issue_type,description,attachments,supplier_id,solution,occurrence_date,resolution_deadline,priority,opened_by) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)`,
		purchaseID, input.IssueType, input.Description, attachments, input.SupplierID,
		input.Solution, input.OccurrenceDate, input.ResolutionDeadline, input.Priority, userID)
	return err
}

func (r *PurchaseRepository) UpdateIssue(issueID, userID int, input *models.PurchaseIssueUpdateInput) (int, error) {
	var purchaseID int
	if input.Status == "Resolvida" {
		err := r.db.QueryRow(`UPDATE purchase_issues SET solution=$1,resolution_deadline=COALESCE($2,resolution_deadline),status=$3,resolved_by=$4,resolved_at=NOW(),updated_at=NOW() WHERE id=$5 RETURNING purchase_id`, input.Solution, input.ResolutionDeadline, input.Status, userID, issueID).Scan(&purchaseID)
		return purchaseID, err
	}
	err := r.db.QueryRow(`UPDATE purchase_issues SET solution=$1,resolution_deadline=COALESCE($2,resolution_deadline),status=$3,updated_at=NOW() WHERE id=$4 RETURNING purchase_id`, input.Solution, input.ResolutionDeadline, input.Status, issueID).Scan(&purchaseID)
	return purchaseID, err
}

func (r *PurchaseRepository) AddHistory(purchaseID int, action, fromStatus, toStatus, details string, userID int) error {
	_, err := r.db.Exec(`INSERT INTO purchase_history (purchase_id,action,from_status,to_status,details,user_id) VALUES ($1,$2,$3,$4,$5,$6)`, purchaseID, action, fromStatus, toStatus, details, userID)
	return err
}

func (r *PurchaseRepository) Notify(purchaseID int, permission, notificationType, message string, userID *int) error {
	_, err := r.db.Exec(`INSERT INTO notifications (purchase_id,recipient_user_id,recipient_permission,notification_type,message) VALUES ($1,$2,$3,$4,$5)`, purchaseID, userID, permission, notificationType, message)
	return err
}

func (r *PurchaseRepository) ListNotifications(userID int, permissions []string) ([]models.Notification, error) {
	permJSON, _ := json.Marshal(permissions)
	rows, err := r.db.Query(`SELECT id,purchase_id,recipient_user_id,recipient_permission,notification_type,message,read_at,created_at FROM notifications WHERE read_at IS NULL AND (recipient_user_id=$1 OR recipient_permission IN (SELECT jsonb_array_elements_text($2::jsonb))) ORDER BY created_at DESC`, userID, string(permJSON))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := []models.Notification{}
	for rows.Next() {
		var n models.Notification
		if err := rows.Scan(&n.ID, &n.PurchaseID, &n.RecipientUserID, &n.RecipientPermission, &n.NotificationType, &n.Message, &n.ReadAt, &n.CreatedAt); err != nil {
			return nil, err
		}
		result = append(result, n)
	}
	return result, rows.Err()
}

func (r *PurchaseRepository) MarkNotificationRead(id, userID int) error {
	_, err := r.db.Exec(`UPDATE notifications SET read_at=NOW() WHERE id=$1 AND (recipient_user_id IS NULL OR recipient_user_id=$2)`, id, userID)
	return err
}

func (r *PurchaseRepository) FinancialSummary() ([]models.PurchaseFinancialSummary, error) {
	rows, err := r.db.Query(`
		WITH base_costs AS (
			SELECT po.id AS purchase_id,po.general_number,po.sale_id,po.updated_at,
			       cost.cost_type,cost.amount,cost.supplier_id,po.payment_method
			FROM purchase_orders po
			CROSS JOIN LATERAL (VALUES
				('material'::text,po.material_total_cost,po.material_supplier_id),
				('gravacao'::text,po.engraving_cost,po.engraving_supplier_id),
				('frete'::text,po.freight_cost,NULL::integer),
				('outros'::text,po.other_cost,NULL::integer)
			) AS cost(cost_type,amount,supplier_id)
			WHERE cost.amount > 0
		), financial_rows AS (
			SELECT po.id AS purchase_id,po.general_number,po.sale_id,p.updated_at,p.cost_type,p.amount,
			       p.supplier_id,p.method,p.status,p.receipt_url
			FROM purchase_payments p
			JOIN purchase_orders po ON po.id=p.purchase_id
			UNION ALL
			SELECT b.purchase_id,b.general_number,b.sale_id,b.updated_at,b.cost_type,b.amount,
			       b.supplier_id,COALESCE(NULLIF(b.payment_method,''),'Não informado'),'Pendente de Pagamento',''
			FROM base_costs b
			WHERE NOT EXISTS (SELECT 1 FROM purchase_payments p WHERE p.purchase_id=b.purchase_id AND p.cost_type=b.cost_type)
		)
		SELECT f.purchase_id,f.general_number,COALESCE(c.name,''),f.cost_type,
		       COALESCE(sup.name,''),f.amount,f.method,f.status,f.receipt_url
		FROM financial_rows f
		JOIN sales s ON s.id=f.sale_id
		LEFT JOIN customers c ON c.id=s.customer_id
		LEFT JOIN suppliers sup ON sup.id=f.supplier_id
		ORDER BY f.updated_at DESC,f.purchase_id,f.cost_type`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := []models.PurchaseFinancialSummary{}
	for rows.Next() {
		var item models.PurchaseFinancialSummary
		if err := rows.Scan(&item.PurchaseID, &item.GeneralNumber, &item.CustomerName, &item.CostType, &item.SupplierName, &item.Amount, &item.Method, &item.Status, &item.ReceiptURL); err != nil {
			return nil, err
		}
		result = append(result, item)
	}
	return result, rows.Err()
}

func (r *PurchaseRepository) loadRelations(p *models.PurchaseOrder, full bool) error {
	p.Sale, _ = r.saleRepo.GetByID(p.SaleID)
	if p.MaterialSupplierID != nil {
		p.MaterialSupplier, _ = r.getSupplier(*p.MaterialSupplierID)
	}
	if p.EngravingSupplierID != nil {
		p.EngravingSupplier, _ = r.getSupplier(*p.EngravingSupplierID)
	}
	if !full {
		return nil
	}
	var err error
	if p.Attachments, err = r.listAttachments(p.ID); err != nil {
		return err
	}
	if p.Emails, err = r.listEmails(p.ID); err != nil {
		return err
	}
	if p.Payments, err = r.listPayments(p.ID); err != nil {
		return err
	}
	if p.Issues, err = r.listIssues(p.ID); err != nil {
		return err
	}
	if p.History, err = r.listHistory(p.ID); err != nil {
		return err
	}
	return nil
}

func (r *PurchaseRepository) getSupplier(id int) (*models.Supplier, error) {
	var s models.Supplier
	err := r.db.QueryRow(`SELECT id,name,cnpj,state_registration,contact_person,email,landline_phone,mobile_phone,responsible_email,commercial_address,website,created_at,updated_at FROM suppliers WHERE id=$1`, id).Scan(
		&s.ID, &s.Name, &s.CNPJ, &s.StateRegistration, &s.ContactPerson, &s.Email,
		&s.LandlinePhone, &s.MobilePhone, &s.ResponsibleEmail, &s.CommercialAddress,
		&s.Website, &s.CreatedAt, &s.UpdatedAt)
	return &s, err
}

func (r *PurchaseRepository) listAttachments(id int) ([]models.PurchaseAttachment, error) {
	rows, err := r.db.Query(`SELECT id,purchase_id,category,file_name,url,uploaded_by,created_at FROM purchase_attachments WHERE purchase_id=$1 ORDER BY created_at DESC`, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []models.PurchaseAttachment{}
	for rows.Next() {
		var x models.PurchaseAttachment
		if err := rows.Scan(&x.ID, &x.PurchaseID, &x.Category, &x.FileName, &x.URL, &x.UploadedBy, &x.CreatedAt); err != nil {
			return nil, err
		}
		items = append(items, x)
	}
	return items, rows.Err()
}

func (r *PurchaseRepository) listEmails(id int) ([]models.PurchaseEmail, error) {
	rows, err := r.db.Query(`SELECT id,purchase_id,kind,recipient,subject,body,observation,attachments,sent_by,sent_at FROM purchase_emails WHERE purchase_id=$1 ORDER BY sent_at DESC`, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []models.PurchaseEmail{}
	for rows.Next() {
		var x models.PurchaseEmail
		var raw []byte
		if err := rows.Scan(&x.ID, &x.PurchaseID, &x.Kind, &x.Recipient, &x.Subject, &x.Body, &x.Observation, &raw, &x.SentBy, &x.SentAt); err != nil {
			return nil, err
		}
		x.Attachments = raw
		items = append(items, x)
	}
	return items, rows.Err()
}

func (r *PurchaseRepository) listPayments(id int) ([]models.PurchasePayment, error) {
	rows, err := r.db.Query(`SELECT id,purchase_id,cost_type,supplier_id,amount,method,status,justification,receipt_url,requested_by,approved_by,requested_at,approved_at,updated_at FROM purchase_payments WHERE purchase_id=$1 ORDER BY requested_at DESC`, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []models.PurchasePayment{}
	for rows.Next() {
		var x models.PurchasePayment
		if err := rows.Scan(&x.ID, &x.PurchaseID, &x.CostType, &x.SupplierID, &x.Amount, &x.Method, &x.Status, &x.Justification, &x.ReceiptURL, &x.RequestedBy, &x.ApprovedBy, &x.RequestedAt, &x.ApprovedAt, &x.UpdatedAt); err != nil {
			return nil, err
		}
		items = append(items, x)
	}
	return items, rows.Err()
}

func (r *PurchaseRepository) listIssues(id int) ([]models.PurchaseIssue, error) {
	rows, err := r.db.Query(`SELECT id,purchase_id,issue_type,description,attachments,supplier_id,solution,occurrence_date,resolution_deadline,priority,status,opened_by,resolved_by,created_at,updated_at,resolved_at FROM purchase_issues WHERE purchase_id=$1 ORDER BY CASE WHEN status='Resolvida' THEN 1 ELSE 0 END,priority DESC,resolution_deadline ASC,created_at DESC`, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []models.PurchaseIssue{}
	for rows.Next() {
		var x models.PurchaseIssue
		var raw []byte
		if err := rows.Scan(&x.ID, &x.PurchaseID, &x.IssueType, &x.Description, &raw, &x.SupplierID, &x.Solution, &x.OccurrenceDate, &x.ResolutionDeadline, &x.Priority, &x.Status, &x.OpenedBy, &x.ResolvedBy, &x.CreatedAt, &x.UpdatedAt, &x.ResolvedAt); err != nil {
			return nil, err
		}
		x.Attachments = raw
		items = append(items, x)
	}
	return items, rows.Err()
}

func (r *PurchaseRepository) listHistory(id int) ([]models.PurchaseHistory, error) {
	rows, err := r.db.Query(`SELECT h.id,h.purchase_id,h.action,h.from_status,h.to_status,h.details,h.user_id,COALESCE(NULLIF(u.full_name,''),u.username,'Sistema'),h.created_at FROM purchase_history h LEFT JOIN users u ON u.id=h.user_id WHERE h.purchase_id=$1 ORDER BY h.created_at DESC`, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []models.PurchaseHistory{}
	for rows.Next() {
		var x models.PurchaseHistory
		if err := rows.Scan(&x.ID, &x.PurchaseID, &x.Action, &x.FromStatus, &x.ToStatus, &x.Details, &x.UserID, &x.UserName, &x.CreatedAt); err != nil {
			return nil, err
		}
		items = append(items, x)
	}
	return items, rows.Err()
}
