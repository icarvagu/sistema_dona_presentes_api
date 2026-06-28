package repositories

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"donapresentes/models"

	"github.com/lib/pq"
)

type SalesWorkflowRepository struct{ db *sql.DB }

func NewSalesWorkflowRepository(db *sql.DB) *SalesWorkflowRepository {
	return &SalesWorkflowRepository{db: db}
}

func (r *SalesWorkflowRepository) QuoteOwner(id int) (int, error) {
	var owner int
	err := r.db.QueryRow(`SELECT seller_id FROM quotes WHERE id=$1`, id).Scan(&owner)
	return owner, err
}

func (r *SalesWorkflowRepository) SaleOwner(id int) (int, error) {
	var owner int
	err := r.db.QueryRow(`SELECT seller_id FROM sales WHERE id=$1`, id).Scan(&owner)
	return owner, err
}

func (r *SalesWorkflowRepository) ItemOwner(entity string, id int) (int, error) {
	var owner int
	var err error
	if entity == "quote_item" {
		err = r.db.QueryRow(`SELECT q.seller_id FROM quote_items i JOIN quotes q ON q.id=i.quote_id WHERE i.id=$1`, id).Scan(&owner)
	} else {
		err = r.db.QueryRow(`SELECT s.seller_id FROM sale_items i JOIN sales s ON s.id=i.sale_id WHERE i.id=$1`, id).Scan(&owner)
	}
	return owner, err
}

func (r *SalesWorkflowRepository) AddFeedback(quoteID, userID int, scheduledAt *time.Time, observation string) (*models.QuoteFeedbackEvent, error) {
	e := &models.QuoteFeedbackEvent{QuoteID: quoteID, ScheduledAt: scheduledAt, Observation: observation, CreatedBy: userID}
	tx, err := r.db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	err = tx.QueryRow(`INSERT INTO quote_feedback_events (quote_id,scheduled_at,observation,created_by) VALUES ($1,$2,$3,$4) RETURNING id,created_at`, quoteID, scheduledAt, observation, userID).Scan(&e.ID, &e.CreatedAt)
	if err != nil {
		return nil, err
	}
	_, err = tx.Exec(`UPDATE quotes SET feedback_datetime=$1,feedback_observation=$2,updated_at=now() WHERE id=$3`, scheduledAt, observation, quoteID)
	if err != nil {
		return nil, err
	}
	if _, err = tx.Exec(`INSERT INTO notifications(recipient_user_id,notification_type,message)
		SELECT seller_id,'feedback_orcamento',$1 FROM quotes WHERE id=$2`, fmt.Sprintf("Feedback do orçamento %d agendado", quoteID), quoteID); err != nil {
		return nil, err
	}
	if err = tx.Commit(); err != nil {
		return nil, err
	}
	return e, nil
}

func (r *SalesWorkflowRepository) Feedbacks(quoteID int) ([]models.QuoteFeedbackEvent, error) {
	rows, err := r.db.Query(`SELECT id,quote_id,scheduled_at,observation,created_by,created_at FROM quote_feedback_events WHERE quote_id=$1 ORDER BY created_at DESC,id DESC`, quoteID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := []models.QuoteFeedbackEvent{}
	for rows.Next() {
		var e models.QuoteFeedbackEvent
		if err = rows.Scan(&e.ID, &e.QuoteID, &e.ScheduledAt, &e.Observation, &e.CreatedBy, &e.CreatedAt); err != nil {
			return nil, err
		}
		result = append(result, e)
	}
	return result, rows.Err()
}

func (r *SalesWorkflowRepository) CreateLayout(entity string, itemID, userID int, fileURL string) (*models.ItemLayoutVersion, error) {
	tx, err := r.db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	var snapshot []byte
	var query string
	if entity == "quote_item" {
		query = `SELECT to_jsonb(i) FROM quote_items i WHERE id=$1`
	} else {
		query = `SELECT to_jsonb(i) FROM sale_items i WHERE id=$1`
	}
	if err = tx.QueryRow(query, itemID).Scan(&snapshot); err != nil {
		return nil, err
	}
	v := &models.ItemLayoutVersion{EntityType: entity, ItemID: itemID, FileURL: fileURL, CreatedBy: userID, ItemSnapshot: snapshot, ApprovalStatus: "pending"}
	if _, err = tx.Exec(`SELECT pg_advisory_xact_lock(hashtext($1),$2)`, entity, itemID); err != nil {
		return nil, err
	}
	err = tx.QueryRow(`SELECT COALESCE(MAX(version),0)+1 FROM item_layout_versions WHERE entity_type=$1 AND item_id=$2`, entity, itemID).Scan(&v.Version)
	if err != nil {
		return nil, err
	}
	if v.Version == 1 {
		v.Label = "Original"
	} else {
		v.Label = fmt.Sprintf("Alteração %d", v.Version-1)
	}
	err = tx.QueryRow(`INSERT INTO item_layout_versions(entity_type,item_id,version,label,file_url,item_snapshot,created_by) VALUES($1,$2,$3,$4,$5,$6,$7) RETURNING id,created_at`, entity, itemID, v.Version, v.Label, fileURL, snapshot, userID).Scan(&v.ID, &v.CreatedAt)
	if err != nil {
		return nil, err
	}
	if _, err = tx.Exec(`INSERT INTO notifications(recipient_permission,notification_type,message) VALUES('arte_final','novo_layout',$1)`, fmt.Sprintf("Novo layout para análise: %s item %d, versão %d", entity, itemID, v.Version)); err != nil {
		return nil, err
	}
	if err = tx.Commit(); err != nil {
		return nil, err
	}
	return v, nil
}

func (r *SalesWorkflowRepository) ApproveLayout(id int64, userID int, status, note string) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	res, err := tx.Exec(`UPDATE item_layout_versions SET approval_status=$1,approval_note=$2,approved_by=$3,approved_at=now() WHERE id=$4 AND approval_status='pending'`, status, note, userID, id)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return sql.ErrNoRows
	}
	if _, err = tx.Exec(`INSERT INTO layout_approval_events(layout_version_id,status,note,created_by) VALUES($1,$2,$3,$4)`, id, status, note, userID); err != nil {
		return err
	}
	if _, err = tx.Exec(`INSERT INTO notifications(recipient_user_id,notification_type,message)
		SELECT COALESCE(q.seller_id,s.seller_id),'layout_analisado',$1
		FROM item_layout_versions l
		LEFT JOIN quote_items qi ON l.entity_type='quote_item' AND qi.id=l.item_id LEFT JOIN quotes q ON q.id=qi.quote_id
		LEFT JOIN sale_items si ON l.entity_type='sale_item' AND si.id=l.item_id LEFT JOIN sales s ON s.id=si.sale_id
		WHERE l.id=$2`, fmt.Sprintf("Layout %d atualizado para %s", id, status), id); err != nil {
		return err
	}
	return tx.Commit()
}

func (r *SalesWorkflowRepository) ConvertQuote(quoteID, userID int, dates map[string]*time.Time) (int, error) {
	tx, err := r.db.Begin()
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()
	var converted sql.NullInt64
	var sellerID, customerID, installments int
	var paymentMethod, careOf string
	var total float64
	var installmentDates []byte
	err = tx.QueryRow(`SELECT converted_sale_id,seller_id,customer_id,COALESCE(payment_method,''),COALESCE(installments,1),COALESCE(installment_dates,'[]'::jsonb),total_value,COALESCE(care_of,'') FROM quotes WHERE id=$1 FOR UPDATE`, quoteID).Scan(&converted, &sellerID, &customerID, &paymentMethod, &installments, &installmentDates, &total, &careOf)
	if err != nil {
		return 0, err
	}
	if converted.Valid {
		return int(converted.Int64), tx.Commit()
	}
	rows, err := tx.Query(`SELECT id,product_id,quantity,unit_price,total_price,COALESCE(personalization_type,''),COALESCE(engravings,'[]'::jsonb) FROM quote_items WHERE quote_id=$1 ORDER BY id`, quoteID)
	if err != nil {
		return 0, err
	}
	type qi struct {
		id, product, qty int
		unit, total      float64
		personal         string
		eng              []byte
	}
	items := []qi{}
	for rows.Next() {
		var i qi
		if err = rows.Scan(&i.id, &i.product, &i.qty, &i.unit, &i.total, &i.personal, &i.eng); err != nil {
			rows.Close()
			return 0, err
		}
		if i.personal != "" {
			var ok bool
			err = tx.QueryRow(`SELECT EXISTS(SELECT 1 FROM item_layout_versions WHERE entity_type='quote_item' AND item_id=$1 AND approval_status='approved')`, i.id).Scan(&ok)
			if err != nil || !ok {
				rows.Close()
				if err != nil {
					return 0, err
				}
				return 0, fmt.Errorf("item personalizado %d exige layout aprovado", i.id)
			}
			if dates[fmt.Sprint(i.id)] == nil {
				rows.Close()
				return 0, fmt.Errorf("item personalizado %d exige data de retirada da gravação", i.id)
			}
		}
		items = append(items, i)
	}
	rows.Close()
	if paymentMethod == "" {
		paymentMethod = "A definir"
	}
	if installments < 1 {
		installments = 1
	}
	var saleID int
	err = tx.QueryRow(`INSERT INTO sales(seller_id,customer_id,payment_method,installments,total_value,status,care_of,installment_dates,quote_id) VALUES($1,$2,$3,$4,$5,'Aguardando aprovações',$6,$7,$8) RETURNING id`, sellerID, customerID, paymentMethod, installments, total, careOf, installmentDates, quoteID).Scan(&saleID)
	if err != nil {
		return 0, err
	}
	for _, i := range items {
		var saleItemID int
		err = tx.QueryRow(`INSERT INTO sale_items(sale_id,product_id,quantity,unit_price,total_price,engravings,quote_item_id,personalization_type,engraving_withdrawal_date) VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9) RETURNING id`, saleID, i.product, i.qty, i.unit, i.total, i.eng, i.id, i.personal, dates[fmt.Sprint(i.id)]).Scan(&saleItemID)
		if err != nil {
			return 0, err
		}
		_, err = tx.Exec(`INSERT INTO item_layout_versions(entity_type,item_id,version,label,file_url,item_snapshot,approval_status,approval_note,approved_by,approved_at,created_by,created_at) SELECT 'sale_item',$1,version,label,file_url,item_snapshot,approval_status,approval_note,approved_by,approved_at,created_by,created_at FROM item_layout_versions WHERE entity_type='quote_item' AND item_id=$2`, saleItemID, i.id)
		if err != nil {
			return 0, err
		}
	}
	_, err = tx.Exec(`UPDATE quotes SET converted_sale_id=$1,updated_at=now() WHERE id=$2`, saleID, quoteID)
	if err != nil {
		return 0, err
	}
	if err = tx.Commit(); err != nil {
		return 0, err
	}
	return saleID, nil
}

func (r *SalesWorkflowRepository) UpsertFinancial(saleID, userID int, input models.FinancialAnalysisInput) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if _, err = tx.Exec(`INSERT INTO financial_analyses(sale_id,status,tags,observation,attachment_url,analyzed_by) VALUES($1,$2,$3,$4,$5,$6) ON CONFLICT(sale_id) DO UPDATE SET status=EXCLUDED.status,tags=EXCLUDED.tags,observation=EXCLUDED.observation,attachment_url=EXCLUDED.attachment_url,analyzed_by=EXCLUDED.analyzed_by,analyzed_at=now(),updated_at=now()`, saleID, input.Status, pq.Array(input.Tags), input.Observation, input.AttachmentURL, userID); err != nil {
		return err
	}
	if _, err = tx.Exec(`INSERT INTO financial_analysis_events(sale_id,status,tags,observation,attachment_url,created_by) VALUES($1,$2,$3,$4,$5,$6)`, saleID, input.Status, pq.Array(input.Tags), input.Observation, input.AttachmentURL, userID); err != nil {
		return err
	}
	if _, err = tx.Exec(`INSERT INTO notifications(recipient_user_id,notification_type,message)
		SELECT seller_id,'analise_financeira',$1 FROM sales WHERE id=$2`, fmt.Sprintf("Análise financeira do pedido %d: %s", saleID, input.Status), saleID); err != nil {
		return err
	}
	return tx.Commit()
}

func (r *SalesWorkflowRepository) AddReceipt(saleID, userID int, url string) (int64, error) {
	tx, err := r.db.Begin()
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()
	var id int64
	if err = tx.QueryRow(`INSERT INTO sale_payment_receipts(sale_id,file_url,uploaded_by) VALUES($1,$2,$3) RETURNING id`, saleID, url, userID).Scan(&id); err != nil {
		return 0, err
	}
	if _, err = tx.Exec(`INSERT INTO notifications(recipient_permission,notification_type,message) VALUES('financeiro','novo_comprovante',$1)`, fmt.Sprintf("Novo comprovante no pedido %d", saleID)); err != nil {
		return 0, err
	}
	return id, tx.Commit()
}
func (r *SalesWorkflowRepository) ValidateReceipt(id int64, userID int, status, note string) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	res, err := tx.Exec(`UPDATE sale_payment_receipts SET status=$1,observation=$2,validated_by=$3,validated_at=now() WHERE id=$4`, status, note, userID, id)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return sql.ErrNoRows
	}
	if _, err = tx.Exec(`INSERT INTO notifications(recipient_user_id,notification_type,message)
		SELECT s.seller_id,'comprovante_validado',$1 FROM sale_payment_receipts r JOIN sales s ON s.id=r.sale_id WHERE r.id=$2`, fmt.Sprintf("Comprovante %d: %s", id, status), id); err != nil {
		return err
	}
	return tx.Commit()
}

func (r *SalesWorkflowRepository) AddPending(saleID, userID int, input models.SalePendingInput) (int64, error) {
	var id int64
	err := r.db.QueryRow(`INSERT INTO sale_pending_events(sale_id,sector,description,blocking,created_by) VALUES($1,$2,$3,$4,$5) RETURNING id`, saleID, input.Sector, input.Description, input.Blocking, userID).Scan(&id)
	return id, err
}
func (r *SalesWorkflowRepository) ResolvePending(id int64, userID int, note string) error {
	res, err := r.db.Exec(`UPDATE sale_pending_events SET resolved_at=now(),resolved_by=$1,resolution_note=$2 WHERE id=$3 AND resolved_at IS NULL`, userID, note, id)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (r *SalesWorkflowRepository) PendingSaleOwner(id int64) (int, error) {
	var owner int
	err := r.db.QueryRow(`SELECT s.seller_id FROM sale_pending_events p JOIN sales s ON s.id=p.sale_id WHERE p.id=$1`, id).Scan(&owner)
	return owner, err
}

func (r *SalesWorkflowRepository) AddEngravingApproval(itemID, userID int, input models.EngravingApprovalInput) (int64, error) {
	tx, err := r.db.Begin()
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()
	var id int64
	if err = tx.QueryRow(`INSERT INTO engraving_approvals(sale_item_id,response,observation,responded_by) VALUES($1,$2,$3,$4) RETURNING id`, itemID, input.Response, input.Observation, userID).Scan(&id); err != nil {
		return 0, err
	}
	if _, err = tx.Exec(`INSERT INTO notifications(recipient_permission,notification_type,message) VALUES('compras','aprovacao_gravacao',$1)`, fmt.Sprintf("Resposta de gravação do item %d: %s", itemID, input.Response)); err != nil {
		return 0, err
	}
	return id, tx.Commit()
}
func (r *SalesWorkflowRepository) RecordEngravingChannel(id int64, userID int, channel string) error {
	res, err := r.db.Exec(`UPDATE engraving_approvals SET channel=$1,recorded_by=$2,recorded_at=now() WHERE id=$3 AND recorded_at IS NULL`, channel, userID, id)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (r *SalesWorkflowRepository) SetTarget(sellerID, userID int, month time.Time, value float64) error {
	month = time.Date(month.Year(), month.Month(), 1, 0, 0, 0, 0, time.UTC)
	_, err := r.db.Exec(`INSERT INTO seller_monthly_targets(seller_id,month,target_value,created_by) VALUES($1,$2,$3,$4) ON CONFLICT(seller_id,month) DO UPDATE SET target_value=EXCLUDED.target_value,updated_at=now()`, sellerID, month, value, userID)
	return err
}

func (r *SalesWorkflowRepository) SetSellerApproval(saleID, userID int) error {
	if err := r.validateChecklist(r.db, saleID, false); err != nil {
		return err
	}
	res, err := r.db.Exec(`UPDATE sales SET seller_approved_at=COALESCE(seller_approved_at, now()),seller_approved_by=COALESCE(seller_approved_by, $1),updated_at=now() WHERE id=$2`, userID, saleID)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return sql.ErrNoRows
	}
	return nil
}

type workflowQueryer interface {
	QueryRow(string, ...interface{}) *sql.Row
}

func (r *SalesWorkflowRepository) validateChecklist(q workflowQueryer, saleID int, requireSeller bool) error {
	var sellerApproved bool
	if err := q.QueryRow(`SELECT seller_approved_at IS NOT NULL FROM sales WHERE id=$1`, saleID).Scan(&sellerApproved); err != nil {
		return err
	}
	if requireSeller && !sellerApproved {
		return fmt.Errorf("vendedor ainda não deu o OK final")
	}
	var financialOK bool
	if err := q.QueryRow(`SELECT EXISTS(SELECT 1 FROM financial_analyses WHERE sale_id=$1 AND status='approved' AND NOT ('NEGADO'=ANY(tags)))`, saleID).Scan(&financialOK); err != nil {
		return err
	}
	if !financialOK {
		return fmt.Errorf("análise financeira ainda não foi aprovada")
	}
	var blockers int
	err := q.QueryRow(`SELECT
	 (SELECT count(*) FROM sale_pending_events WHERE sale_id=$1 AND blocking AND resolved_at IS NULL)+
	 (SELECT count(*) FROM sale_payment_receipts WHERE sale_id=$1 AND status<>'validated')+
	 (SELECT CASE WHEN EXISTS(
	    SELECT 1 FROM financial_analyses
	    WHERE sale_id=$1
	      AND (tags && ARRAY['50% + BOLETO','SOMENTE À VISTA','50% + SAÍDA']::text[])
	  ) AND NOT EXISTS(
	    SELECT 1 FROM sale_payment_receipts WHERE sale_id=$1 AND status='validated'
	  ) THEN 1 ELSE 0 END)+
	 (SELECT count(*) FROM sale_items i WHERE i.sale_id=$1 AND i.personalization_type<>'' AND
	  (i.engraving_withdrawal_date IS NULL OR
	   NOT EXISTS(SELECT 1 FROM item_layout_versions l WHERE l.entity_type='sale_item' AND l.item_id=i.id AND l.approval_status='approved') OR
	   NOT EXISTS(SELECT 1 FROM engraving_approvals a WHERE a.sale_item_id=i.id AND a.response='APROVADO' AND a.channel IS NOT NULL AND a.recorded_at IS NOT NULL))))`, saleID).Scan(&blockers)
	if err != nil {
		return err
	}
	if blockers > 0 {
		return fmt.Errorf("pedido possui %d pendência(s) bloqueante(s)", blockers)
	}
	return nil
}

func (r *SalesWorkflowRepository) EnsureReadyForPurchases(saleID int) error {
	return r.validateChecklist(r.db, saleID, true)
}

func (r *SalesWorkflowRepository) MarkReleasedToPurchases(saleID int) error {
	res, err := r.db.Exec(`UPDATE sales SET released_to_purchases_at=COALESCE(released_to_purchases_at, now()),status='Liberado para Compras',updated_at=now() WHERE id=$1`, saleID)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (r *SalesWorkflowRepository) ReleaseToPurchases(saleID int) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	var exists int
	if err = tx.QueryRow(`SELECT 1 FROM sales WHERE id=$1 FOR UPDATE`, saleID).Scan(&exists); err != nil {
		return err
	}
	if err = r.validateChecklist(tx, saleID, true); err != nil {
		return err
	}
	_, err = tx.Exec(`UPDATE sales SET released_to_purchases_at=COALESCE(released_to_purchases_at, now()),status='Liberado para Compras',updated_at=now() WHERE id=$1`, saleID)
	if err != nil {
		return err
	}
	return tx.Commit()
}

func (r *SalesWorkflowRepository) SetQuoteImportant(id int, important bool) error {
	res, err := r.db.Exec(`UPDATE quotes SET important=$1,updated_at=now() WHERE id=$2`, important, id)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (r *SalesWorkflowRepository) SaleWorkflow(id int) (map[string]interface{}, error) {
	result := map[string]interface{}{}
	queries := map[string]string{
		"layouts":             `SELECT COALESCE(jsonb_agg(x ORDER BY x.version DESC),'[]'::jsonb) FROM item_layout_versions x JOIN sale_items i ON x.entity_type='sale_item' AND x.item_id=i.id WHERE i.sale_id=$1`,
		"financial_history":   `SELECT COALESCE(jsonb_agg(x ORDER BY x.created_at DESC),'[]'::jsonb) FROM financial_analysis_events x WHERE x.sale_id=$1`,
		"receipts":            `SELECT COALESCE(jsonb_agg(x ORDER BY x.created_at DESC),'[]'::jsonb) FROM sale_payment_receipts x WHERE x.sale_id=$1`,
		"pending_events":      `SELECT COALESCE(jsonb_agg(x ORDER BY x.created_at DESC),'[]'::jsonb) FROM sale_pending_events x WHERE x.sale_id=$1`,
		"engraving_approvals": `SELECT COALESCE(jsonb_agg(a ORDER BY a.responded_at DESC),'[]'::jsonb) FROM engraving_approvals a JOIN sale_items i ON i.id=a.sale_item_id WHERE i.sale_id=$1`,
	}
	for key, query := range queries {
		var raw []byte
		if err := r.db.QueryRow(query, id).Scan(&raw); err != nil {
			return nil, err
		}
		result[key] = json.RawMessage(raw)
	}
	return result, nil
}

func (r *SalesWorkflowRepository) Dashboard(sellerID int, all bool) (map[string]interface{}, error) {
	result := map[string]interface{}{}
	var quotes, important int
	err := r.db.QueryRow(`SELECT count(*),count(*) FILTER(WHERE important) FROM quotes WHERE ($1 OR seller_id=$2)`, all, sellerID).Scan(&quotes, &important)
	if err != nil {
		return nil, err
	}
	result["quotes"] = quotes
	result["important_quotes"] = important
	var due int
	err = r.db.QueryRow(`SELECT count(*) FROM quotes q
		LEFT JOIN LATERAL (SELECT scheduled_at FROM quote_feedback_events WHERE quote_id=q.id ORDER BY created_at DESC,id DESC LIMIT 1) e ON true
		WHERE ($1 OR q.seller_id=$2) AND q.converted_sale_id IS NULL
		AND COALESCE(e.scheduled_at,q.feedback_datetime,q.created_at+interval '3 days')<=now()`, all, sellerID).Scan(&due)
	if err != nil {
		return nil, err
	}
	result["overdue_feedbacks"] = due
	var pending, layouts, approvals int
	err = r.db.QueryRow(`SELECT count(DISTINCT p.id) FILTER(WHERE p.resolved_at IS NULL),count(DISTINCT l.id) FILTER(WHERE l.approval_status='pending'),count(DISTINCT a.id) FILTER(WHERE a.recorded_at IS NULL) FROM sales s LEFT JOIN sale_pending_events p ON p.sale_id=s.id LEFT JOIN sale_items i ON i.sale_id=s.id LEFT JOIN item_layout_versions l ON l.entity_type='sale_item' AND l.item_id=i.id LEFT JOIN engraving_approvals a ON a.sale_item_id=i.id WHERE ($1 OR s.seller_id=$2)`, all, sellerID).Scan(&pending, &layouts, &approvals)
	if err != nil {
		return nil, err
	}
	result["open_sector_pending"] = pending
	result["pending_layouts"] = layouts
	result["pending_engraving_records"] = approvals
	var sold, target float64
	month := time.Date(time.Now().Year(), time.Now().Month(), 1, 0, 0, 0, 0, time.UTC)
	err = r.db.QueryRow(`SELECT COALESCE((SELECT sum(total_value) FROM sales WHERE seller_id=$1 AND created_at>=date_trunc('month',now())),0),COALESCE((SELECT target_value FROM seller_monthly_targets WHERE seller_id=$1 AND month=$2),0)`, sellerID, month).Scan(&sold, &target)
	if err != nil {
		return nil, err
	}
	result["monthly_sold"] = sold
	result["monthly_target"] = target
	result["target_remaining"] = maxFloat(target-sold, 0)

	queueQueries := map[string]string{
		"quote_queue": `SELECT COALESCE(jsonb_agg(row_to_json(x) ORDER BY x.important DESC,x.feedback_at,x.id),'[]'::jsonb) FROM (
			SELECT q.id,q.quote_number,q.responsible_name,c.name customer_name,q.total_value total,q.important,q.converted_sale_id,
			COALESCE(e.scheduled_at,q.feedback_datetime,q.created_at+interval '3 days') feedback_at,
			COALESCE(e.observation,q.feedback_observation,'') feedback_observation
			FROM quotes q JOIN customers c ON c.id=q.customer_id
			LEFT JOIN LATERAL (SELECT scheduled_at,observation FROM quote_feedback_events WHERE quote_id=q.id ORDER BY created_at DESC,id DESC LIMIT 1) e ON true
			WHERE ($1 OR q.seller_id=$2) AND q.converted_sale_id IS NULL
			ORDER BY q.important DESC,feedback_at,q.id LIMIT 40) x`,
		"feedback_queue": `SELECT COALESCE(jsonb_agg(row_to_json(x) ORDER BY x.overdue DESC,x.scheduled_at,x.quote_id),'[]'::jsonb) FROM (
			SELECT q.id quote_id,q.quote_number,c.name customer_name,q.important,
			COALESCE(e.scheduled_at,q.feedback_datetime,q.created_at+interval '3 days') scheduled_at,
			COALESCE(e.observation,q.feedback_observation,'') observation,
			COALESCE(e.scheduled_at,q.feedback_datetime,q.created_at+interval '3 days')<now() overdue
			FROM quotes q JOIN customers c ON c.id=q.customer_id
			LEFT JOIN LATERAL (SELECT scheduled_at,observation FROM quote_feedback_events WHERE quote_id=q.id ORDER BY created_at DESC,id DESC LIMIT 1) e ON true
			WHERE ($1 OR q.seller_id=$2) AND q.converted_sale_id IS NULL
			AND COALESCE(e.scheduled_at,q.feedback_datetime,q.created_at+interval '3 days')<=now()+interval '7 days'
			ORDER BY overdue DESC,scheduled_at LIMIT 40) x`,
		"layout_queue": `SELECT COALESCE(jsonb_agg(row_to_json(x) ORDER BY x.approval_status,x.created_at),'[]'::jsonb) FROM (
			SELECT DISTINCT ON (l.entity_type,l.item_id) l.id,l.entity_type,l.item_id,l.version,l.label,l.file_url,l.approval_status,l.approval_note,l.created_at,
			COALESCE(q.id,s.id) entity_id,COALESCE(q.quote_number,s.id::text) entity_number,p.product_name
			FROM item_layout_versions l
			LEFT JOIN quote_items qi ON l.entity_type='quote_item' AND qi.id=l.item_id LEFT JOIN quotes q ON q.id=qi.quote_id
			LEFT JOIN sale_items si ON l.entity_type='sale_item' AND si.id=l.item_id LEFT JOIN sales s ON s.id=si.sale_id
			JOIN products p ON p.id=COALESCE(qi.product_id,si.product_id)
			WHERE ($1 OR COALESCE(q.seller_id,s.seller_id)=$2)
			ORDER BY l.entity_type,l.item_id,l.version DESC) x`,
		"sale_stages": `SELECT COALESCE(jsonb_agg(row_to_json(x) ORDER BY x.created_at,x.id),'[]'::jsonb) FROM (
			SELECT s.id,c.name customer_name,s.total_value total,s.status,s.created_at,
			CASE WHEN s.released_to_purchases_at IS NOT NULL THEN 'released'
			 WHEN s.seller_approved_at IS NOT NULL THEN 'seller_approved'
			 WHEN EXISTS(SELECT 1 FROM sale_pending_events p WHERE p.sale_id=s.id AND p.blocking AND p.resolved_at IS NULL) THEN 'blocked'
			 WHEN EXISTS(SELECT 1 FROM financial_analyses f WHERE f.sale_id=s.id AND f.status='approved') THEN 'final_checklist'
			 WHEN EXISTS(SELECT 1 FROM financial_analyses f WHERE f.sale_id=s.id) THEN 'financial_analysis'
			 ELSE 'awaiting_financial' END stage
			FROM sales s JOIN customers c ON c.id=s.customer_id WHERE ($1 OR s.seller_id=$2)
			ORDER BY s.created_at DESC LIMIT 60) x`,
		"pending_queue": `SELECT COALESCE(jsonb_agg(row_to_json(x) ORDER BY x.blocking DESC,x.sector,x.created_at),'[]'::jsonb) FROM (
			SELECT p.id,p.sale_id,p.sector,p.description,p.blocking,p.created_at,c.name customer_name
			FROM sale_pending_events p JOIN sales s ON s.id=p.sale_id JOIN customers c ON c.id=s.customer_id
			WHERE p.resolved_at IS NULL AND ($1 OR s.seller_id=$2) ORDER BY p.blocking DESC,p.created_at LIMIT 60) x`,
		"engraving_queue": `SELECT COALESCE(jsonb_agg(row_to_json(x) ORDER BY x.responded_at),'[]'::jsonb) FROM (
			SELECT a.id,a.sale_item_id,i.sale_id,a.response,a.observation,a.channel,a.recorded_at,a.responded_at,p.product_name
			FROM engraving_approvals a JOIN sale_items i ON i.id=a.sale_item_id JOIN sales s ON s.id=i.sale_id JOIN products p ON p.id=i.product_id
			WHERE a.recorded_at IS NULL AND ($1 OR s.seller_id=$2) ORDER BY a.responded_at LIMIT 40) x`,
	}
	for key, query := range queueQueries {
		var raw []byte
		if err = r.db.QueryRow(query, all, sellerID).Scan(&raw); err != nil {
			return nil, err
		}
		result[key] = json.RawMessage(raw)
	}
	return result, nil
}

func maxFloat(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}

func MarshalWorkflow(v interface{}) json.RawMessage { b, _ := json.Marshal(v); return b }
