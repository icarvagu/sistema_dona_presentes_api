package services

import (
	"fmt"
	"net/url"
	"strings"
	"time"

	"donapresentes/models"
	"donapresentes/repositories"
)

type SalesWorkflowService struct {
	repo         *repositories.SalesWorkflowRepository
	auditService *AuditService
}

func NewSalesWorkflowService(repo *repositories.SalesWorkflowRepository, auditService *AuditService) *SalesWorkflowService {
	return &SalesWorkflowService{repo: repo, auditService: auditService}
}

func (s *SalesWorkflowService) QuoteOwner(id int) (int, error) { return s.repo.QuoteOwner(id) }
func (s *SalesWorkflowService) SaleOwner(id int) (int, error)  { return s.repo.SaleOwner(id) }
func (s *SalesWorkflowService) ItemOwner(entity string, id int) (int, error) {
	return s.repo.ItemOwner(entity, id)
}
func (s *SalesWorkflowService) AddFeedback(quoteID, userID int, when *time.Time, note string) (*models.QuoteFeedbackEvent, error) {
	if strings.TrimSpace(note) == "" {
		return nil, fmt.Errorf("observation is required")
	}
	if when == nil {
		defaultDate := time.Now().Add(72 * time.Hour)
		when = &defaultDate
	}
	return s.repo.AddFeedback(quoteID, userID, when, strings.TrimSpace(note))
}
func (s *SalesWorkflowService) Feedbacks(id int) ([]models.QuoteFeedbackEvent, error) {
	return s.repo.Feedbacks(id)
}
func (s *SalesWorkflowService) CreateLayout(entity string, itemID, userID int, fileURL string) (*models.ItemLayoutVersion, error) {
	if entity != "quote_item" && entity != "sale_item" {
		return nil, fmt.Errorf("invalid entity_type")
	}
	if err := validateFileURL(fileURL); err != nil {
		return nil, err
	}
	return s.repo.CreateLayout(entity, itemID, userID, strings.TrimSpace(fileURL))
}
func (s *SalesWorkflowService) ApproveLayout(id int64, userID int, status, note string) error {
	valid := map[string]bool{"approved": true, "changes_requested": true, "rejected": true}
	if !valid[status] {
		return fmt.Errorf("invalid layout approval status")
	}
	err := s.repo.ApproveLayout(id, userID, status, note)
	if err == nil {
		s.auditService.LogSimple(&userID, "layout_approved", "sales_workflow", fmt.Sprintf("version_id=%d status=%s", id, status), "")
	}
	return err
}
func (s *SalesWorkflowService) ConvertQuote(id, userID int, input models.QuoteConversionInput) (int, error) {
	saleID, err := s.repo.ConvertQuote(id, userID, input.WithdrawalDates)
	if err == nil {
		s.auditService.LogSimple(&userID, "quote_converted", "sales_workflow", fmt.Sprintf("quote_id=%d sale_id=%d", id, saleID), "")
	}
	return saleID, err
}

var allowedFinancialTags = map[string]bool{
	"SERASA OK":                 true,
	"50% + BOLETO":              true,
	"SOMENTE À VISTA":           true,
	"50% + SAÍDA":               true,
	"RECORRENTE":                true,
	"LIBERADO COM APONTAMENTOS": true,
	"NEGADO":                    true,
}

var financialTagAliases = map[string]string{
	"50%+boleto":   "50% + BOLETO",
	"só à vista":   "SOMENTE À VISTA",
	"so a vista":   "SOMENTE À VISTA",
	"50%+saída":    "50% + SAÍDA",
	"50%+saida":    "50% + SAÍDA",
	"recorrente":   "RECORRENTE",
	"apontamentos": "LIBERADO COM APONTAMENTOS",
	"negado":       "NEGADO",
}

func (s *SalesWorkflowService) UpsertFinancial(saleID, userID int, input models.FinancialAnalysisInput) error {
	if input.Status != "pending" && input.Status != "approved" && input.Status != "rejected" {
		return fmt.Errorf("invalid financial status")
	}
	for i, tag := range input.Tags {
		normalized := normalizeFinancialTag(tag)
		if !allowedFinancialTags[normalized] {
			return fmt.Errorf("invalid financial tag: %s", tag)
		}
		input.Tags[i] = normalized
	}
	if input.Status == "rejected" && !contains(input.Tags, "NEGADO") {
		input.Tags = append(input.Tags, "NEGADO")
	}
	err := s.repo.UpsertFinancial(saleID, userID, input)
	if err == nil {
		s.auditService.LogSimple(&userID, "financial_upsert", "sales_workflow", fmt.Sprintf("sale_id=%d status=%s", saleID, input.Status), "")
	}
	return err
}
func normalizeFinancialTag(tag string) string {
	tag = strings.TrimSpace(tag)
	if alias, ok := financialTagAliases[strings.ToLower(tag)]; ok {
		return alias
	}
	return strings.ToUpper(tag)
}
func contains(values []string, want string) bool {
	for _, v := range values {
		if v == want {
			return true
		}
	}
	return false
}
func (s *SalesWorkflowService) AddReceipt(saleID, userID int, fileURL string) (int64, error) {
	if err := validateFileURL(fileURL); err != nil {
		return 0, err
	}
	return s.repo.AddReceipt(saleID, userID, strings.TrimSpace(fileURL))
}
func (s *SalesWorkflowService) ValidateReceipt(id int64, userID int, status, note string) error {
	if status != "validated" && status != "rejected" {
		return fmt.Errorf("invalid receipt status")
	}
	return s.repo.ValidateReceipt(id, userID, status, note)
}
func (s *SalesWorkflowService) AddPending(saleID, userID int, input models.SalePendingInput) (int64, error) {
	if strings.TrimSpace(input.Sector) == "" || strings.TrimSpace(input.Description) == "" {
		return 0, fmt.Errorf("sector and description are required")
	}
	return s.repo.AddPending(saleID, userID, input)
}
func (s *SalesWorkflowService) ResolvePending(id int64, userID int, note string) error {
	return s.repo.ResolvePending(id, userID, note)
}
func (s *SalesWorkflowService) PendingSaleOwner(id int64) (int, error) {
	return s.repo.PendingSaleOwner(id)
}
func (s *SalesWorkflowService) AddEngravingApproval(itemID, userID int, input models.EngravingApprovalInput) (int64, error) {
	input.Response = normalizeEngravingResponse(input.Response)
	valid := map[string]bool{"APROVADO": true, "ERRADO": true, "COR_DIFERENTE": true, "NOVA_FOTO": true}
	if !valid[input.Response] {
		return 0, fmt.Errorf("invalid engraving response")
	}
	return s.repo.AddEngravingApproval(itemID, userID, input)
}
func normalizeEngravingResponse(response string) string {
	response = strings.TrimSpace(response)
	aliases := map[string]string{
		"approved":            "APROVADO",
		"approved_with_notes": "APROVADO",
		"changes_requested":   "NOVA_FOTO",
		"rejected":            "ERRADO",
		"aprovado":            "APROVADO",
		"errado":              "ERRADO",
		"cor_diferente":       "COR_DIFERENTE",
		"nova_foto":           "NOVA_FOTO",
	}
	if value, ok := aliases[strings.ToLower(response)]; ok {
		return value
	}
	return strings.ToUpper(response)
}
func (s *SalesWorkflowService) RecordEngravingChannel(id int64, userID int, channel string) error {
	valid := map[string]bool{"whatsapp": true, "email": true, "presencial": true, "telefone": true, "outro": true}
	if !valid[channel] {
		return fmt.Errorf("invalid channel")
	}
	return s.repo.RecordEngravingChannel(id, userID, channel)
}
func (s *SalesWorkflowService) ApproveSeller(saleID, userID int) error {
	err := s.repo.SetSellerApproval(saleID, userID)
	if err == nil {
		s.auditService.LogSimple(&userID, "seller_approved", "sales_workflow", fmt.Sprintf("sale_id=%d", saleID), "")
	}
	return err
}
func (s *SalesWorkflowService) EnsureReadyForPurchases(saleID int) error {
	return s.repo.EnsureReadyForPurchases(saleID)
}
func (s *SalesWorkflowService) MarkReleasedToPurchases(saleID int) error {
	return s.repo.MarkReleasedToPurchases(saleID)
}
func (s *SalesWorkflowService) Release(saleID int) error { return s.repo.ReleaseToPurchases(saleID) }
func (s *SalesWorkflowService) Dashboard(sellerID int, all bool) (map[string]interface{}, error) {
	return s.repo.Dashboard(sellerID, all)
}
func (s *SalesWorkflowService) SetQuoteImportant(id int, important bool) error {
	return s.repo.SetQuoteImportant(id, important)
}
func (s *SalesWorkflowService) SaleWorkflow(id int) (map[string]interface{}, error) {
	return s.repo.SaleWorkflow(id)
}

func validateFileURL(raw string) error {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return fmt.Errorf("file_url is required")
	}
	if strings.HasPrefix(raw, "/uploads/") || strings.HasPrefix(raw, "/files/") || strings.HasPrefix(raw, "/storage/") {
		return nil
	}
	u, err := url.ParseRequestURI(raw)
	if err != nil || !(u.Scheme == "http" || u.Scheme == "https") || u.Host == "" {
		return fmt.Errorf("file_url must be an http(s) URL or an internal /uploads, /files or /storage path")
	}
	return nil
}
