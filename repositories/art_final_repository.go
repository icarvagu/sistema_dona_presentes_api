package repositories

import (
	"database/sql"
	"encoding/json"
	"fmt"

	"donapresentes/models"

	"github.com/lib/pq"
)

type ArtFinalRepository struct{ db *sql.DB }

func NewArtFinalRepository(db *sql.DB) *ArtFinalRepository { return &ArtFinalRepository{db: db} }

func ensureActiveArtFinalUser(tx *sql.Tx, id *int, permissions ...string) error {
	if id == nil {
		return nil
	}
	var ok bool
	if err := tx.QueryRow(`SELECT EXISTS(SELECT 1 FROM users WHERE id=$1 AND status='Ativo' AND (role='admin' OR permissions && $2::text[]))`, *id, pq.Array(permissions)).Scan(&ok); err != nil {
		return err
	}
	if !ok {
		return fmt.Errorf("responsavel inexistente ou inativo")
	}
	return nil
}

func (r *ArtFinalRepository) Dashboard(access models.ArtFinalAccess, filters models.ArtFinalFilters) (map[string]interface{}, error) {
	result := map[string]interface{}{"manage": access.Manage, "marketing": access.Marketing || access.Admin, "layout": access.Layout, "purchases": access.Purchases, "production_profile": access.ProductionProfile, "corel": access.Corel, "engraving": access.Engraving, "can_confirm_product": access.Admin || access.ProductionProfile || access.Purchases}
	queries := []struct {
		key  string
		sql  string
		args []interface{}
	}{
		{
			key: "pending",
			sql: `WITH latest_layout AS (
				SELECT DISTINCT ON (l.entity_type,l.item_id) l.*,COALESCE(q.id,s.id) sale_or_quote_id,
				       COALESCE(q.seller_id,s.seller_id) seller_id,p.product_name
				FROM item_layout_versions l
				LEFT JOIN quote_items qi ON l.entity_type='quote_item' AND qi.id=l.item_id
				LEFT JOIN quotes q ON q.id=qi.quote_id
				LEFT JOIN sale_items si ON l.entity_type='sale_item' AND si.id=l.item_id
				LEFT JOIN sales s ON s.id=si.sale_id
				JOIN products p ON p.id=COALESCE(qi.product_id,si.product_id)
				ORDER BY l.entity_type,l.item_id,l.version DESC
			), queue AS (
				SELECT 'layout'::text source,l.id,l.entity_type,
				 CASE WHEN l.version>1 OR l.approval_status='changes_requested' THEN 'alteracao' ELSE 'layout' END category,
				 l.product_name title,COALESCE(l.approval_note,'') description,
				 CASE WHEN l.entity_type='sale_item' THEN l.sale_or_quote_id END sale_id,NULL::integer purchase_id,
				 NULL::date due_date,CASE WHEN l.approval_status='changes_requested' THEN 3 ELSE 1 END priority,
				 ARRAY[l.label,l.approval_status]::text[] tags,l.file_url attachment_url,'open' status,l.created_at
				FROM latest_layout l WHERE l.approval_status IN ('pending','changes_requested') AND $1
				AND ($9 OR ($10 AND l.seller_id=$11))
				UNION ALL
				SELECT 'corel',po.id,'purchase_order', 'corel', 'Pedido '||po.general_number,
				 'Arquivo Corel solicitado e ainda nao anexado',po.sale_id,po.id,po.material_deadline,
				 CASE WHEN po.material_deadline IS NOT NULL AND po.material_deadline<=CURRENT_DATE+1 THEN 3 ELSE 2 END,
				 ARRAY[po.status]::text[],'','open',po.created_at
				FROM purchase_orders po WHERE po.corel_required AND po.corel_attached_at IS NULL AND $2
				AND ($9 OR ($12 AND (po.buyer_id=$11 OR po.buyer_id IS NULL)))
				UNION ALL
				SELECT 'gravacao',si.id,'sale_item','gravacao',p.product_name,
				 'Aguardando aprovacao/registro da gravacao',si.sale_id,po.id,s.departure_date,
				 CASE WHEN s.departure_date IS NOT NULL AND s.departure_date<=CURRENT_DATE+2 THEN 3 ELSE 1 END,
				 ARRAY['GRAVACAO']::text[],COALESCE(l.file_url,''),'open',si.created_at
				FROM sale_items si JOIN sales s ON s.id=si.sale_id JOIN products p ON p.id=si.product_id
				LEFT JOIN purchase_orders po ON po.sale_id=s.id
				LEFT JOIN LATERAL (SELECT file_url FROM item_layout_versions WHERE entity_type='sale_item' AND item_id=si.id ORDER BY version DESC LIMIT 1) l ON true
				WHERE $3 AND ($9 OR ($10 AND s.seller_id=$11))
				AND (COALESCE(si.personalization_type,'')<>'' OR COALESCE(si.engravings,'[]'::jsonb)<>'[]'::jsonb)
				AND NOT EXISTS (SELECT 1 FROM engraving_approvals ea WHERE ea.sale_item_id=si.id AND ea.response='APROVADO' AND ea.recorded_at IS NOT NULL)
				UNION ALL
				SELECT 'task',t.id,'art_final_task',t.category,t.title,t.description,t.sale_id,t.purchase_id,t.due_date,t.priority,t.tags,t.attachment_url,t.status,t.created_at
				FROM art_final_tasks t WHERE t.panel='pending'
				AND (($1 AND t.category IN ('layout','alteracao')) OR ($2 AND t.category='corel') OR ($3 AND t.category='gravacao'))
				AND ($9
				 OR ($10 AND t.category IN ('layout','alteracao','gravacao') AND EXISTS(SELECT 1 FROM sales own_s WHERE own_s.id=t.sale_id AND own_s.seller_id=$11))
				 OR ($12 AND t.category='corel' AND EXISTS(SELECT 1 FROM purchase_orders own_po WHERE own_po.id=t.purchase_id AND (own_po.buyer_id=$11 OR own_po.buyer_id IS NULL))))
			)
			SELECT COALESCE(jsonb_agg(row_to_json(q) ORDER BY q.priority DESC,q.due_date NULLS LAST,q.created_at),'[]'::jsonb)
			FROM queue q WHERE ($4='' OR q.category=$4) AND ($5='' OR q.status=$5)
			AND ($6='' OR q.title ILIKE '%%'||$6||'%%' OR q.description ILIKE '%%'||$6||'%%')
			AND ($7::date IS NULL OR q.due_date >= $7::date) AND ($8::date IS NULL OR q.due_date <= $8::date)`,
			args: []interface{}{access.Layout, access.Corel, access.Engraving, filters.Category, filters.Status, filters.Search, filters.DateFrom, filters.DateTo, access.Admin || access.ArtFinal, access.Sales, access.UserID, access.Purchases},
		},
		{
			key: "media",
			sql: `SELECT COALESCE(jsonb_agg(row_to_json(t) ORDER BY t.priority DESC,t.due_at NULLS LAST,t.created_at),'[]'::jsonb)
			FROM (SELECT id,panel,category,title,description,sale_id,purchase_id,due_date,due_at,assigned_to,channel,format,priority,tags,attachment_url,status,created_at,updated_at
			FROM art_final_tasks WHERE panel='media' AND $1 AND ($2='' OR category=$2) AND ($3='' OR status=$3)
			AND ($4='' OR title ILIKE '%%'||$4||'%%' OR description ILIKE '%%'||$4||'%%')
			AND ($5::date IS NULL OR COALESCE(due_at,due_date::timestamp) >= $5::date) AND ($6::date IS NULL OR COALESCE(due_at,due_date::timestamp) < $6::date + INTERVAL '1 day')) t`,
			args: []interface{}{access.Media, filters.Category, filters.Status, filters.Search, filters.DateFrom, filters.DateTo},
		},
		{
			key: "production",
			sql: `SELECT COALESCE(jsonb_agg(row_to_json(x) ORDER BY x.departure_date NULLS LAST,x.priority DESC,x.sale_id),'[]'::jsonb)
			FROM (SELECT s.id sale_id,si.id sale_item_id,po.id purchase_id,po.general_number,c.name customer_name,s.departure_date,s.delivery_date,s.priority,p.product_name,
			 1 item_count,(COALESCE(si.engravings,'[]'::jsonb)<>'[]'::jsonb OR COALESCE(si.personalization_type,'')<>'') engraved,
			 p.photos[1] photo_url,l.file_url layout_url,
			 ARRAY_REMOVE(ARRAY[CASE WHEN s.priority='high' THEN 'PRIORIDADE' END,CASE WHEN (COALESCE(si.engravings,'[]'::jsonb)<>'[]'::jsonb OR COALESCE(si.personalization_type,'')<>'') THEN 'GRAVADO' ELSE 'SEM GRAVACAO' END],NULL) tags,
			 EXISTS(SELECT 1 FROM art_final_stories st WHERE st.sale_item_id=si.id) in_stories
			FROM purchase_orders po JOIN sales s ON s.id=po.sale_id JOIN customers c ON c.id=s.customer_id
			JOIN sale_items si ON si.sale_id=s.id JOIN products p ON p.id=si.product_id
			LEFT JOIN LATERAL (SELECT file_url FROM item_layout_versions WHERE entity_type='sale_item' AND item_id=si.id AND approval_status='approved' ORDER BY version DESC LIMIT 1) l ON true
		WHERE $1 AND ($6 OR $7 OR ($8 AND s.seller_id=$10) OR ($9 AND (po.buyer_id=$10 OR po.buyer_id IS NULL)))
			AND $11 IN ('','open')
			AND si.product_received_at IS NOT NULL
			AND (po.production_released_at IS NOT NULL OR po.status IN ('Liberado para Produção','Pedido na Gravação','Aguardando Aprovação da Gravação'))
			AND ($2='' OR ($2='gravados' AND (COALESCE(si.engravings,'[]'::jsonb)<>'[]'::jsonb OR COALESCE(si.personalization_type,'')<>'')) OR ($2='sem_gravacao' AND COALESCE(si.engravings,'[]'::jsonb)='[]'::jsonb AND COALESCE(si.personalization_type,'')=''))
			AND ($3='' OR po.general_number ILIKE '%%'||$3||'%%' OR c.name ILIKE '%%'||$3||'%%' OR p.product_name ILIKE '%%'||$3||'%%')
			AND ($4::date IS NULL OR s.departure_date >= $4::date) AND ($5::date IS NULL OR s.departure_date <= $5::date)
			) x`,
			args: []interface{}{access.Production, filters.Category, filters.Search, filters.DateFrom, filters.DateTo, access.Admin || access.ArtFinal, access.ProductionProfile || access.Marketing, access.Sales, access.Purchases, access.UserID, filters.Status},
		},
		{
			key: "stories",
			sql: `SELECT COALESCE(jsonb_agg(row_to_json(x) ORDER BY x.checked,x.created_at DESC),'[]'::jsonb) FROM (
			SELECT st.id,st.sale_id,st.sale_item_id,st.title,st.tags,st.checked,st.checked_at,st.created_at,st.lifecycle_status,st.observation,st.published_at,st.expires_at,
			 CASE WHEN st.lifecycle_status IN ('published','expired') THEN 'done' ELSE 'open' END status,
			 po.general_number,c.name customer_name,s.departure_date,p.photos[1] photo_url,l.file_url layout_url
			FROM art_final_stories st JOIN sales s ON s.id=st.sale_id JOIN customers c ON c.id=s.customer_id
			LEFT JOIN purchase_orders po ON po.sale_id=s.id LEFT JOIN sale_items si ON si.id=st.sale_item_id
			LEFT JOIN products p ON p.id=COALESCE(si.product_id,(SELECT product_id FROM sale_items WHERE sale_id=s.id ORDER BY id LIMIT 1))
			LEFT JOIN LATERAL (SELECT file_url FROM item_layout_versions WHERE entity_type='sale_item' AND item_id=COALESCE(st.sale_item_id,(SELECT id FROM sale_items WHERE sale_id=s.id ORDER BY id LIMIT 1)) ORDER BY version DESC LIMIT 1) l ON true
			WHERE $1 AND ($3 OR $4 OR ($5 AND s.seller_id=$6))
			AND st.lifecycle_status<>'expired' AND NOT(st.lifecycle_status='published' AND st.expires_at IS NOT NULL AND st.expires_at<=now())
			AND ($2='' OR st.title ILIKE '%%'||$2||'%%' OR c.name ILIKE '%%'||$2||'%%')
			AND ($7='' OR ($7='open' AND NOT st.checked) OR ($7='done' AND st.checked))
			AND ($8::date IS NULL OR s.departure_date >= $8::date) AND ($9::date IS NULL OR s.departure_date <= $9::date)) x`,
			args: []interface{}{access.Stories, filters.Search, access.Admin || access.ArtFinal, access.ProductionProfile, access.Sales, access.UserID, filters.Status, filters.DateFrom, filters.DateTo},
		},
	}
	for _, query := range queries {
		var raw []byte
		if err := r.db.QueryRow(query.sql, query.args...).Scan(&raw); err != nil {
			return nil, err
		}
		result[query.key] = json.RawMessage(raw)
	}
	return result, nil
}

func (r *ArtFinalRepository) ListLayoutRequests(access models.ArtFinalAccess) (json.RawMessage, error) {
	var raw []byte
	err := r.db.QueryRow(`SELECT COALESCE(jsonb_agg(row_to_json(x) ORDER BY x.priority DESC,x.due_at NULLS LAST,x.created_at DESC),'[]'::jsonb) FROM (
		SELECT r.*,u.username assigned_name,COALESCE(req.full_name,req.username,'Usuário') requested_name,
		 CASE WHEN 'compras'=ANY(COALESCE(req.permissions,'{}'::text[])) THEN 'purchases' ELSE 'sales' END requested_sector,COUNT(i.id) item_count,
		 COUNT(i.id) FILTER(WHERE i.status='approved') approved_items,
		 MAX(v.version) latest_version
		FROM layout_requests r LEFT JOIN users u ON u.id=r.assigned_to LEFT JOIN users req ON req.id=r.requested_by
		JOIN layout_request_items i ON i.request_id=r.id
		LEFT JOIN item_layout_versions v ON v.request_item_id=i.id
		WHERE ($1 OR $4 OR ($2 AND r.source_type='sale' AND EXISTS(SELECT 1 FROM sales s WHERE s.id=r.source_id AND s.seller_id=$3))
		 OR ($2 AND r.source_type='quote' AND EXISTS(SELECT 1 FROM quotes q WHERE q.id=r.source_id AND q.seller_id=$3))
		 OR ($5 AND r.source_type='sale' AND EXISTS(SELECT 1 FROM purchase_orders po WHERE po.sale_id=r.source_id AND (po.buyer_id=$3 OR po.buyer_id IS NULL))))
		GROUP BY r.id,u.username,req.id) x`, access.Admin || access.ArtFinal, access.Sales, access.UserID, access.ProductionProfile, access.Purchases).Scan(&raw)
	return json.RawMessage(raw), err
}

func (r *ArtFinalRepository) GetLayoutRequest(id int64, access models.ArtFinalAccess) (json.RawMessage, error) {
	var raw []byte
	err := r.db.QueryRow(`SELECT jsonb_build_object(
		'request',to_jsonb(r)||jsonb_build_object('requested_sector',COALESCE((SELECT CASE WHEN 'compras'=ANY(u.permissions) THEN 'purchases' ELSE 'sales' END FROM users u WHERE u.id=r.requested_by),'sales'),'requested_name',COALESCE((SELECT COALESCE(u.full_name,u.username) FROM users u WHERE u.id=r.requested_by),'Usuário')),
		'groups',COALESCE((SELECT jsonb_agg(to_jsonb(g) ORDER BY g.group_key) FROM layout_request_groups g WHERE g.request_id=r.id),'[]'::jsonb),
		'items',COALESCE((SELECT jsonb_agg(to_jsonb(i)||jsonb_build_object('versions',COALESCE((SELECT jsonb_agg(to_jsonb(v) ORDER BY v.version DESC) FROM item_layout_versions v WHERE v.request_item_id=i.id),'[]'::jsonb),'job_history',COALESCE((SELECT jsonb_agg(to_jsonb(j) ORDER BY j.kind,j.version DESC) FROM layout_job_versions j WHERE j.request_item_id=i.id),'[]'::jsonb),'corel',(SELECT to_jsonb(c) FROM layout_corel_jobs c WHERE c.request_item_id=i.id),'engraving',(SELECT to_jsonb(e) FROM layout_engraving_jobs e WHERE e.request_item_id=i.id)) ORDER BY i.id) FROM layout_request_items i WHERE i.request_id=r.id),'[]'::jsonb),
		'timeline',COALESCE((SELECT jsonb_agg(to_jsonb(a) ORDER BY a.created_at DESC) FROM art_final_audit_log a WHERE a.entity_type IN ('layout_request','layout_item','layout_version','corel','engraving') AND (a.entity_type='layout_request' AND a.entity_id=r.id OR (a.details->>'request_id')::bigint=r.id)),'[]'::jsonb))
		FROM layout_requests r WHERE r.id=$1 AND ($2 OR $5 OR ($3 AND r.source_type='sale' AND EXISTS(SELECT 1 FROM sales s WHERE s.id=r.source_id AND s.seller_id=$4)) OR ($3 AND r.source_type='quote' AND EXISTS(SELECT 1 FROM quotes q WHERE q.id=r.source_id AND q.seller_id=$4)) OR ($6 AND r.source_type='sale' AND EXISTS(SELECT 1 FROM purchase_orders po WHERE po.sale_id=r.source_id AND (po.buyer_id=$4 OR po.buyer_id IS NULL))))`, id, access.Admin || access.ArtFinal, access.Sales, access.UserID, access.ProductionProfile, access.Purchases).Scan(&raw)
	return json.RawMessage(raw), err
}

func (r *ArtFinalRepository) AddLayoutMessage(id int64, input models.LayoutMessageInput, userID int, access models.ArtFinalAccess) error {
	result, err := r.db.Exec(`INSERT INTO art_final_audit_log(entity_type,entity_id,action,details,user_id)
		SELECT 'layout_request',$1,'message',jsonb_build_object('message',$2,'file_url',$3),$4
		FROM layout_requests lr WHERE lr.id=$1 AND ($5 OR ($6 AND lr.source_type='sale' AND EXISTS(SELECT 1 FROM sales s WHERE s.id=lr.source_id AND s.seller_id=$4)) OR ($6 AND lr.source_type='quote' AND EXISTS(SELECT 1 FROM quotes q WHERE q.id=lr.source_id AND q.seller_id=$4)) OR ($7 AND lr.source_type='sale'))`, id, input.Message, input.FileURL, userID, access.Admin || access.ArtFinal, access.Sales, access.Purchases)
	if err != nil {
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return fmt.Errorf("acesso negado a solicitacao")
	}
	return nil
}

func (r *ArtFinalRepository) CreateLayoutRequest(input models.LayoutRequestInput, userID int, access models.ArtFinalAccess) (int64, error) {
	tx, err := r.db.Begin()
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()
	if err = ensureActiveArtFinalUser(tx, input.AssignedTo, "arte_final"); err != nil {
		return 0, err
	}
	var sourceSnapshot []byte
	var owner int
	if input.SourceType == "sale" {
		err = tx.QueryRow(`SELECT to_jsonb(s),s.seller_id FROM sales s WHERE s.id=$1 FOR SHARE`, input.SourceID).Scan(&sourceSnapshot, &owner)
	} else {
		err = tx.QueryRow(`SELECT to_jsonb(q),q.seller_id FROM quotes q WHERE q.id=$1 FOR SHARE`, input.SourceID).Scan(&sourceSnapshot, &owner)
	}
	if err != nil {
		return 0, err
	}
	if !(access.Admin || access.ArtFinal) && (!access.Sales || owner != userID) {
		return 0, fmt.Errorf("acesso negado ao documento de origem")
	}
	var id int64
	err = tx.QueryRow(`INSERT INTO layout_requests(source_type,source_id,title,instructions,file_mode,common_file_url,status,priority,due_at,assigned_to,requested_by,source_snapshot) VALUES($1,$2,$3,$4,$5,$6,'requested',$7,$8,$9,$10,$11::jsonb) RETURNING id`, input.SourceType, input.SourceID, input.Title, input.Instructions, input.FileMode, input.CommonFileURL, input.Priority, input.DueAt, input.AssignedTo, userID, sourceSnapshot).Scan(&id)
	if err != nil {
		return 0, err
	}
	for _, item := range input.Items {
		var snapshot []byte
		var parent int
		if item.EntityType == "sale_item" {
			err = tx.QueryRow(`SELECT to_jsonb(si)||jsonb_build_object('product',to_jsonb(p)),si.sale_id FROM sale_items si JOIN products p ON p.id=si.product_id WHERE si.id=$1`, item.ItemID).Scan(&snapshot, &parent)
		} else {
			err = tx.QueryRow(`SELECT to_jsonb(qi)||jsonb_build_object('product',to_jsonb(p)),qi.quote_id FROM quote_items qi JOIN products p ON p.id=qi.product_id WHERE qi.id=$1`, item.ItemID).Scan(&snapshot, &parent)
		}
		if err != nil {
			return 0, err
		}
		if parent != input.SourceID {
			return 0, fmt.Errorf("item %d nao pertence ao documento de origem", item.ItemID)
		}
		if _, err = tx.Exec(`INSERT INTO layout_request_items(request_id,entity_type,item_id,group_key,source_file_url,item_snapshot) VALUES($1,$2,$3,$4,$5,$6::jsonb)`, id, item.EntityType, item.ItemID, item.GroupKey, item.SourceFileURL, snapshot); err != nil {
			return 0, err
		}
	}
	groups := map[string]string{}
	for _, item := range input.Items {
		if item.GroupKey == "" {
			continue
		}
		file := item.SourceFileURL
		if input.FileMode == "common" {
			file = input.CommonFileURL
		}
		if previous, ok := groups[item.GroupKey]; ok && previous != file {
			return 0, fmt.Errorf("itens do grupo %s devem usar o mesmo arquivo", item.GroupKey)
		}
		groups[item.GroupKey] = file
	}
	for key, file := range groups {
		if _, err = tx.Exec(`INSERT INTO layout_request_groups(request_id,group_key,file_url) VALUES($1,$2,$3)`, id, key, file); err != nil {
			return 0, err
		}
	}
	if _, err = tx.Exec(`INSERT INTO art_final_audit_log(entity_type,entity_id,action,details,user_id) VALUES('layout_request',$1,'requested',jsonb_build_object('source_type',$2,'source_id',$3,'items',$4),$5)`, id, input.SourceType, input.SourceID, len(input.Items), userID); err != nil {
		return 0, err
	}
	if _, err = tx.Exec(`INSERT INTO notifications(recipient_permission,notification_type,message) VALUES('arte_final','layout_requested',$1)`, fmt.Sprintf("Nova solicitacao de layout: %s", input.Title)); err != nil {
		return 0, err
	}
	return id, tx.Commit()
}

func (r *ArtFinalRepository) TransitionLayoutRequest(id int64, status, note string, userID int) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	var current string
	if err = tx.QueryRow(`SELECT status FROM layout_requests WHERE id=$1 FOR UPDATE`, id).Scan(&current); err != nil {
		return err
	}
	if current == status {
		return tx.Commit()
	}
	allowed := map[string]map[string]bool{"draft": {"requested": true, "cancelled": true}, "requested": {"in_progress": true, "cancelled": true}, "in_progress": {"awaiting_approval": true, "cancelled": true}, "awaiting_approval": {"changes_requested": true, "cancelled": true}, "changes_requested": {"in_progress": true, "awaiting_approval": true, "cancelled": true}}
	if !allowed[current][status] {
		return fmt.Errorf("transicao de layout invalida: %s para %s", current, status)
	}
	if _, err = tx.Exec(`UPDATE layout_requests SET status=$1,updated_at=now() WHERE id=$2`, status, id); err != nil {
		return err
	}
	if status != "cancelled" {
		_, err = tx.Exec(`UPDATE layout_request_items SET status=$1,updated_at=now() WHERE request_id=$2 AND status<>'approved'`, status, id)
		if err != nil {
			return err
		}
	}
	_, err = tx.Exec(`INSERT INTO art_final_audit_log(entity_type,entity_id,action,details,user_id) VALUES('layout_request',$1,'transition',jsonb_build_object('before',$2,'after',$3,'note',$4),$5)`, id, current, status, note, userID)
	if err != nil {
		return err
	}
	var sellerID int
	if err = tx.QueryRow(`SELECT CASE WHEN r.source_type='sale' THEN (SELECT seller_id FROM sales WHERE id=r.source_id) ELSE (SELECT seller_id FROM quotes WHERE id=r.source_id) END FROM layout_requests r WHERE r.id=$1`, id).Scan(&sellerID); err != nil {
		return err
	}
	if _, err = tx.Exec(`INSERT INTO notifications(recipient_user_id,notification_type,message) VALUES($1,'layout_transition',$2)`, sellerID, fmt.Sprintf("Layout atualizado para %s", status)); err != nil {
		return err
	}
	return tx.Commit()
}

func (r *ArtFinalRepository) AddLayoutVersion(requestItemID int64, input models.LayoutVersionInput, userID int) (int64, error) {
	tx, err := r.db.Begin()
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()
	var entity string
	var itemID int
	var snapshot []byte
	var requestID int64
	if err = tx.QueryRow(`SELECT entity_type,item_id,item_snapshot,request_id FROM layout_request_items WHERE id=$1 FOR UPDATE`, requestItemID).Scan(&entity, &itemID, &snapshot, &requestID); err != nil {
		return 0, err
	}
	if _, err = tx.Exec(`SELECT pg_advisory_xact_lock(hashtext($1),$2)`, entity, itemID); err != nil {
		return 0, err
	}
	var version int
	if err = tx.QueryRow(`SELECT COALESCE(MAX(version),0)+1 FROM item_layout_versions WHERE entity_type=$1 AND item_id=$2`, entity, itemID).Scan(&version); err != nil {
		return 0, err
	}
	var id int64
	if err = tx.QueryRow(`INSERT INTO item_layout_versions(entity_type,item_id,version,label,file_url,item_snapshot,approval_status,created_by,request_item_id) VALUES($1,$2,$3,$4,$5,$6::jsonb,'pending',$7,$8) RETURNING id`, entity, itemID, version, input.Label, input.FileURL, snapshot, userID, requestItemID).Scan(&id); err != nil {
		return 0, err
	}
	_, err = tx.Exec(`UPDATE layout_request_items SET status='awaiting_approval',updated_at=now() WHERE id=$1; UPDATE layout_requests SET status='awaiting_approval',updated_at=now() WHERE id=$2`, requestItemID, requestID)
	if err != nil {
		return 0, err
	}
	_, err = tx.Exec(`INSERT INTO art_final_audit_log(entity_type,entity_id,action,details,user_id) VALUES('layout_version',$1,'created',jsonb_build_object('request_id',$2,'request_item_id',$3,'version',$4),$5)`, id, requestID, requestItemID, version, userID)
	if err != nil {
		return 0, err
	}
	var sellerID int
	if err = tx.QueryRow(`SELECT CASE WHEN r.source_type='sale' THEN (SELECT seller_id FROM sales WHERE id=r.source_id) ELSE (SELECT seller_id FROM quotes WHERE id=r.source_id) END FROM layout_requests r WHERE r.id=$1`, requestID).Scan(&sellerID); err != nil {
		return 0, err
	}
	if _, err = tx.Exec(`INSERT INTO notifications(recipient_user_id,notification_type,message) VALUES($1,'layout_version_pending',$2)`, sellerID, fmt.Sprintf("Nova versao %d de layout aguardando aprovacao", version)); err != nil {
		return 0, err
	}
	return id, tx.Commit()
}

func (r *ArtFinalRepository) DecideLayoutVersion(versionID int64, status, note string, userID int, access models.ArtFinalAccess) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	var itemID, latest, requestID int64
	var current string
	if err = tx.QueryRow(`SELECT v.request_item_id,v.approval_status,i.request_id FROM item_layout_versions v JOIN layout_request_items i ON i.id=v.request_item_id WHERE v.id=$1 FOR UPDATE`, versionID).Scan(&itemID, &current, &requestID); err != nil {
		return err
	}
	var allowed bool
	if err = tx.QueryRow(`SELECT $2 OR ($3 AND (r.source_type='sale' AND EXISTS(SELECT 1 FROM sales s WHERE s.id=r.source_id AND s.seller_id=$4) OR r.source_type='quote' AND EXISTS(SELECT 1 FROM quotes q WHERE q.id=r.source_id AND q.seller_id=$4))) FROM layout_requests r WHERE r.id=$1`, requestID, access.Admin || access.ArtFinal, access.Sales, userID).Scan(&allowed); err != nil {
		return err
	}
	if !allowed {
		return sql.ErrNoRows
	}
	if err = tx.QueryRow(`SELECT id FROM item_layout_versions WHERE request_item_id=$1 ORDER BY version DESC LIMIT 1 FOR UPDATE`, itemID).Scan(&latest); err != nil {
		return err
	}
	if latest != versionID {
		return fmt.Errorf("somente a versao vigente pode receber retorno")
	}
	if current == status {
		return tx.Commit()
	}
	nextItem := "changes_requested"
	if status == "approved" {
		nextItem = "approved"
	}
	_, err = tx.Exec(`UPDATE item_layout_versions SET approval_status=$1,approval_note=$2,approved_by=$3,approved_at=now() WHERE id=$4; UPDATE layout_request_items SET status=$5,updated_at=now() WHERE id=$6`, status, note, userID, versionID, nextItem, itemID)
	if err != nil {
		return err
	}
	var allApproved bool
	if err = tx.QueryRow(`SELECT bool_and(status='approved') FROM layout_request_items WHERE request_id=$1`, requestID).Scan(&allApproved); err != nil {
		return err
	}
	nextRequest := nextItem
	if allApproved {
		nextRequest = "approved"
	}
	if _, err = tx.Exec(`UPDATE layout_requests SET status=$1,updated_at=now() WHERE id=$2`, nextRequest, requestID); err != nil {
		return err
	}
	_, err = tx.Exec(`INSERT INTO layout_approval_events(layout_version_id,status,note,created_by) VALUES($1,$2,$3,$4)`, versionID, status, note, userID)
	if err == nil {
		_, err = tx.Exec(`INSERT INTO art_final_audit_log(entity_type,entity_id,action,details,user_id) VALUES('layout_version',$1,'decision',jsonb_build_object('request_id',$2,'request_item_id',$3,'before',$4,'after',$5,'note',$6),$7)`, versionID, requestID, itemID, current, status, note, userID)
	}
	if err != nil {
		return err
	}
	var sellerID int
	if err = tx.QueryRow(`SELECT CASE WHEN r.source_type='sale' THEN (SELECT seller_id FROM sales WHERE id=r.source_id) ELSE (SELECT seller_id FROM quotes WHERE id=r.source_id) END FROM layout_requests r WHERE r.id=$1`, requestID).Scan(&sellerID); err != nil {
		return err
	}
	if _, err = tx.Exec(`INSERT INTO notifications(recipient_user_id,notification_type,message) VALUES($1,'layout_decision',$2)`, sellerID, fmt.Sprintf("Decisao do layout: %s", status)); err != nil {
		return err
	}
	if _, err = tx.Exec(`INSERT INTO notifications(recipient_permission,notification_type,message) VALUES('arte_final','layout_decision',$1)`, fmt.Sprintf("Versao %d recebeu decisao %s", versionID, status)); err != nil {
		return err
	}
	return tx.Commit()
}

func (r *ArtFinalRepository) UpsertLayoutJob(kind string, requestItemID int64, input models.LayoutJobInput, userID int, access models.ArtFinalAccess) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	jobPermissions := []string{"arte_final", "producao"}
	if kind == "corel" {
		jobPermissions = []string{"arte_final", "compras"}
	}
	if err = ensureActiveArtFinalUser(tx, input.ResponsibleID, jobPermissions...); err != nil {
		return err
	}
	if _, err = tx.Exec(`SELECT pg_advisory_xact_lock(hashtext($1),$2)`, "layout_job_"+kind, requestItemID); err != nil {
		return err
	}
	var allowed, received bool
	if err = tx.QueryRow(`SELECT ($2 OR ($3='corel' AND $4 AND r.source_type='sale' AND EXISTS(SELECT 1 FROM purchase_orders po WHERE po.sale_id=r.source_id AND (po.buyer_id=$5 OR po.buyer_id IS NULL))) OR ($3='engraving' AND $6)),i.product_received_at IS NOT NULL FROM layout_request_items i JOIN layout_requests r ON r.id=i.request_id WHERE i.id=$1`, requestItemID, access.Admin || access.ArtFinal, kind, access.Purchases, userID, access.ProductionProfile).Scan(&allowed, &received); err != nil {
		return err
	}
	if !allowed {
		return sql.ErrNoRows
	}
	table := "layout_engraving_jobs"
	if kind == "corel" {
		table = "layout_corel_jobs"
	}
	var current sql.NullString
	scanErr := tx.QueryRow(`SELECT status FROM `+table+` WHERE request_item_id=$1 FOR UPDATE`, requestItemID).Scan(&current)
	if scanErr != nil && scanErr != sql.ErrNoRows {
		return scanErr
	}
	if current.Valid && current.String == input.Status {
		return tx.Commit()
	}
	states := map[string]map[string]bool{"corel:new": {"pending": true}, "corel:pending": {"sent": true, "cancelled": true}, "corel:sent": {"received": true, "cancelled": true}, "engraving:new": {"pending": true}, "engraving:pending": {"in_progress": true, "cancelled": true}, "engraving:in_progress": {"ready": true, "cancelled": true}, "engraving:ready": {"approved": true, "in_progress": true, "cancelled": true}}
	stateKey := kind + ":new"
	if current.Valid {
		stateKey = kind + ":" + current.String
	}
	if !states[stateKey][input.Status] {
		return fmt.Errorf("transicao de %s invalida: %s para %s", kind, map[bool]string{true: current.String, false: "new"}[current.Valid], input.Status)
	}
	if kind == "engraving" && current.Valid && current.String == "ready" && input.Status == "in_progress" && input.Reason == "" {
		return fmt.Errorf("motivo obrigatorio para correcao da gravacao")
	}
	if kind == "engraving" && input.Status != "pending" && !received {
		return fmt.Errorf("produto ainda nao foi confirmado fisicamente na empresa")
	}
	var version int
	if err = tx.QueryRow(`SELECT COALESCE(MAX(version),0)+1 FROM layout_job_versions WHERE kind=$1 AND request_item_id=$2`, kind, requestItemID).Scan(&version); err != nil {
		return err
	}
	if kind == "corel" {
		_, err = tx.Exec(`INSERT INTO layout_corel_jobs(request_item_id,status,responsible_id,due_at,file_url,tags,reason,updated_by,external_contact,sent_at,received_at) VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,CASE WHEN $2='sent' THEN now() END,CASE WHEN $2='received' THEN now() END) ON CONFLICT(request_item_id) DO UPDATE SET status=EXCLUDED.status,responsible_id=EXCLUDED.responsible_id,due_at=EXCLUDED.due_at,file_url=EXCLUDED.file_url,tags=EXCLUDED.tags,reason=EXCLUDED.reason,updated_by=EXCLUDED.updated_by,external_contact=EXCLUDED.external_contact,sent_at=COALESCE(layout_corel_jobs.sent_at,EXCLUDED.sent_at),received_at=COALESCE(layout_corel_jobs.received_at,EXCLUDED.received_at),updated_at=now()`, requestItemID, input.Status, input.ResponsibleID, input.DueAt, input.FileURL, pq.Array(input.Tags), input.Reason, userID, input.ExternalContact)
	} else {
		_, err = tx.Exec(`INSERT INTO layout_engraving_jobs(request_item_id,status,responsible_id,due_at,file_url,tags,reason,updated_by,started_at,completed_at) VALUES($1,$2,$3,$4,$5,$6,$7,$8,CASE WHEN $2='in_progress' THEN now() END,CASE WHEN $2 IN ('ready','approved') THEN now() END) ON CONFLICT(request_item_id) DO UPDATE SET status=EXCLUDED.status,responsible_id=EXCLUDED.responsible_id,due_at=EXCLUDED.due_at,file_url=EXCLUDED.file_url,tags=EXCLUDED.tags,reason=EXCLUDED.reason,updated_by=EXCLUDED.updated_by,started_at=COALESCE(layout_engraving_jobs.started_at,EXCLUDED.started_at),completed_at=COALESCE(layout_engraving_jobs.completed_at,EXCLUDED.completed_at),updated_at=now()`, requestItemID, input.Status, input.ResponsibleID, input.DueAt, input.FileURL, pq.Array(input.Tags), input.Reason, userID)
	}
	if err != nil {
		return err
	}
	if _, err = tx.Exec(`INSERT INTO layout_job_versions(kind,request_item_id,version,status,responsible_id,external_contact,due_at,file_url,tags,reason,created_by) VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)`, kind, requestItemID, version, input.Status, input.ResponsibleID, input.ExternalContact, input.DueAt, input.FileURL, pq.Array(input.Tags), input.Reason, userID); err != nil {
		return err
	}
	var requestID int64
	if err = tx.QueryRow(`SELECT request_id FROM layout_request_items WHERE id=$1`, requestItemID).Scan(&requestID); err != nil {
		return err
	}
	_, err = tx.Exec(`INSERT INTO art_final_audit_log(entity_type,entity_id,action,details,user_id) VALUES($1,$2,'upsert',jsonb_build_object('request_id',$3,'request_item_id',$2,'status',$4,'file_url',$5,'reason',$6),$7)`, kind, requestItemID, requestID, input.Status, input.FileURL, input.Reason, userID)
	if err != nil {
		return err
	}
	destination := "arte_final"
	if kind == "corel" && (input.Status == "sent" || input.Status == "received") {
		destination = "compras"
	} else if kind == "engraving" && (input.Status == "ready" || input.Status == "approved") {
		destination = "producao"
	}
	if _, err = tx.Exec(`INSERT INTO notifications(recipient_permission,notification_type,message) VALUES($1,$2,$3)`, destination, "layout_"+kind, fmt.Sprintf("%s do item %d atualizado para %s", kind, requestItemID, input.Status)); err != nil {
		return err
	}
	return tx.Commit()
}

func (r *ArtFinalRepository) ConfirmProductReceived(requestItemID int64, received bool, userID int, access models.ArtFinalAccess) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	var entity string
	var itemID int
	var receivedAt sql.NullTime
	err = tx.QueryRow(`SELECT i.entity_type,i.item_id,i.product_received_at FROM layout_request_items i JOIN layout_requests r ON r.id=i.request_id WHERE i.id=$1 AND ($2 OR $3 OR ($4 AND r.source_type='sale' AND EXISTS(SELECT 1 FROM purchase_orders po WHERE po.sale_id=r.source_id AND (po.buyer_id=$5 OR po.buyer_id IS NULL)))) FOR UPDATE`, requestItemID, access.Admin, access.ProductionProfile, access.Purchases, userID).Scan(&entity, &itemID, &receivedAt)
	if err != nil {
		return err
	}
	if entity != "sale_item" {
		return fmt.Errorf("presenca fisica disponivel apenas para item de pedido")
	}
	if receivedAt.Valid == received {
		return tx.Commit()
	}
	if _, err = tx.Exec(`UPDATE layout_request_items SET product_received_at=CASE WHEN $1 THEN COALESCE(product_received_at,now()) ELSE NULL END,product_received_by=CASE WHEN $1 THEN $2 ELSE NULL END,updated_at=now() WHERE id=$3`, received, userID, requestItemID); err != nil {
		return err
	}
	if _, err = tx.Exec(`UPDATE sale_items SET product_received_at=CASE WHEN $1 THEN COALESCE(product_received_at,now()) ELSE NULL END,product_received_by=CASE WHEN $1 THEN $2 ELSE NULL END WHERE id=$3`, received, userID, itemID); err != nil {
		return err
	}
	if _, err = tx.Exec(`INSERT INTO art_final_audit_log(entity_type,entity_id,action,details,user_id) VALUES('layout_item',$1,'product_received',jsonb_build_object('received',$2),$3)`, requestItemID, received, userID); err != nil {
		return err
	}
	if _, err = tx.Exec(`INSERT INTO notifications(recipient_permission,notification_type,message) VALUES('producao','product_received',$1)`, fmt.Sprintf("Presenca fisica do item %d: %t", itemID, received)); err != nil {
		return err
	}
	return tx.Commit()
}

func (r *ArtFinalRepository) UpdateStoryLifecycle(id int64, input models.StoryLifecycleInput, userID int) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	var current, title string
	if err = tx.QueryRow(`SELECT lifecycle_status,title FROM art_final_stories WHERE id=$1 FOR UPDATE`, id).Scan(&current, &title); err != nil {
		return err
	}
	if current == input.Status {
		return tx.Commit()
	}
	allowed := map[string]map[string]bool{"draft": {"returned": true, "published": true}, "returned": {"corrected": true}, "corrected": {"returned": true, "published": true}, "published": {"returned": true, "expired": true}, "expired": {"corrected": true}}
	if !allowed[current][input.Status] {
		return fmt.Errorf("transicao de Story invalida: %s para %s", current, input.Status)
	}
	_, err = tx.Exec(`UPDATE art_final_stories SET lifecycle_status=$1,observation=$2,expires_at=$3,published_at=CASE WHEN $1='published' THEN COALESCE(published_at,now()) ELSE published_at END,checked=($1='published'),checked_at=CASE WHEN $1='published' THEN COALESCE(checked_at,now()) ELSE checked_at END,updated_by=$4,updated_at=now() WHERE id=$5`, input.Status, input.Observation, input.ExpiresAt, userID, id)
	if err != nil {
		return err
	}
	_, err = tx.Exec(`INSERT INTO art_final_audit_log(entity_type,entity_id,action,details,user_id) VALUES('story',$1,'lifecycle',jsonb_build_object('before',$2,'after',$3,'observation',$4),$5)`, id, current, input.Status, input.Observation, userID)
	if err != nil {
		return err
	}
	if _, err = tx.Exec(`INSERT INTO notifications(recipient_permission,notification_type,message) VALUES('marketing','story_lifecycle',$1)`, fmt.Sprintf("Story %s atualizado para %s", title, input.Status)); err != nil {
		return err
	}
	return tx.Commit()
}

func (r *ArtFinalRepository) CreateTask(input models.ArtFinalTaskInput, userID int) (int64, error) {
	tx, err := r.db.Begin()
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()
	taskPermissions := []string{"arte_final"}
	if input.Panel == "media" {
		taskPermissions = []string{"marketing"}
	}
	if err = ensureActiveArtFinalUser(tx, input.AssignedTo, taskPermissions...); err != nil {
		return 0, err
	}
	var id int64
	err = tx.QueryRow(`INSERT INTO art_final_tasks(panel,category,title,description,sale_id,purchase_id,due_date,due_at,assigned_to,channel,format,priority,tags,attachment_url,status,created_by)
		VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16) RETURNING id`, input.Panel, input.Category, input.Title, input.Description, input.SaleID, input.PurchaseID, input.DueDate, input.DueAt, input.AssignedTo, input.Channel, input.Format, input.Priority, pq.Array(input.Tags), input.AttachmentURL, input.Status, userID).Scan(&id)
	if err != nil {
		return 0, err
	}
	if _, err = tx.Exec(`INSERT INTO art_final_audit_log(entity_type,entity_id,action,details,user_id) VALUES('task',$1,'created',jsonb_build_object('category',$2,'title',$3),$4)`, id, input.Category, input.Title, userID); err != nil {
		return 0, err
	}
	destination := "arte_final"
	if input.Panel == "media" {
		destination = "marketing"
	}
	if _, err = tx.Exec(`INSERT INTO notifications(recipient_permission,notification_type,message) VALUES($1,'art_final_task',$2)`, destination, fmt.Sprintf("Nova tarefa de arte-final: %s", input.Title)); err != nil {
		return 0, err
	}
	return id, tx.Commit()
}

func (r *ArtFinalRepository) UpdateTask(id int64, input models.ArtFinalTaskInput, userID int, access models.ArtFinalAccess) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	taskPermissions := []string{"arte_final"}
	if input.Panel == "media" {
		taskPermissions = []string{"marketing"}
	}
	if err = ensureActiveArtFinalUser(tx, input.AssignedTo, taskPermissions...); err != nil {
		return err
	}
	var before, after []byte
	var persistedPanel string
	if err = tx.QueryRow(`SELECT to_jsonb(t),panel FROM art_final_tasks t WHERE id=$1 FOR UPDATE`, id).Scan(&before, &persistedPanel); err != nil {
		return err
	}
	if !access.Admin && persistedPanel != input.Panel {
		return sql.ErrNoRows
	}
	if persistedPanel == "media" && !access.Admin && !access.Marketing {
		return sql.ErrNoRows
	}
	if persistedPanel != "media" && !access.Admin && !access.Manage {
		return sql.ErrNoRows
	}
	err = tx.QueryRow(`UPDATE art_final_tasks SET panel=$1,category=$2,title=$3,description=$4,sale_id=$5,purchase_id=$6,due_date=$7,due_at=$8,assigned_to=$9,channel=$10,format=$11,priority=$12,tags=$13,attachment_url=$14,status=$15,
		completed_by=CASE WHEN $15='done' THEN $16 ELSE NULL END,completed_at=CASE WHEN $15='done' THEN COALESCE(completed_at,now()) ELSE NULL END,updated_at=now()
		WHERE id=$17 AND (panel IS DISTINCT FROM $1 OR category IS DISTINCT FROM $2 OR title IS DISTINCT FROM $3 OR description IS DISTINCT FROM $4
		 OR sale_id IS DISTINCT FROM $5 OR purchase_id IS DISTINCT FROM $6 OR due_date IS DISTINCT FROM $7 OR due_at IS DISTINCT FROM $8 OR assigned_to IS DISTINCT FROM $9 OR channel IS DISTINCT FROM $10 OR format IS DISTINCT FROM $11 OR priority IS DISTINCT FROM $12
		 OR tags IS DISTINCT FROM $13 OR attachment_url IS DISTINCT FROM $14 OR status IS DISTINCT FROM $15)
		RETURNING to_jsonb(art_final_tasks)`, input.Panel, input.Category, input.Title, input.Description, input.SaleID, input.PurchaseID, input.DueDate, input.DueAt, input.AssignedTo, input.Channel, input.Format, input.Priority, pq.Array(input.Tags), input.AttachmentURL, input.Status, userID, id).Scan(&after)
	if err == sql.ErrNoRows {
		return tx.Commit()
	}
	if err != nil {
		return err
	}
	if _, err = tx.Exec(`INSERT INTO art_final_audit_log(entity_type,entity_id,action,details,user_id) VALUES('task',$1,'updated',jsonb_build_object('before',$2::jsonb,'after',$3::jsonb),$4)`, id, before, after, userID); err != nil {
		return err
	}
	destination := "arte_final"
	if persistedPanel == "media" {
		destination = "marketing"
	}
	if _, err = tx.Exec(`INSERT INTO notifications(recipient_permission,notification_type,message) VALUES($1,'art_final_task_updated',$2)`, destination, fmt.Sprintf("Tarefa de arte-final atualizada: %s", input.Title)); err != nil {
		return err
	}
	return tx.Commit()
}

func (r *ArtFinalRepository) CreateStory(input models.ArtFinalStoryInput, userID int) (int64, error) {
	tx, err := r.db.Begin()
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()
	if _, err = tx.Exec(`SELECT pg_advisory_xact_lock(hashtext('art_final_story'),$1::int)`, input.SaleID); err != nil {
		return 0, err
	}
	if input.SaleItemID != nil {
		var ok bool
		if err = tx.QueryRow(`SELECT EXISTS(SELECT 1 FROM sale_items WHERE id=$1 AND sale_id=$2)`, *input.SaleItemID, input.SaleID).Scan(&ok); err != nil || !ok {
			if err != nil {
				return 0, err
			}
			return 0, fmt.Errorf("item nao pertence ao pedido")
		}
	}
	var id int64
	err = tx.QueryRow(`SELECT id FROM art_final_stories WHERE sale_id=$1 AND sale_item_id IS NOT DISTINCT FROM $2`, input.SaleID, input.SaleItemID).Scan(&id)
	if err == nil {
		return id, tx.Commit()
	}
	if err != sql.ErrNoRows {
		return 0, err
	}
	err = tx.QueryRow(`INSERT INTO art_final_stories(sale_id,sale_item_id,title,tags,created_by) VALUES($1,$2,$3,$4,$5) RETURNING id`, input.SaleID, input.SaleItemID, input.Title, pq.Array(input.Tags), userID).Scan(&id)
	if err != nil {
		return 0, err
	}
	if _, err = tx.Exec(`INSERT INTO art_final_audit_log(entity_type,entity_id,action,details,user_id) VALUES('story',$1,'created',jsonb_build_object('sale_id',$2),$3)`, id, input.SaleID, userID); err != nil {
		return 0, err
	}
	if _, err = tx.Exec(`INSERT INTO notifications(recipient_permission,notification_type,message) VALUES('marketing','art_final_story_created',$1)`, fmt.Sprintf("Novo Story: %s", input.Title)); err != nil {
		return 0, err
	}
	return id, tx.Commit()
}

func (r *ArtFinalRepository) CheckStory(id int64, checked bool, userID int) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	var title, lifecycle string
	var current bool
	if err = tx.QueryRow(`SELECT title,checked,lifecycle_status FROM art_final_stories WHERE id=$1 FOR UPDATE`, id).Scan(&title, &current, &lifecycle); err != nil {
		return err
	}
	if current == checked {
		return tx.Commit()
	}
	if checked && (lifecycle != "draft" && lifecycle != "corrected") {
		return fmt.Errorf("Story somente pode ser publicado quando novo ou corrigido")
	}
	if !checked && lifecycle != "published" {
		return fmt.Errorf("somente Story publicado pode ser devolvido")
	}
	next := "returned"
	observation := "Devolvido pelo controle de publicacao"
	if checked {
		next = "published"
		observation = ""
	}
	if _, err = tx.Exec(`UPDATE art_final_stories SET checked=$1,checked_by=CASE WHEN $1 THEN $2 ELSE NULL END,checked_at=CASE WHEN $1 THEN now() ELSE NULL END,lifecycle_status=$4,observation=$5,published_at=CASE WHEN $1 THEN COALESCE(published_at,now()) ELSE published_at END,updated_by=$2,updated_at=now() WHERE id=$3`, checked, userID, id, next, observation); err != nil {
		return err
	}
	if _, err = tx.Exec(`INSERT INTO art_final_audit_log(entity_type,entity_id,action,details,user_id) VALUES('story',$1,'checked',jsonb_build_object('before',jsonb_build_object('checked',$2),'after',jsonb_build_object('checked',$3)),$4)`, id, current, checked, userID); err != nil {
		return err
	}
	if _, err = tx.Exec(`INSERT INTO notifications(recipient_permission,notification_type,message) VALUES('marketing','art_final_story_checked',$1)`, fmt.Sprintf("Story %s: %s", map[bool]string{true: "concluido", false: "reaberto"}[checked], title)); err != nil {
		return err
	}
	return tx.Commit()
}

func (r *ArtFinalRepository) DeleteStory(id int64, userID int) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	var before []byte
	var title string
	if err = tx.QueryRow(`SELECT to_jsonb(st),title FROM art_final_stories st WHERE id=$1 FOR UPDATE`, id).Scan(&before, &title); err != nil {
		return err
	}
	if _, err = tx.Exec(`DELETE FROM art_final_stories WHERE id=$1`, id); err != nil {
		return err
	}
	if _, err = tx.Exec(`INSERT INTO art_final_audit_log(entity_type,entity_id,action,details,user_id) VALUES('story',$1,'deleted',jsonb_build_object('before',$2::jsonb),$3)`, id, before, userID); err != nil {
		return err
	}
	if _, err = tx.Exec(`INSERT INTO notifications(recipient_permission,notification_type,message) VALUES('marketing','art_final_story_deleted',$1)`, fmt.Sprintf("Story removido: %s", title)); err != nil {
		return err
	}
	return tx.Commit()
}
