package repositories

import (
	"database/sql"
	apperrors "donapresentes/errors"
	"time"
)

type DashboardRepository struct {
	db *sql.DB
}

func NewDashboardRepository(db *sql.DB) *DashboardRepository {
	return &DashboardRepository{db: db}
}

func (r *DashboardRepository) CountCustomers() (int, error) {
	var total int
	if err := r.db.QueryRow(`SELECT COUNT(*) FROM customers`).Scan(&total); err != nil {
		return 0, apperrors.NewDatabaseError(err)
	}
	return total, nil
}

func (r *DashboardRepository) CountProducts() (int, error) {
	var total int
	if err := r.db.QueryRow(`SELECT COUNT(*) FROM products`).Scan(&total); err != nil {
		return 0, apperrors.NewDatabaseError(err)
	}
	return total, nil
}

func (r *DashboardRepository) CountQuotesBySeller(sellerID int) (int, error) {
	var total int
	if err := r.db.QueryRow(`SELECT COUNT(*) FROM quotes WHERE seller_id=$1`, sellerID).Scan(&total); err != nil {
		return 0, apperrors.NewDatabaseError(err)
	}
	return total, nil
}

func (r *DashboardRepository) GetQuoteSituationSummary(sellerID int) (int, int, int, int, error) {
	var total, overdue, soon, withoutDate int
	err := r.db.QueryRow(
		`SELECT
			COUNT(*) AS total,
			COALESCE(SUM(CASE WHEN quote_valid_until < NOW() THEN 1 ELSE 0 END), 0) AS overdue,
			COALESCE(SUM(CASE WHEN quote_valid_until BETWEEN NOW() AND NOW() + interval '7 days' THEN 1 ELSE 0 END), 0) AS soon,
			COALESCE(SUM(CASE WHEN quote_valid_until IS NULL THEN 1 ELSE 0 END), 0) AS without_date
		FROM quotes WHERE seller_id=$1`, sellerID).
		Scan(&total, &overdue, &soon, &withoutDate)
	if err != nil {
		return 0, 0, 0, 0, apperrors.NewDatabaseError(err)
	}
	return total, overdue, soon, withoutDate, nil
}

type QuoteFeedbackAlert struct {
	OrcamentoID         int
	Cliente             string
	QuoteValidUntil     sql.NullTime
	FeedbackDateTime    *time.Time
	FeedbackObservation string
	Observacao          string
}

func (r *DashboardRepository) ListQuoteFeedbackAlerts(sellerID int) ([]QuoteFeedbackAlert, error) {
	rows, err := r.db.Query(`SELECT q.id, COALESCE(c.name, ''), q.quote_valid_until, q.feedback_datetime, COALESCE(q.feedback_observation, ''), COALESCE(q.observations, '')
		FROM quotes q
		LEFT JOIN customers c ON c.id=q.customer_id
		WHERE q.seller_id=$1 AND (q.quote_valid_until IS NULL OR q.quote_valid_until <= NOW() + interval '7 days')
		ORDER BY q.quote_valid_until NULLS LAST LIMIT 20`, sellerID)
	if err != nil {
		return nil, apperrors.NewDatabaseError(err)
	}
	defer rows.Close()

	var alerts []QuoteFeedbackAlert
	for rows.Next() {
		var alert QuoteFeedbackAlert
		var validUntil sql.NullTime
		var feedbackDatetime sql.NullTime
		if err := rows.Scan(&alert.OrcamentoID, &alert.Cliente, &validUntil, &feedbackDatetime, &alert.FeedbackObservation, &alert.Observacao); err != nil {
			return nil, err
		}
		alert.QuoteValidUntil = validUntil
		if feedbackDatetime.Valid {
			alert.FeedbackDateTime = &feedbackDatetime.Time
		}
		alerts = append(alerts, alert)
	}
	return alerts, nil
}

type SaleSituationRow struct {
	SaleID        int
	Status        string
	DeliveryDate  sql.NullTime
	DepartureDate sql.NullTime
	ArrivalDate   sql.NullTime
}

func (r *DashboardRepository) ListSalesSituacao(sellerID int) ([]SaleSituationRow, error) {
	rows, err := r.db.Query(`SELECT id, status, delivery_date, departure_date, arrival_date
		FROM sales
		WHERE seller_id=$1 AND status IN ('Em Andamento', 'Remessa', 'Amostra', 'Reposição de Pedido')
		ORDER BY created_at DESC LIMIT 30`, sellerID)
	if err != nil {
		return nil, apperrors.NewDatabaseError(err)
	}
	defer rows.Close()

	var sales []SaleSituationRow
	for rows.Next() {
		var row SaleSituationRow
		if err := rows.Scan(&row.SaleID, &row.Status, &row.DeliveryDate, &row.DepartureDate, &row.ArrivalDate); err != nil {
			return nil, err
		}
		sales = append(sales, row)
	}
	return sales, nil
}

func (r *DashboardRepository) CountSalesThisMonth(sellerID int) (int, int, float64, error) {
	var totalMonth, completedMonth int
	var commission float64
	err := r.db.QueryRow(`SELECT
		COUNT(*) FILTER (WHERE date_part('year', created_at)=date_part('year', NOW()) AND date_part('month', created_at)=date_part('month', NOW())) AS total_month,
		COUNT(*) FILTER (WHERE status='Concretizado' AND date_part('year', created_at)=date_part('year', NOW()) AND date_part('month', created_at)=date_part('month', NOW())) AS completed_month,
		COALESCE(SUM(total_value) FILTER (WHERE status='Concretizado' AND date_part('year', created_at)=date_part('year', NOW()) AND date_part('month', created_at)=date_part('month', NOW())),0) * 0.05 AS commission
		FROM sales WHERE seller_id=$1`, sellerID).Scan(&totalMonth, &completedMonth, &commission)
	if err != nil {
		return 0, 0, 0, apperrors.NewDatabaseError(err)
	}
	return totalMonth, completedMonth, commission, nil
}
