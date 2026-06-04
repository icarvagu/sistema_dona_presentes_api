package services

import (
	"database/sql"
	"os"
	"strconv"
	"time"

	"donapresentes/repositories"
)

type DashboardFeedbackAlert struct {
	OrcamentoID         int        `json:"orcamento_id"`
	Cliente             string     `json:"cliente"`
	QuoteValidUntil     *time.Time `json:"quote_valid_until,omitempty"`
	FeedbackDateTime    *time.Time `json:"feedback_datetime,omitempty"`
	FeedbackObservation string     `json:"feedback_observation,omitempty"`
	Observacao          string     `json:"observacao"`
}

type DashboardPedidoSituacao struct {
	PedidoID     int        `json:"pedido_id"`
	Status       string     `json:"status"`
	Problema     string     `json:"problema"`
	DeliveryDate *time.Time `json:"delivery_date,omitempty"`
}

type DashboardAlertaIntervencao struct {
	PedidoID int    `json:"pedido_id"`
	Motivo   string `json:"motivo"`
	Status   string `json:"status"`
}

type DashboardResponse struct {
	Cadastros struct {
		Clientes int `json:"clientes"`
		Produtos int `json:"produtos"`
	} `json:"cadastros"`
	Orcamentos struct {
		Total           int                      `json:"total"`
		PorSituacao     map[string]int           `json:"por_situacao"`
		AlertasFeedback []DashboardFeedbackAlert `json:"alertas_feedback"`
	} `json:"orcamentos"`
	Vendas struct {
		TotalMes           int                          `json:"total_mes"`
		MetaMes            int                          `json:"meta_mes"`
		FaltamParaMeta     int                          `json:"faltam_para_meta"`
		PedidosFinalizados int                          `json:"pedidos_finalizados"`
		Comissoes          float64                      `json:"comissoes"`
		SituacaoPedidos    []DashboardPedidoSituacao    `json:"situacao_pedidos"`
		AlertasIntervencao []DashboardAlertaIntervencao `json:"alertas_intervencao"`
	} `json:"vendas"`
}

type DashboardService struct {
	repo        *repositories.DashboardRepository
	monthlyGoal int
}

func NewDashboardService(repo *repositories.DashboardRepository) *DashboardService {
	goal := 20
	if raw := os.Getenv("DASHBOARD_MONTHLY_GOAL"); raw != "" {
		if parsed, err := strconv.Atoi(raw); err == nil && parsed > 0 {
			goal = parsed
		}
	}
	return &DashboardService{repo: repo, monthlyGoal: goal}
}

func (s *DashboardService) GetDashboardData(sellerID int) (DashboardResponse, error) {
	var resp DashboardResponse

	clientes, err := s.repo.CountCustomers()
	if err != nil {
		return resp, err
	}
	produtos, err := s.repo.CountProducts()
	if err != nil {
		return resp, err
	}
	resp.Cadastros.Clientes = clientes
	resp.Cadastros.Produtos = produtos

	quoteTotal, overdue, soon, withoutDate, err := s.repo.GetQuoteSituationSummary(sellerID)
	if err != nil {
		return resp, err
	}
	resp.Orcamentos.Total = quoteTotal
	resp.Orcamentos.PorSituacao = map[string]int{
		"vencidos": overdue,
		"proximos": soon,
		"sem_data": withoutDate,
		"abertos":  quoteTotal - overdue - soon - withoutDate,
	}
	feedbackAlerts, err := s.repo.ListQuoteFeedbackAlerts(sellerID)
	if err != nil {
		return resp, err
	}
	for _, alert := range feedbackAlerts {
		var quoteUntil *time.Time
		if alert.QuoteValidUntil.Valid {
			quoteUntil = &alert.QuoteValidUntil.Time
		}
		resp.Orcamentos.AlertasFeedback = append(resp.Orcamentos.AlertasFeedback, DashboardFeedbackAlert{
			OrcamentoID:         alert.OrcamentoID,
			Cliente:             alert.Cliente,
			QuoteValidUntil:     quoteUntil,
			FeedbackDateTime:    alert.FeedbackDateTime,
			FeedbackObservation: alert.FeedbackObservation,
			Observacao:          alert.Observacao,
		})
	}

	totalMonth, completedMonth, commission, err := s.repo.CountSalesThisMonth(sellerID)
	if err != nil {
		return resp, err
	}
	resp.Vendas.TotalMes = totalMonth
	resp.Vendas.PedidosFinalizados = completedMonth
	resp.Vendas.MetaMes = s.monthlyGoal
	resp.Vendas.FaltamParaMeta = s.monthlyGoal - completedMonth
	if resp.Vendas.FaltamParaMeta < 0 {
		resp.Vendas.FaltamParaMeta = 0
	}
	resp.Vendas.Comissoes = commission

	salesSituacao, err := s.repo.ListSalesSituacao(sellerID)
	if err != nil {
		return resp, err
	}
	now := time.Now()
	for _, row := range salesSituacao {
		problem := "Acompanhar status do pedido"
		if row.DeliveryDate.Valid && row.DeliveryDate.Time.Before(now) {
			problem = "Pedido atrasado"
		} else if row.DepartureDate.Valid && row.DepartureDate.Time.Before(now) {
			problem = "Atraso na expedição"
		} else if row.ArrivalDate.Valid && row.ArrivalDate.Time.Before(now) {
			problem = "Problema na produção"
		}
		item := DashboardPedidoSituacao{
			PedidoID:     row.SaleID,
			Status:       row.Status,
			Problema:     problem,
			DeliveryDate: toTimePtrFromNull(row.DeliveryDate),
		}
		resp.Vendas.SituacaoPedidos = append(resp.Vendas.SituacaoPedidos, item)
		if row.Status == "Em Andamento" || row.Status == "Remessa" {
			if !row.DeliveryDate.Valid || row.DeliveryDate.Time.Before(now) {
				resp.Vendas.AlertasIntervencao = append(resp.Vendas.AlertasIntervencao, DashboardAlertaIntervencao{
					PedidoID: row.SaleID,
					Status:   row.Status,
					Motivo:   "Pedido precisa de atenção",
				})
			}
		}
	}

	return resp, nil
}

func toTimePtr(value *time.Time) *time.Time {
	if value == nil {
		return nil
	}
	return value
}

func toTimePtrFromNull(value sql.NullTime) *time.Time {
	if !value.Valid {
		return nil
	}
	return &value.Time
}
