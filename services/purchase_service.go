package services

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"donapresentes/models"
	"donapresentes/repositories"
)

// PurchaseService manages the complete purchase order workflow from release through
// supplier emails, Corel file management, payment approval, issue tracking,
// first piece approval, and production release.
type PurchaseService struct {
	repo         *repositories.PurchaseRepository
	mailer       PurchaseMailer
	auditService *AuditService
}

// NewPurchaseService creates a PurchaseService with the required repository and audit trail.
// It initializes the SMTP mailer from environment variables.
func NewPurchaseService(repo *repositories.PurchaseRepository, auditService *AuditService) *PurchaseService {
	return &PurchaseService{repo: repo, mailer: NewSMTPPurchaseMailerFromEnv(), auditService: auditService}
}

// List returns purchase orders, optionally filtered by sample flag.
func (s *PurchaseService) List(isSample *bool) ([]models.PurchaseOrder, error) {
	return s.repo.List(isSample)
}

// GetByID returns a single purchase order by its ID.
func (s *PurchaseService) GetByID(id int) (*models.PurchaseOrder, error) {
	return s.repo.GetByID(id)
}

// GetBySaleID returns the purchase order associated with a given sale ID.
func (s *PurchaseService) GetBySaleID(saleID int) (*models.PurchaseOrder, error) {
	return s.repo.GetBySaleID(saleID)
}

// ReleaseSale creates a purchase order from a sale, optionally marking it as a sample
// and indicating whether the sample requires engraving.
func (s *PurchaseService) ReleaseSale(saleID, userID int, isSample, sampleHasEngraving bool) (*models.PurchaseOrder, error) {
	if saleID <= 0 {
		return nil, errors.New("pedido inválido")
	}
	p, err := s.repo.ReleaseSale(saleID, userID, isSample, sampleHasEngraving)
	if err == nil {
		detail := fmt.Sprintf("sale_id=%d is_sample=%v", saleID, isSample)
		s.auditService.LogSimple(&userID, "purchase_released", "purchase", detail, "")
	}
	return p, err
}

// Update modifies cost values on a purchase order. All costs must be non-negative.
func (s *PurchaseService) Update(id, userID int, input *models.PurchaseUpdateInput) (*models.PurchaseOrder, error) {
	if input.MaterialTotalCost < 0 || input.EngravingCost < 0 || input.FreightCost < 0 || input.OtherCost < 0 {
		return nil, errors.New("os valores de custo não podem ser negativos")
	}
	if input.IsSample && input.SampleHasEngraving {
		input.HasEngraving = true
	}
	if err := s.repo.Update(id, input, userID); err != nil {
		return nil, err
	}
	return s.repo.GetByID(id)
}

// ExecuteAction performs a workflow action on a purchase order. Supported actions:
// start_review, request_corel, attach_corel, send_material_email, send_engraving_email,
// accept_material, accept_engraving, finalize_purchase, release_production,
// first_piece_received, approve_first_piece.
func (s *PurchaseService) ExecuteAction(id, userID int, input *models.PurchaseActionInput) (*models.PurchaseOrder, error) {
	p, err := s.repo.GetByID(id)
	if err != nil {
		return nil, err
	}
	action := strings.TrimSpace(input.Action)
	switch action {
	case "start_review":
		err = s.repo.SetStatus(id, models.PurchaseReview, userID, "Conferência iniciada", input.Observation)
	case "request_corel":
		if !p.HasEngraving {
			return nil, errors.New("o pedido não possui gravação")
		}
		err = s.repo.MarkCorelRequested(id, userID)
		if err == nil {
			err = s.repo.SetStatus(id, models.PurchaseWaitingCorel, userID, "Arquivo Corel solicitado", input.Observation)
		}
		if err == nil {
			err = s.repo.Notify(id, "arte_final", "corel_solicitado", fmt.Sprintf("Arquivo Corel solicitado para o pedido %s", p.GeneralNumber), nil)
		}
	case "attach_corel":
		if !p.CorelRequired || p.CorelRequestedAt == nil {
			return nil, errors.New("o arquivo Corel precisa ter sido solicitado antes do anexo")
		}
		if strings.TrimSpace(input.URL) == "" {
			return nil, errors.New("informe o arquivo Corel")
		}
		name := input.FileName
		if name == "" {
			name = "arquivo-corel.cdr"
		}
		err = s.repo.AddAttachment(id, "corel", name, input.URL, userID)
		if err == nil {
			err = s.repo.MarkCorelAttached(id, userID)
		}
		if err == nil {
			err = s.repo.SetStatus(id, models.PurchaseCorelReady, userID, "Corel anexado", input.Observation)
		}
		if err == nil {
			err = s.repo.Notify(id, "compras", "corel_anexado", fmt.Sprintf("Corel do pedido %s foi anexado", p.GeneralNumber), nil)
		}
	case "send_material_email":
		err = s.sendSupplierEmail(p, userID, "material", input)
		if err == nil {
			status := models.PurchaseMaterialEmailSent
			if !p.HasEngraving || hasEmail(p, "gravacao") {
				status = models.PurchaseWaitingAcceptance
			}
			err = s.repo.SetStatus(id, status, userID, "E-mail de material registrado", input.Observation)
		}
	case "send_engraving_email":
		if p.CorelRequired && p.CorelAttachedAt == nil {
			return nil, errors.New("a formalização da gravação está bloqueada até o Corel ser anexado")
		}
		err = s.sendSupplierEmail(p, userID, "gravacao", input)
		if err == nil {
			status := models.PurchaseEngravingEmailSent
			if hasEmail(p, "material") {
				status = models.PurchaseWaitingAcceptance
			}
			err = s.repo.SetStatus(id, status, userID, "E-mail de gravação registrado", input.Observation)
		}
	case "accept_material":
		err = s.repo.SetAcceptance(id, false)
		if err == nil {
			err = s.repo.AddHistory(id, "Aceite do fornecedor de material", p.Status, p.Status, input.Observation, userID)
		}
	case "accept_engraving":
		err = s.repo.SetAcceptance(id, true)
		if err == nil {
			err = s.repo.AddHistory(id, "Aceite do fornecedor de gravação", p.Status, p.Status, input.Observation, userID)
		}
	case "finalize_purchase":
		if !p.MaterialAccepted {
			return nil, errors.New("confirme o aceite do fornecedor de material")
		}
		if !p.EngravingAccepted {
			return nil, errors.New("confirme o aceite do pedido de gravação")
		}
		for _, payment := range p.Payments {
			if payment.Status != "Pagamento Registrado" {
				return nil, errors.New("há pagamento pendente de aprovação")
			}
		}
		err = s.repo.SetStatus(id, models.PurchaseBought, userID, "Compra finalizada", input.Observation)
	case "release_production":
		if p.Status != models.PurchaseBought && p.Status != models.PurchaseInEngraving {
			return nil, errors.New("o pedido precisa estar comprado antes da liberação para Produção")
		}
		if !p.MaterialAccepted || !p.EngravingAccepted {
			return nil, errors.New("a liberação está bloqueada até o aceite dos pedidos de material e gravação")
		}
		err = s.repo.SetStatus(id, models.PurchaseReleased, userID, "Liberado para Produção", input.Observation)
	case "first_piece_received":
		if !p.FirstPieceRequired {
			return nil, errors.New("o pedido não exige aprovação da primeira peça")
		}
		if input.URL == "" {
			return nil, errors.New("anexe a foto da primeira peça")
		}
		name := input.FileName
		if name == "" {
			name = "primeira-peca"
		}
		err = s.repo.AddAttachment(id, "primeira_peca", name, input.URL, userID)
		if err == nil {
			err = s.repo.SetFirstPiece(id, "Primeira Peça Recebida", input.URL)
		}
		if err == nil {
			err = s.repo.SetStatus(id, models.PurchaseWaitingFirstPiece, userID, "Primeira peça enviada para aprovação", input.Observation)
		}
		if err == nil && p.Sale != nil {
			sellerID := p.Sale.SellerID
			err = s.repo.Notify(id, "vendas", "primeira_peca", fmt.Sprintf("Aprove a primeira peça do pedido %s", p.GeneralNumber), &sellerID)
		}
	case "approve_first_piece":
		if p.Status != models.PurchaseWaitingFirstPiece {
			return nil, errors.New("o pedido não está aguardando aprovação da primeira peça")
		}
		err = s.repo.SetFirstPiece(id, "Gravação Aprovada", "")
		if err == nil {
			err = s.repo.SetStatus(id, models.PurchaseInEngraving, userID, "Primeira peça aprovada", input.Observation)
		}
		if err == nil {
			err = s.repo.Notify(id, "compras", "primeira_peca_aprovada", fmt.Sprintf("Primeira peça do pedido %s aprovada", p.GeneralNumber), nil)
		}
	default:
		return nil, errors.New("ação de compra inválida")
	}
	if err != nil {
		return nil, err
	}
	s.auditService.LogSimple(&userID, "purchase_action_"+action, "purchase", fmt.Sprintf("purchase_id=%d user_id=%d", id, userID), "")
	return s.repo.GetByID(id)
}

// AddAttachment adds a file attachment to a purchase order.
func (s *PurchaseService) AddAttachment(id, userID int, a *models.PurchaseAttachment) (*models.PurchaseOrder, error) {
	if a.Category == "" || a.URL == "" {
		return nil, errors.New("categoria e arquivo são obrigatórios")
	}
	if a.FileName == "" {
		a.FileName = "anexo"
	}
	if err := s.repo.AddAttachment(id, a.Category, a.FileName, a.URL, userID); err != nil {
		return nil, err
	}
	if err := s.repo.AddHistory(id, "Anexo incluído", "", "", a.Category+": "+a.FileName, userID); err != nil {
		return nil, err
	}
	return s.repo.GetByID(id)
}

// AddPayment records a payment on a purchase order. PIX and cheque payments require
// justification; cheques also require a receipt image URL.
func (s *PurchaseService) AddPayment(id, userID int, input *models.PurchasePaymentInput) (*models.PurchaseOrder, error) {
	if input.Amount <= 0 || input.CostType == "" || input.Method == "" {
		return nil, errors.New("tipo de custo, valor e forma de pagamento são obrigatórios")
	}
	if (input.Method == "PIX" || input.Method == "Cheque" || input.Method == "Cartão de Crédito") && strings.TrimSpace(input.Justification) == "" {
		return nil, errors.New("a justificativa é obrigatória para esta forma de pagamento")
	}
	if input.Method == "Cheque" && input.ReceiptURL == "" {
		return nil, errors.New("anexe a foto do cheque")
	}
	if err := s.repo.AddPayment(id, userID, input); err != nil {
		return nil, err
	}
	if input.Method == "PIX" || input.Method == "Cheque" {
		_ = s.repo.Notify(id, "diretoria_financeira", "pagamento_aprovacao", fmt.Sprintf("%s de R$ %.2f aguarda aprovação", input.Method, input.Amount), nil)
	}
	_ = s.repo.SetStatus(id, models.PurchaseWaitingPayment, userID, "Pagamento registrado", input.Justification)
	s.auditService.LogSimple(&userID, "purchase_payment_added", "purchase", fmt.Sprintf("purchase_id=%d amount=%.2f method=%s", id, input.Amount, input.Method), "")
	return s.repo.GetByID(id)
}

// ApprovePayment marks a payment as approved and notifies Compras and Financeiro.
func (s *PurchaseService) ApprovePayment(paymentID, userID int, receiptURL string) (*models.PurchaseOrder, error) {
	purchaseID, err := s.repo.ApprovePayment(paymentID, userID, receiptURL)
	if err != nil {
		return nil, err
	}
	_ = s.repo.AddHistory(purchaseID, "Pagamento aprovado", "", "Pagamento Registrado", "Comprovante/liberação financeira registrada", userID)
	_ = s.repo.Notify(purchaseID, "compras", "pagamento_aprovado", "Pagamento aprovado e registrado", nil)
	_ = s.repo.Notify(purchaseID, "financeiro", "pagamento_aprovado", "Pagamento aprovado e registrado", nil)
	s.auditService.LogSimple(&userID, "purchase_payment_approved", "purchase", fmt.Sprintf("purchase_id=%d payment_id=%d", purchaseID, paymentID), "")
	return s.repo.GetByID(purchaseID)
}

// AddIssue registers a supplier-related issue (material or engraving) on a purchase order
// with description, occurrence date, resolution deadline, and priority.
func (s *PurchaseService) AddIssue(id, userID int, input *models.PurchaseIssueInput) (*models.PurchaseOrder, error) {
	if input.IssueType != "material" && input.IssueType != "gravacao" {
		return nil, errors.New("tipo de pendência inválido")
	}
	if strings.TrimSpace(input.Description) == "" || input.OccurrenceDate.IsZero() || input.ResolutionDeadline.IsZero() {
		return nil, errors.New("descrição, data da ocorrência e prazo de resolução são obrigatórios")
	}
	if input.Priority < 1 || input.Priority > 3 {
		input.Priority = 1
	}
	if err := s.repo.AddIssue(id, userID, input); err != nil {
		return nil, err
	}
	status := models.PurchaseMaterialIssue
	if input.IssueType == "gravacao" {
		status = models.PurchaseEngravingIssue
	}
	if err := s.repo.SetStatus(id, status, userID, "Pendência registrada", input.Description); err != nil {
		return nil, err
	}
	_ = s.repo.Notify(id, "compras", "pendencia_"+input.IssueType, input.Description, nil)
	_ = s.repo.Notify(id, "producao", "pendencia_"+input.IssueType, input.Description, nil)
	return s.repo.GetByID(id)
}

// UpdateIssue updates or resolves a purchase issue. Resolution requires a solution description.
func (s *PurchaseService) UpdateIssue(issueID, userID int, input *models.PurchaseIssueUpdateInput) (*models.PurchaseOrder, error) {
	if input.Status == "Resolvida" && strings.TrimSpace(input.Solution) == "" {
		return nil, errors.New("informe a solução do fornecedor")
	}
	purchaseID, err := s.repo.UpdateIssue(issueID, userID, input)
	if err != nil {
		return nil, err
	}
	_ = s.repo.AddHistory(purchaseID, "Pendência atualizada", "", input.Status, input.Solution, userID)
	return s.repo.GetByID(purchaseID)
}

// FinancialSummary returns aggregated financial data for all purchase orders.
func (s *PurchaseService) FinancialSummary() ([]models.PurchaseFinancialSummary, error) {
	return s.repo.FinancialSummary()
}

// ListNotifications returns notifications for a user filtered by their permissions.
func (s *PurchaseService) ListNotifications(userID int, permissions []string) ([]models.Notification, error) {
	return s.repo.ListNotifications(userID, permissions)
}

// MarkNotificationRead marks a notification as read for a user.
func (s *PurchaseService) MarkNotificationRead(id, userID int) error {
	return s.repo.MarkNotificationRead(id, userID)
}

func (s *PurchaseService) sendSupplierEmail(p *models.PurchaseOrder, userID int, kind string, input *models.PurchaseActionInput) error {
	var supplier *models.Supplier
	var deadline string
	if kind == "material" {
		supplier = p.MaterialSupplier
		if p.MaterialDeadline == nil {
			return errors.New("informe o prazo de recebimento ou retirada do material")
		}
		deadline = p.MaterialDeadline.Format("02/01/2006")
	} else {
		if !p.HasEngraving {
			return errors.New("o pedido não possui gravação")
		}
		supplier = p.EngravingSupplier
		if p.EngravingDeadline == nil {
			return errors.New("informe o prazo de gravação/retirada")
		}
		deadline = p.EngravingDeadline.Format("02/01/2006")
	}
	if supplier == nil {
		return errors.New("selecione o fornecedor antes de formalizar o e-mail")
	}
	recipient := strings.TrimSpace(input.Recipient)
	if recipient == "" {
		if supplier.ResponsibleEmail != nil {
			recipient = *supplier.ResponsibleEmail
		}
		if recipient == "" && supplier.Email != nil {
			recipient = *supplier.Email
		}
	}
	if recipient == "" {
		return errors.New("o fornecedor não possui e-mail cadastrado")
	}
	product, code, color, quantity, engraving := purchaseProductSummary(p)
	subject := fmt.Sprintf("Pedido de Compra - Pedido nº %s - %s/%s", p.GeneralNumber, product, code)
	body := fmt.Sprintf("Fornecedor: %s\nPedido: %s\nProduto: %s\nCódigo: %s\nCor: %s\nQuantidade: %d\nValor total: R$ %.2f\nPrazo: %s\nForma de pagamento: %s", supplier.Name, p.GeneralNumber, product, code, color, quantity, p.MaterialTotalCost, deadline, p.PaymentMethod)
	statusKind := "material"
	if kind == "gravacao" {
		customer := ""
		if p.Sale != nil && p.Sale.Customer != nil {
			customer = p.Sale.Customer.Name
		}
		subject = fmt.Sprintf("Pedido de Gravação - Pedido nº %s - %s/%s", p.GeneralNumber, customer, product)
		body = fmt.Sprintf("Fornecedor: %s\nPedido: %s\nCliente: %s\nProduto: %s\nCor: %s\nQuantidade: %d\nGravação: %s\nPrazo/retirada: %s\nArquivo Corel e layout aprovado anexos.", supplier.Name, p.GeneralNumber, customer, product, color, quantity, engraving, deadline)
		statusKind = "gravacao"
	}
	if input.Observation != "" {
		body += "\n\nObservações: " + input.Observation
	}
	attachments, _ := json.Marshal(input.Attachments)
	email := &models.PurchaseEmail{PurchaseID: p.ID, Kind: statusKind, Recipient: recipient, Subject: subject, Body: body, Observation: input.Observation, Attachments: attachments, SentBy: &userID}
	if err := s.mailer.Send(recipient, subject, body, input.Attachments); err != nil {
		return fmt.Errorf("não foi possível enviar o e-mail ao fornecedor: %w", err)
	}
	return s.repo.AddEmail(email)
}

func hasEmail(p *models.PurchaseOrder, kind string) bool {
	for _, email := range p.Emails {
		if email.Kind == kind {
			return true
		}
	}
	return false
}

func purchaseProductSummary(p *models.PurchaseOrder) (string, string, string, int, string) {
	if p.Sale == nil || len(p.Sale.Items) == 0 {
		return "Produto", "", "", 0, ""
	}
	item := p.Sale.Items[0]
	product, code, color := "Produto", "", ""
	if item.Product != nil {
		product = item.Product.ProductName
		code = item.Product.SupplierCode
		color = item.Product.Color
	}
	engraving := ""
	var entries []map[string]interface{}
	if json.Unmarshal(item.Engravings, &entries) == nil && len(entries) > 0 {
		engraving = fmt.Sprint(entries[0]["type"])
	}
	return product, code, color, item.Quantity, engraving
}
