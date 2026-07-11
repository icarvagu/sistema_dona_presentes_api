package services

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strings"

	"donapresentes/models"
	"donapresentes/repositories"
)

type ArtFinalRepositoryPort interface {
	Dashboard(models.ArtFinalAccess, models.ArtFinalFilters) (map[string]interface{}, error)
	CreateTask(models.ArtFinalTaskInput, int) (int64, error)
	UpdateTask(int64, models.ArtFinalTaskInput, int, models.ArtFinalAccess) error
	CreateStory(models.ArtFinalStoryInput, int) (int64, error)
	CheckStory(int64, bool, int) error
	DeleteStory(int64, int) error
	ListLayoutRequests(models.ArtFinalAccess) (json.RawMessage, error)
	GetLayoutRequest(int64, models.ArtFinalAccess) (json.RawMessage, error)
	CreateLayoutRequest(models.LayoutRequestInput, int, models.ArtFinalAccess) (int64, error)
	TransitionLayoutRequest(int64, string, string, int) error
	AddLayoutMessage(int64, models.LayoutMessageInput, int, models.ArtFinalAccess) error
	AddLayoutVersion(int64, models.LayoutVersionInput, int) (int64, error)
	DecideLayoutVersion(int64, string, string, int, models.ArtFinalAccess) error
	UpsertLayoutJob(string, int64, models.LayoutJobInput, int, models.ArtFinalAccess) error
	ConfirmProductReceived(int64, bool, int, models.ArtFinalAccess) error
	UpdateStoryLifecycle(int64, models.StoryLifecycleInput, int) error
}

func (s *ArtFinalService) AddLayoutMessage(id int64, input models.LayoutMessageInput, userID int, access models.ArtFinalAccess) error {
	input.Message = strings.TrimSpace(input.Message)
	input.FileURL = strings.TrimSpace(input.FileURL)
	if id < 1 || (input.Message == "" && input.FileURL == "") {
		return fmt.Errorf("informe uma mensagem ou anexo")
	}
	return s.repo.AddLayoutMessage(id, input, userID, access)
}

func (s *ArtFinalService) ListLayoutRequests(access models.ArtFinalAccess) (json.RawMessage, error) {
	return s.repo.ListLayoutRequests(access)
}
func (s *ArtFinalService) GetLayoutRequest(id int64, access models.ArtFinalAccess) (json.RawMessage, error) {
	if id < 1 {
		return nil, fmt.Errorf("id invalido")
	}
	return s.repo.GetLayoutRequest(id, access)
}
func (s *ArtFinalService) CreateLayoutRequest(input models.LayoutRequestInput, userID int, access models.ArtFinalAccess) (int64, error) {
	input.SourceType = strings.ToLower(strings.TrimSpace(input.SourceType))
	input.Title = strings.TrimSpace(input.Title)
	input.Instructions = strings.TrimSpace(input.Instructions)
	input.FileMode = strings.ToLower(strings.TrimSpace(input.FileMode))
	input.CommonFileURL = strings.TrimSpace(input.CommonFileURL)
	if input.SourceType != "sale" && input.SourceType != "quote" {
		return 0, fmt.Errorf("origem invalida")
	}
	if input.SourceID < 1 {
		return 0, fmt.Errorf("documento de origem obrigatorio")
	}
	if input.Title == "" {
		return 0, fmt.Errorf("titulo obrigatorio")
	}
	if input.Instructions == "" {
		return 0, fmt.Errorf("observacao do layout obrigatoria")
	}
	if input.FileMode != "none" && input.FileMode != "common" && input.FileMode != "separate" {
		return 0, fmt.Errorf("modo de arquivo invalido")
	}
	if input.FileMode == "common" && input.CommonFileURL == "" {
		return 0, fmt.Errorf("arquivo comum obrigatorio")
	}
	if input.Priority < 0 || input.Priority > 3 {
		return 0, fmt.Errorf("prioridade invalida")
	}
	if len(input.Items) == 0 {
		return 0, fmt.Errorf("selecione ao menos um item")
	}
	expected := "sale_item"
	if input.SourceType == "quote" {
		expected = "quote_item"
	}
	seen := map[int]bool{}
	groups := map[string]string{}
	for i := range input.Items {
		item := &input.Items[i]
		item.EntityType = strings.ToLower(strings.TrimSpace(item.EntityType))
		item.GroupKey = strings.TrimSpace(item.GroupKey)
		item.SourceFileURL = strings.TrimSpace(item.SourceFileURL)
		if item.EntityType != expected || item.ItemID < 1 {
			return 0, fmt.Errorf("item invalido para a origem")
		}
		if seen[item.ItemID] {
			return 0, fmt.Errorf("item duplicado")
		}
		seen[item.ItemID] = true
		if input.FileMode == "separate" && item.SourceFileURL == "" {
			return 0, fmt.Errorf("arquivo separado obrigatorio para todos os itens")
		}
		if item.GroupKey != "" {
			if file, ok := groups[item.GroupKey]; ok && file != item.SourceFileURL {
				return 0, fmt.Errorf("arquivo inconsistente no grupo %s", item.GroupKey)
			}
			groups[item.GroupKey] = item.SourceFileURL
		}
	}
	id, err := s.repo.CreateLayoutRequest(input, userID, access)
	if err == nil {
		s.auditService.LogSimple(&userID, "layout_request_created", "art_final", fmt.Sprintf("source_type=%s source_id=%d request_id=%d", input.SourceType, input.SourceID, id), "")
	}
	return id, err
}

func (s *ArtFinalService) TransitionLayoutRequest(id int64, next, note string, userID int) error {
	next = strings.ToLower(strings.TrimSpace(next))
	if id < 1 {
		return fmt.Errorf("id invalido")
	}
	allowed := map[string]bool{"requested": true, "in_progress": true, "awaiting_approval": true, "changes_requested": true, "approved": true, "cancelled": true}
	if !allowed[next] {
		return fmt.Errorf("status de layout invalido")
	}
	err := s.repo.TransitionLayoutRequest(id, next, strings.TrimSpace(note), userID)
	if err == nil {
		s.auditService.LogSimple(&userID, "layout_transition", "art_final", fmt.Sprintf("request_id=%d to=%s", id, next), "")
	}
	return err
}
func (s *ArtFinalService) AddLayoutVersion(itemID int64, input models.LayoutVersionInput, userID int) (int64, error) {
	input.FileURL = strings.TrimSpace(input.FileURL)
	input.Label = strings.TrimSpace(input.Label)
	if itemID < 1 || input.FileURL == "" {
		return 0, fmt.Errorf("item e arquivo sao obrigatorios")
	}
	if input.Label == "" {
		input.Label = "Layout"
	}
	return s.repo.AddLayoutVersion(itemID, input, userID)
}
func (s *ArtFinalService) DecideLayoutVersion(versionID int64, input models.LayoutDecisionInput, userID int, access models.ArtFinalAccess) error {
	input.Status = strings.ToLower(strings.TrimSpace(input.Status))
	input.Note = strings.TrimSpace(input.Note)
	if versionID < 1 || (input.Status != "approved" && input.Status != "changes_requested") {
		return fmt.Errorf("decisao invalida")
	}
	if input.Status == "changes_requested" && input.Note == "" {
		return fmt.Errorf("motivo da alteracao obrigatorio")
	}
	err := s.repo.DecideLayoutVersion(versionID, input.Status, input.Note, userID, access)
	if err == nil {
		s.auditService.LogSimple(&userID, "layout_version_decided", "art_final", fmt.Sprintf("version_id=%d status=%s", versionID, input.Status), "")
	}
	return err
}
func (s *ArtFinalService) UpsertLayoutJob(kind string, itemID int64, input models.LayoutJobInput, userID int, access models.ArtFinalAccess) error {
	kind = strings.ToLower(strings.TrimSpace(kind))
	input.Status = strings.ToLower(strings.TrimSpace(input.Status))
	input.FileURL = strings.TrimSpace(input.FileURL)
	input.Reason = strings.TrimSpace(input.Reason)
	input.ExternalContact = strings.TrimSpace(input.ExternalContact)
	valid := map[string]map[string]bool{"corel": {"pending": true, "sent": true, "received": true, "cancelled": true}, "engraving": {"pending": true, "in_progress": true, "ready": true, "approved": true, "cancelled": true}}
	if itemID < 1 || !valid[kind][input.Status] {
		return fmt.Errorf("job invalido")
	}
	if (kind == "corel" && input.Status == "received" || kind == "engraving" && (input.Status == "ready" || input.Status == "approved")) && input.FileURL == "" {
		return fmt.Errorf("arquivo obrigatorio para concluir")
	}
	err := s.repo.UpsertLayoutJob(kind, itemID, input, userID, access)
	if err == nil {
		s.auditService.LogSimple(&userID, "layout_job_upsert", "art_final", fmt.Sprintf("item_id=%d kind=%s status=%s", itemID, kind, input.Status), "")
	}
	return err
}
func (s *ArtFinalService) ConfirmProductReceived(itemID int64, received bool, userID int, access models.ArtFinalAccess) error {
	if itemID < 1 {
		return fmt.Errorf("item invalido")
	}
	return s.repo.ConfirmProductReceived(itemID, received, userID, access)
}
func (s *ArtFinalService) UpdateStoryLifecycle(id int64, input models.StoryLifecycleInput, userID int) error {
	input.Status = strings.ToLower(strings.TrimSpace(input.Status))
	input.Observation = strings.TrimSpace(input.Observation)
	allowed := map[string]bool{"draft": true, "returned": true, "corrected": true, "published": true, "expired": true}
	if id < 1 || !allowed[input.Status] {
		return fmt.Errorf("status do Story invalido")
	}
	if input.Status == "returned" && input.Observation == "" {
		return fmt.Errorf("motivo da devolucao obrigatorio")
	}
	return s.repo.UpdateStoryLifecycle(id, input, userID)
}

type ArtFinalService struct {
	repo         ArtFinalRepositoryPort
	auditService *AuditService
}

func NewArtFinalService(repo *repositories.ArtFinalRepository, auditService *AuditService) *ArtFinalService {
	return &ArtFinalService{repo: repo, auditService: auditService}
}
func NewArtFinalServiceWithRepository(repo ArtFinalRepositoryPort, auditService *AuditService) *ArtFinalService {
	return &ArtFinalService{repo: repo, auditService: auditService}
}

func (s *ArtFinalService) Dashboard(access models.ArtFinalAccess, filters models.ArtFinalFilters) (map[string]interface{}, error) {
	allowedCategories := map[string]bool{"": true, "layout": true, "alteracao": true, "corel": true, "gravacao": true, "video": true, "post": true, "banner": true, "mala": true, "gravados": true, "sem_gravacao": true}
	allowedStatuses := map[string]bool{"": true, "open": true, "done": true, "cancelled": true, "pending": true, "changes_requested": true}
	filters.Category = strings.ToLower(strings.TrimSpace(filters.Category))
	filters.Status = strings.ToLower(strings.TrimSpace(filters.Status))
	filters.Search = strings.TrimSpace(filters.Search)
	if !allowedCategories[filters.Category] {
		return nil, fmt.Errorf("categoria invalida")
	}
	if !allowedStatuses[filters.Status] {
		return nil, fmt.Errorf("status invalido")
	}
	return s.repo.Dashboard(access, filters)
}

func validateArtFinalTask(input *models.ArtFinalTaskInput) error {
	input.Panel = strings.ToLower(strings.TrimSpace(input.Panel))
	input.Category = strings.ToLower(strings.TrimSpace(input.Category))
	input.Title = strings.TrimSpace(input.Title)
	input.Description = strings.TrimSpace(input.Description)
	input.Status = strings.ToLower(strings.TrimSpace(input.Status))
	input.AttachmentURL = strings.TrimSpace(input.AttachmentURL)
	input.Channel = strings.TrimSpace(input.Channel)
	input.Format = strings.TrimSpace(input.Format)
	if input.Status == "" {
		input.Status = "open"
	}
	if input.Panel != "pending" && input.Panel != "media" {
		return fmt.Errorf("painel invalido")
	}
	pending := map[string]bool{"layout": true, "alteracao": true, "corel": true, "gravacao": true}
	media := map[string]bool{"video": true, "post": true, "banner": true, "mala": true}
	if (input.Panel == "pending" && !pending[input.Category]) || (input.Panel == "media" && !media[input.Category]) {
		return fmt.Errorf("categoria invalida para o painel")
	}
	if input.Title == "" {
		return fmt.Errorf("titulo obrigatorio")
	}
	if input.Panel == "media" && (input.DueAt == nil || input.Channel == "" || input.Format == "") {
		return fmt.Errorf("data/hora, canal e formato sao obrigatorios para midia")
	}
	if input.Priority < 0 || input.Priority > 3 {
		return fmt.Errorf("prioridade deve estar entre 0 e 3")
	}
	if input.Status != "open" && input.Status != "done" && input.Status != "cancelled" {
		return fmt.Errorf("status invalido")
	}
	if input.AttachmentURL != "" {
		u, err := url.ParseRequestURI(input.AttachmentURL)
		if err != nil || (!(u.Scheme == "http" || u.Scheme == "https") || u.Host == "") && !strings.HasPrefix(input.AttachmentURL, "/uploads/") && !strings.HasPrefix(input.AttachmentURL, "/files/") && !strings.HasPrefix(input.AttachmentURL, "/storage/") {
			return fmt.Errorf("anexo deve ser uma URL http(s) ou caminho interno")
		}
	}
	return nil
}

func (s *ArtFinalService) CreateTask(input models.ArtFinalTaskInput, userID int) (int64, error) {
	if err := validateArtFinalTask(&input); err != nil {
		return 0, err
	}
	return s.repo.CreateTask(input, userID)
}
func (s *ArtFinalService) UpdateTask(id int64, input models.ArtFinalTaskInput, userID int, access models.ArtFinalAccess) error {
	if id < 1 {
		return fmt.Errorf("id invalido")
	}
	if err := validateArtFinalTask(&input); err != nil {
		return err
	}
	return s.repo.UpdateTask(id, input, userID, access)
}
func (s *ArtFinalService) CreateStory(input models.ArtFinalStoryInput, userID int) (int64, error) {
	input.Title = strings.TrimSpace(input.Title)
	if input.SaleID < 1 {
		return 0, fmt.Errorf("pedido obrigatorio")
	}
	if input.SaleItemID != nil && *input.SaleItemID < 1 {
		return 0, fmt.Errorf("item invalido")
	}
	if input.Title == "" {
		input.Title = fmt.Sprintf("Pedido %d", input.SaleID)
	}
	return s.repo.CreateStory(input, userID)
}
func (s *ArtFinalService) CheckStory(id int64, checked bool, userID int) error {
	if id < 1 {
		return fmt.Errorf("id invalido")
	}
	return s.repo.CheckStory(id, checked, userID)
}
func (s *ArtFinalService) DeleteStory(id int64, userID int) error {
	if id < 1 {
		return fmt.Errorf("id invalido")
	}
	return s.repo.DeleteStory(id, userID)
}
