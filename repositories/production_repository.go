package repositories

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"donapresentes/models"
)

// ProductionRepository handles all database operations for production_orders and related tables
// (production_order_items, production_receipts, production_occurrences, production_engraving_events,
// production_volumes, production_fiscal_documents, production_shipments, production_supplies).
type ProductionRepository struct{ db *sql.DB }

// NewProductionRepository creates a new ProductionRepository with the given database connection.
func NewProductionRepository(db *sql.DB) *ProductionRepository { return &ProductionRepository{db: db} }

const productionSelect = `SELECT po.id,po.purchase_id,po.sale_id,po.status,po.has_engraving,po.first_piece_required,
 po.priority,po.owner_id,po.version,po.released_at,po.completed_at,po.created_at,po.updated_at,
	 po.conference_started_at,po.conference_completed_at,po.conference_user_id,p.general_number,COALESCE(c.name,''),s.seller_id,s.departure_date,p.material_deadline,p.engraving_deadline`
const productionFrom = ` FROM production_orders po JOIN purchase_orders p ON p.id=po.purchase_id
 JOIN sales s ON s.id=po.sale_id LEFT JOIN customers c ON c.id=s.customer_id`

func scanProduction(row scanner) (*models.ProductionOrder, error) {
	var o models.ProductionOrder
	err := row.Scan(&o.ID, &o.PurchaseID, &o.SaleID, &o.Status, &o.HasEngraving, &o.FirstPieceRequired, &o.Priority, &o.OwnerID, &o.Version, &o.ReleasedAt, &o.CompletedAt, &o.CreatedAt, &o.UpdatedAt, &o.ConferenceStartedAt, &o.ConferenceCompletedAt, &o.ConferenceUserID, &o.GeneralNumber, &o.CustomerName, &o.SellerID, &o.DepartureDate, &o.MaterialDeadline, &o.EngravingDeadline)
	return &o, err
}

// List returns production orders filtered by the given criteria from the database.
func (r *ProductionRepository) List(f models.ProductionFilters, a models.ProductionAccess) ([]models.ProductionOrder, error) {
	q := productionSelect + productionFrom + ` WHERE 1=1`
	args := []interface{}{}
	add := func(clause string, v interface{}) { args = append(args, v); q += fmt.Sprintf(clause, len(args)) }
	if f.Status != "" {
		add(` AND po.status=$%d`, f.Status)
	}
	if f.Search != "" {
		args = append(args, f.Search)
		n := len(args)
		q += fmt.Sprintf(` AND (p.general_number ILIKE '%%'||$%d||'%%' OR c.name ILIKE '%%'||$%d||'%%')`, n, n)
	}
	if f.OwnerID != nil {
		add(` AND po.owner_id=$%d`, *f.OwnerID)
	}
	if f.Priority != nil {
		add(` AND po.priority=$%d`, *f.Priority)
	}
	if !a.Admin && a.Sales && !(a.Production || a.Inspection || a.Purchases || a.Finance || a.Board || a.Logistics) {
		add(` AND s.seller_id=$%d`, a.UserID)
	}
	if !a.Admin && a.Driver && !(a.Production || a.Inspection || a.Purchases || a.Sales || a.Finance || a.Board || a.Logistics) {
		add(` AND EXISTS(SELECT 1 FROM production_shipments ps WHERE ps.production_order_id=po.id AND ps.driver_id=$%d)`, a.UserID)
	}
	q += ` ORDER BY po.priority DESC,s.departure_date NULLS LAST,po.updated_at DESC`
	rows, err := r.db.Query(q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []models.ProductionOrder{}
	for rows.Next() {
		o, e := scanProduction(rows)
		if e != nil {
			return nil, e
		}
		out = append(out, *o)
	}
	return out, rows.Err()
}

// Get returns a single production order by its primary key with all related sub-records loaded.
func (r *ProductionRepository) Get(id int64) (*models.ProductionOrder, error) {
	o, err := scanProduction(r.db.QueryRow(productionSelect+productionFrom+` WHERE po.id=$1`, id))
	if err != nil {
		return nil, err
	}
	queries := []struct {
		dst *json.RawMessage
		q   string
	}{
		{&o.Items, `SELECT COALESCE(jsonb_agg(jsonb_build_object('id',i.id,'sale_item_id',i.sale_item_id,'product_id',i.product_id,'product_name',p.product_name,'expected_quantity',i.expected_quantity,'received_quantity',i.received_quantity,'approved_quantity',i.approved_quantity,'rejected_quantity',i.rejected_quantity,'has_engraving',i.has_engraving) ORDER BY i.id),'[]') FROM production_order_items i JOIN products p ON p.id=i.product_id WHERE i.production_order_id=$1`},
		{&o.Receipts, `SELECT COALESCE(jsonb_agg(to_jsonb(x) ORDER BY x.received_at DESC),'[]') FROM (SELECT r.id,r.invoice_number,r.invoice_key,r.invoice_url,r.notes,r.received_at,r.received_by FROM production_receipts r WHERE r.production_order_id=$1) x`},
		{&o.Occurrences, `SELECT COALESCE(jsonb_agg(to_jsonb(x) ORDER BY x.created_at DESC),'[]') FROM (SELECT id,production_item_id,kind,severity,quantity,description,attachment_url,status,resolution,created_at,resolved_at FROM production_occurrences WHERE production_order_id=$1) x`},
		{&o.EngravingEvents, `SELECT COALESCE(jsonb_agg(to_jsonb(x) ORDER BY x.created_at DESC),'[]') FROM (SELECT id,event_type,status,quantity,carrier_id,tracking_code,file_url,notes,created_at FROM production_engraving_events WHERE production_order_id=$1) x`},
		{&o.Volumes, `SELECT COALESCE(jsonb_agg(to_jsonb(x) ORDER BY x.id),'[]') FROM (SELECT id,label,weight_kg,length_cm,width_cm,height_cm,created_at FROM production_volumes WHERE production_order_id=$1) x`},
		{&o.FiscalDocuments, `SELECT COALESCE(jsonb_agg(to_jsonb(x) ORDER BY x.created_at DESC),'[]') FROM (SELECT id,document_type,document_number,access_key,file_url,issued_at,created_at FROM production_fiscal_documents WHERE production_order_id=$1) x`},
		{&o.Shipment, `SELECT COALESCE((SELECT to_jsonb(x) FROM (SELECT method,carrier_id,driver_id,tracking_code,tracking_url,postal_service,proof_url,shipped_at,delivered_at FROM production_shipments WHERE production_order_id=$1) x),'{}')`},
		{&o.History, `SELECT COALESCE(jsonb_agg(to_jsonb(x) ORDER BY x.created_at DESC),'[]') FROM (SELECT id,action,from_status,to_status,details,user_id,created_at FROM production_history WHERE production_order_id=$1) x`},
	}
	for _, query := range queries {
		var raw []byte
		if err = r.db.QueryRow(query.q, id).Scan(&raw); err != nil {
			return nil, err
		}
		*query.dst = json.RawMessage(raw)
	}
	return o, nil
}

// DriverAssigned checks whether a given user is assigned as a driver for a production order.
func (r *ProductionRepository) DriverAssigned(orderID int64, userID int) bool {
	var ok bool
	_ = r.db.QueryRow(`SELECT EXISTS(SELECT 1 FROM production_shipments WHERE production_order_id=$1 AND driver_id=$2)`, orderID, userID).Scan(&ok)
	return ok
}

// Assign updates the owner and priority of a production order and records the change in history.
func (r *ProductionRepository) Assign(orderID int64, in models.ProductionAssignmentInput, userID int) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if in.OwnerID != nil {
		var validOwner bool
		if err = tx.QueryRow(`SELECT EXISTS(
			SELECT 1 FROM users
			WHERE id=$1 AND status='Ativo' AND (role='admin' OR 'producao'=ANY(permissions))
		)`, *in.OwnerID).Scan(&validOwner); err != nil {
			return err
		}
		if !validOwner {
			return errors.New("responsável deve ser um usuário ativo com permissão de Produção")
		}
	}
	var oldOwner *int
	var oldPriority int
	if err = tx.QueryRow(`SELECT owner_id,priority FROM production_orders WHERE id=$1 FOR UPDATE`, orderID).Scan(&oldOwner, &oldPriority); err != nil {
		return err
	}
	if _, err = tx.Exec(`UPDATE production_orders SET owner_id=$1,priority=$2,version=version+1,updated_at=NOW() WHERE id=$3`, in.OwnerID, in.Priority, orderID); err != nil {
		return err
	}
	_, err = tx.Exec(`INSERT INTO production_history(production_order_id,action,details,user_id) VALUES($1,'ATRIBUICAO',jsonb_build_object('old_owner',$2,'owner_id',$3,'old_priority',$4,'priority',$5),$6)`, orderID, oldOwner, in.OwnerID, oldPriority, in.Priority, userID)
	if err != nil {
		return err
	}
	return tx.Commit()
}

// EnsureForPurchaseTx ensures a production order exists for a given purchase within the provided transaction.
func (r *ProductionRepository) EnsureForPurchaseTx(tx *sql.Tx, purchaseID, userID int) (int64, error) {
	var id int64
	err := tx.QueryRow(`INSERT INTO production_orders(purchase_id,sale_id,has_engraving,first_piece_required)
	 SELECT id,sale_id,has_engraving,first_piece_required FROM purchase_orders WHERE id=$1
	 ON CONFLICT(purchase_id) DO UPDATE SET updated_at=production_orders.updated_at RETURNING id`, purchaseID).Scan(&id)
	if err != nil {
		return 0, err
	}
	_, err = tx.Exec(`INSERT INTO production_order_items(production_order_id,sale_item_id,product_id,expected_quantity,has_engraving)
	 SELECT $1,si.id,si.product_id,si.quantity,COALESCE(si.engravings,'[]'::jsonb)<>'[]'::jsonb FROM sale_items si
	 WHERE si.sale_id=(SELECT sale_id FROM production_orders WHERE id=$1) ON CONFLICT(sale_item_id) DO NOTHING`, id)
	if err != nil {
		return 0, err
	}
	_, _ = tx.Exec(`INSERT INTO production_history(production_order_id,action,to_status,details,user_id)
	 SELECT $1,'ORDEM_CRIADA','AGUARDANDO_RECEBIMENTO',jsonb_build_object('purchase_id',$2),$3
	 WHERE NOT EXISTS(SELECT 1 FROM production_history WHERE production_order_id=$1 AND action='ORDEM_CRIADA')`, id, purchaseID, userID)
	return id, nil
}

// AddReceipt records material receipt for a production order and updates the status accordingly.
func (r *ProductionRepository) AddReceipt(orderID int64, in models.ProductionReceiptInput, userID int) (*models.ProductionOrder, error) {
	tx, err := r.db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	var status string
	if err = tx.QueryRow(`SELECT status FROM production_orders WHERE id=$1 FOR UPDATE`, orderID).Scan(&status); err != nil {
		return nil, err
	}
	if status != models.ProductionAwaitingReceipt && status != models.ProductionPartialReceipt {
		return nil, errors.New("recebimento permitido apenas nas etapas de recebimento")
	}
	var receiptID int64
	err = tx.QueryRow(`INSERT INTO production_receipts(production_order_id,idempotency_key,invoice_number,invoice_key,invoice_url,notes,received_by)
	 VALUES($1,$2,$3,$4,$5,$6,$7) ON CONFLICT(production_order_id,idempotency_key) DO NOTHING RETURNING id`, orderID, in.IdempotencyKey, in.InvoiceNumber, in.InvoiceKey, in.InvoiceURL, in.Notes, userID).Scan(&receiptID)
	if err == sql.ErrNoRows {
		tx.Rollback()
		return r.Get(orderID)
	}
	if err != nil {
		return nil, err
	}
	for _, it := range in.Items {
		res, e := tx.Exec(`UPDATE production_order_items SET received_quantity=received_quantity+$1,updated_at=NOW() WHERE id=$2 AND production_order_id=$3 AND received_quantity+$1<=expected_quantity`, it.Quantity, it.ItemID, orderID)
		if e != nil {
			return nil, e
		}
		n, _ := res.RowsAffected()
		if n != 1 {
			return nil, errors.New("quantidade recebida excede a esperada")
		}
		if _, e = tx.Exec(`INSERT INTO production_receipt_items(receipt_id,production_item_id,quantity) VALUES($1,$2,$3)`, receiptID, it.ItemID, it.Quantity); e != nil {
			return nil, e
		}
	}
	var complete bool
	err = tx.QueryRow(`SELECT COALESCE(bool_and(received_quantity=expected_quantity),false) FROM production_order_items WHERE production_order_id=$1`, orderID).Scan(&complete)
	if err != nil {
		return nil, err
	}
	to := models.ProductionPartialReceipt
	if complete {
		to = models.ProductionInspection
	}
	_, err = tx.Exec(`UPDATE production_orders SET status=$1,version=version+1,updated_at=NOW() WHERE id=$2`, to, orderID)
	if err != nil {
		return nil, err
	}
	details, _ := json.Marshal(in)
	_, err = tx.Exec(`INSERT INTO production_history(production_order_id,action,from_status,to_status,details,user_id) VALUES($1,'RECEBIMENTO',$2,$3,$4,$5)`, orderID, status, to, details, userID)
	if err != nil {
		return nil, err
	}
	if _, err = tx.Exec(`INSERT INTO notifications(production_order_id,recipient_permission,notification_type,message) VALUES($1,'conferencia','recebimento_producao',$2)`, orderID, fmt.Sprintf("Ordem %d recebeu material e aguarda conferência", orderID)); err != nil {
		return nil, err
	}
	if err = tx.Commit(); err != nil {
		return nil, err
	}
	return r.Get(orderID)
}

// AddOccurrence records a quality occurrence for a production order and may block progress if severity is BLOQUEANTE.
func (r *ProductionRepository) AddOccurrence(orderID int64, in models.ProductionOccurrenceInput, userID int) (*models.ProductionOrder, error) {
	tx, err := r.db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	var old string
	if err = tx.QueryRow(`SELECT status FROM production_orders WHERE id=$1 FOR UPDATE`, orderID).Scan(&old); err != nil {
		return nil, err
	}
	if old == models.ProductionShipped || old == models.ProductionDelivered || old == models.ProductionCompleted {
		return nil, errors.New("não é possível abrir ocorrência após a expedição")
	}
	_, err = tx.Exec(`INSERT INTO production_occurrences(production_order_id,production_item_id,kind,severity,quantity,description,attachment_url,opened_by) VALUES($1,$2,$3,$4,$5,$6,$7,$8)`, orderID, in.ItemID, in.Kind, in.Severity, in.Quantity, in.Description, in.AttachmentURL, userID)
	if err != nil {
		return nil, err
	}
	if in.ItemID != nil && in.Quantity > 0 {
		var res sql.Result
		res, err = tx.Exec(`UPDATE production_order_items SET rejected_quantity=rejected_quantity+$1 WHERE id=$2 AND production_order_id=$3 AND rejected_quantity+$1<=received_quantity`, in.Quantity, *in.ItemID, orderID)
		if err != nil {
			return nil, err
		}
		n, _ := res.RowsAffected()
		if n != 1 {
			return nil, errors.New("item não pertence à ordem ou quantidade afetada excede a recebida")
		}
	} else if in.ItemID != nil {
		var exists bool
		if err = tx.QueryRow(`SELECT EXISTS(SELECT 1 FROM production_order_items WHERE id=$1 AND production_order_id=$2)`, *in.ItemID, orderID).Scan(&exists); err != nil {
			return nil, err
		}
		if !exists {
			return nil, errors.New("item não pertence à ordem")
		}
	}
	to := old
	if in.Severity == "BLOQUEANTE" {
		to = models.ProductionBlocked
		_, err = tx.Exec(`UPDATE production_orders SET status=$1,version=version+1,updated_at=NOW() WHERE id=$2`, to, orderID)
		if err != nil {
			return nil, err
		}
	}
	_, err = tx.Exec(`INSERT INTO production_history(production_order_id,action,from_status,to_status,details,user_id) VALUES($1,'OCORRENCIA',$2,$3,jsonb_build_object('severity',$4,'description',$5),$6)`, orderID, old, to, in.Severity, in.Description, userID)
	if err != nil {
		return nil, err
	}
	if _, err = tx.Exec(`INSERT INTO notifications(production_order_id,recipient_permission,notification_type,message) VALUES($1,'compras','ocorrencia_producao',$2)`, orderID, in.Description); err != nil {
		return nil, err
	}
	if err = tx.Commit(); err != nil {
		return nil, err
	}
	return r.Get(orderID)
}

// ResolveOccurrence marks a production occurrence as resolved in the database.
func (r *ProductionRepository) ResolveOccurrence(orderID, occurrenceID int64, resolution string, userID int) error {
	res, err := r.db.Exec(`UPDATE production_occurrences SET status='RESOLVIDA',resolution=$1,resolved_by=$2,resolved_at=NOW() WHERE id=$3 AND production_order_id=$4 AND status='ABERTA'`, resolution, userID, occurrenceID, orderID)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return errors.New("ocorrência não encontrada ou já resolvida")
	}
	return nil
}

// Transition moves a production order from one status to another with optimistic locking and validation rules.
func (r *ProductionRepository) Transition(orderID int64, from, to, note string, version, userID int) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	var current string
	var currentVersion int
	if err = tx.QueryRow(`SELECT status,version FROM production_orders WHERE id=$1 FOR UPDATE`, orderID).Scan(&current, &currentVersion); err != nil {
		return err
	}
	if current != from || currentVersion != version {
		return errors.New("a ordem foi atualizada por outro usuário; recarregue a tela")
	}
	var blockers int
	if err = tx.QueryRow(`SELECT count(*) FROM production_occurrences WHERE production_order_id=$1 AND status='ABERTA' AND severity='BLOQUEANTE'`, orderID).Scan(&blockers); err != nil {
		return err
	}
	if blockers > 0 && to != models.ProductionBlocked {
		return errors.New("existem ocorrências bloqueantes em aberto")
	}
	if to == models.ProductionMaterialOK {
		var incomplete int
		if err = tx.QueryRow(`SELECT count(*) FROM production_order_items WHERE production_order_id=$1 AND received_quantity<expected_quantity`, orderID).Scan(&incomplete); err != nil {
			return err
		}
		if incomplete > 0 {
			var partial int
			if err = tx.QueryRow(`SELECT count(*) FROM production_occurrences WHERE production_order_id=$1 AND status='ABERTA' AND severity='PARCIAL'`, orderID).Scan(&partial); err != nil {
				return err
			}
			if partial == 0 {
				return errors.New("recebimento incompleto exige ocorrência PARCIAL aberta")
			}
		}
		var receiptCount, missingInvoice int
		if err = tx.QueryRow(`SELECT count(*),count(*) FILTER(WHERE invoice_url='') FROM production_receipts WHERE production_order_id=$1`, orderID).Scan(&receiptCount, &missingInvoice); err != nil {
			return err
		}
		if receiptCount == 0 || missingInvoice > 0 {
			return errors.New("a nota fiscal e sua foto/arquivo são obrigatórias para concluir a conferência")
		}
		if _, err = tx.Exec(`UPDATE production_order_items SET approved_quantity=received_quantity-rejected_quantity,updated_at=NOW() WHERE production_order_id=$1`, orderID); err != nil {
			return err
		}
		var approved int
		if err = tx.QueryRow(`SELECT COALESCE(sum(received_quantity-rejected_quantity),0) FROM production_order_items WHERE production_order_id=$1`, orderID).Scan(&approved); err != nil {
			return err
		}
		if approved <= 0 {
			return errors.New("não há quantidade aprovada para produção")
		}
	}
	if to == models.ProductionShipped {
		var volumes, fiscal, shipment int
		if err = tx.QueryRow(`SELECT (SELECT count(*) FROM production_volumes WHERE production_order_id=$1),(SELECT count(*) FROM production_fiscal_documents WHERE production_order_id=$1),(SELECT count(*) FROM production_shipments WHERE production_order_id=$1 AND method<>'' AND (method<>'CORREIOS' OR tracking_code<>''))`, orderID).Scan(&volumes, &fiscal, &shipment); err != nil {
			return err
		}
		if volumes == 0 || fiscal == 0 || shipment == 0 {
			return errors.New("expedição exige volume, documento fiscal e transporte/rastreio")
		}
	}
	if to == models.ProductionSentEngraving {
		var sent bool
		if err = tx.QueryRow(`SELECT EXISTS(SELECT 1 FROM production_engraving_events WHERE production_order_id=$1 AND event_type='ENVIO' AND file_url<>'')`, orderID).Scan(&sent); err != nil {
			return err
		}
		if !sent {
			return errors.New("anexe o comprovante do envio à gravação")
		}
	}
	if to == models.ProductionFirstPieceApproved {
		var piece bool
		if err = tx.QueryRow(`SELECT EXISTS(SELECT 1 FROM production_engraving_events WHERE production_order_id=$1 AND event_type='PRIMEIRA_PECA' AND file_url<>'')`, orderID).Scan(&piece); err != nil {
			return err
		}
		if !piece {
			return errors.New("anexe a foto da primeira peça antes da aprovação")
		}
	}
	if to == models.ProductionDelivered || to == models.ProductionCompleted {
		var proof bool
		if err = tx.QueryRow(`SELECT EXISTS(SELECT 1 FROM production_shipments WHERE production_order_id=$1 AND proof_url<>'')`, orderID).Scan(&proof); err != nil {
			return err
		}
		if !proof {
			return errors.New("entrega exige comprovante")
		}
	}
	completed := "NULL"
	if to == models.ProductionCompleted {
		completed = "NOW()"
	}
	q := `UPDATE production_orders SET status=$1,version=version+1,completed_at=` + completed + `,
	 conference_started_at=CASE WHEN $1='EM_CONFERENCIA' THEN COALESCE(conference_started_at,NOW()) ELSE conference_started_at END,
	 conference_completed_at=CASE WHEN $1='MATERIAL_OK' THEN NOW() ELSE conference_completed_at END,
	 conference_user_id=CASE WHEN $1 IN ('EM_CONFERENCIA','MATERIAL_OK') THEN $3 ELSE conference_user_id END,updated_at=NOW() WHERE id=$2`
	if _, err = tx.Exec(q, to, orderID, userID); err != nil {
		return err
	}
	_, err = tx.Exec(`INSERT INTO production_history(production_order_id,action,from_status,to_status,details,user_id) VALUES($1,'TRANSICAO',$2,$3,jsonb_build_object('note',$4),$5)`, orderID, from, to, note, userID)
	if err != nil {
		return err
	}
	recipient := "producao"
	if to == models.ProductionAwaitingFirstPiece {
		recipient = "vendas"
	}
	if to == models.ProductionReadyShipment || to == models.ProductionShipped {
		recipient = "logistica"
	}
	if _, err = tx.Exec(`INSERT INTO notifications(production_order_id,recipient_permission,notification_type,message) VALUES($1,$2,'transicao_producao',$3)`, orderID, recipient, fmt.Sprintf("Ordem %d: %s", orderID, to)); err != nil {
		return err
	}
	return tx.Commit()
}

// AddEvent inserts a new engraving event for a production order with idempotency support.
func (r *ProductionRepository) AddEvent(orderID int64, in models.ProductionEventInput, userID int) error {
	res, err := r.db.Exec(`INSERT INTO production_engraving_events(production_order_id,event_type,status,quantity,carrier_id,tracking_code,file_url,notes,idempotency_key,created_by) VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10) ON CONFLICT(production_order_id,idempotency_key) DO NOTHING`, orderID, in.EventType, in.Status, in.Quantity, in.CarrierID, in.TrackingCode, in.FileURL, in.Notes, in.IdempotencyKey, userID)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n > 0 {
		return r.addOperationalHistory(orderID, "EVENTO_GRAVACAO", map[string]interface{}{"type": in.EventType, "status": in.Status}, userID)
	}
	return nil
}

// AddVolume inserts a new volume record for a production order.
func (r *ProductionRepository) AddVolume(orderID int64, in models.ProductionVolumeInput, userID int) error {
	_, err := r.db.Exec(`INSERT INTO production_volumes(production_order_id,label,weight_kg,length_cm,width_cm,height_cm,created_by) VALUES($1,$2,$3,$4,$5,$6,$7)`, orderID, in.Label, in.WeightKG, in.LengthCM, in.WidthCM, in.HeightCM, userID)
	if err != nil {
		return err
	}
	return r.addOperationalHistory(orderID, "VOLUME_ADICIONADO", map[string]interface{}{"label": in.Label, "weight_kg": in.WeightKG}, userID)
}

// AddFiscal inserts or updates a fiscal document record for a production order.
func (r *ProductionRepository) AddFiscal(orderID int64, in models.ProductionFiscalInput, userID int) error {
	_, err := r.db.Exec(`INSERT INTO production_fiscal_documents(production_order_id,document_type,document_number,access_key,file_url,issued_at,created_by) VALUES($1,$2,$3,$4,$5,$6,$7) ON CONFLICT(production_order_id,document_type,document_number) DO UPDATE SET access_key=EXCLUDED.access_key,file_url=EXCLUDED.file_url,issued_at=EXCLUDED.issued_at`, orderID, in.DocumentType, in.DocumentNumber, in.AccessKey, in.FileURL, in.IssuedAt, userID)
	if err != nil {
		return err
	}
	return r.addOperationalHistory(orderID, "DOCUMENTO_FISCAL", map[string]interface{}{"type": in.DocumentType, "number": in.DocumentNumber}, userID)
}

// UpsertShipment creates or updates shipment information for a production order.
func (r *ProductionRepository) UpsertShipment(orderID int64, in models.ProductionShipmentInput, userID int) error {
	_, err := r.db.Exec(`INSERT INTO production_shipments(production_order_id,method,carrier_id,driver_id,tracking_code,tracking_url,postal_service,proof_url,updated_by,shipped_at,delivered_at) VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,CASE WHEN $5<>'' THEN NOW() END,CASE WHEN $8<>'' THEN NOW() END) ON CONFLICT(production_order_id) DO UPDATE SET method=EXCLUDED.method,carrier_id=EXCLUDED.carrier_id,driver_id=EXCLUDED.driver_id,tracking_code=EXCLUDED.tracking_code,tracking_url=EXCLUDED.tracking_url,postal_service=EXCLUDED.postal_service,proof_url=EXCLUDED.proof_url,updated_by=EXCLUDED.updated_by,updated_at=NOW(),shipped_at=COALESCE(production_shipments.shipped_at,EXCLUDED.shipped_at),delivered_at=COALESCE(production_shipments.delivered_at,EXCLUDED.delivered_at)`, orderID, in.Method, in.CarrierID, in.DriverID, in.TrackingCode, in.TrackingURL, in.PostalService, in.ProofURL, userID)
	if err != nil {
		return err
	}
	return r.addOperationalHistory(orderID, "EXPEDICAO_ATUALIZADA", map[string]interface{}{"method": in.Method, "tracking_code": in.TrackingCode, "proof": in.ProofURL != ""}, userID)
}

func (r *ProductionRepository) addOperationalHistory(orderID int64, action string, details interface{}, userID int) error {
	raw, _ := json.Marshal(details)
	_, err := r.db.Exec(`INSERT INTO production_history(production_order_id,action,details,user_id) VALUES($1,$2,$3,$4)`, orderID, action, raw, userID)
	return err
}

// ListSupplies returns all production supplies from the database.
func (r *ProductionRepository) ListSupplies() ([]map[string]interface{}, error) {
	rows, err := r.db.Query(`SELECT id,name,unit,current_quantity,minimum_quantity,active,updated_at FROM production_supplies ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []map[string]interface{}{}
	for rows.Next() {
		var id int64
		var name, unit string
		var cur, min float64
		var active bool
		var updated interface{}
		if err = rows.Scan(&id, &name, &unit, &cur, &min, &active, &updated); err != nil {
			return nil, err
		}
		out = append(out, map[string]interface{}{"id": id, "name": name, "unit": unit, "current_quantity": cur, "minimum_quantity": min, "active": active, "updated_at": updated})
	}
	return out, rows.Err()
}

// CreateSupply inserts a new production supply record into the database.
func (r *ProductionRepository) CreateSupply(in models.ProductionSupplyInput) (int64, error) {
	var id int64
	err := r.db.QueryRow(`INSERT INTO production_supplies(name,unit,current_quantity,minimum_quantity) VALUES($1,$2,$3,$4) RETURNING id`, strings.TrimSpace(in.Name), in.Unit, in.CurrentQuantity, in.MinimumQuantity).Scan(&id)
	return id, err
}

// MoveSupply records a supply movement (entry or exit) and updates the current quantity.
func (r *ProductionRepository) MoveSupply(in models.ProductionSupplyMovementInput, userID int) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	delta := in.Quantity
	if in.MovementType == "SAIDA" {
		delta = -delta
	}
	res, err := tx.Exec(`UPDATE production_supplies SET current_quantity=current_quantity+$1,updated_at=NOW() WHERE id=$2 AND current_quantity+$1>=0`, delta, in.SupplyID)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return errors.New("saldo de insumo insuficiente")
	}
	_, err = tx.Exec(`INSERT INTO production_supply_movements(supply_id,production_order_id,movement_type,quantity,reason,created_by,supplier_id,invoice_number,invoice_url) VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9)`, in.SupplyID, in.OrderID, in.MovementType, in.Quantity, in.Reason, userID, in.SupplierID, in.InvoiceNumber, in.InvoiceURL)
	if err != nil {
		return err
	}
	if in.OrderID != nil {
		if _, err = tx.Exec(`INSERT INTO production_history(production_order_id,action,details,user_id) VALUES($1,'MOVIMENTO_INSUMO',jsonb_build_object('supply_id',$2,'type',$3,'quantity',$4,'reason',$5),$6)`, *in.OrderID, in.SupplyID, in.MovementType, in.Quantity, in.Reason, userID); err != nil {
			return err
		}
	}
	return tx.Commit()
}
