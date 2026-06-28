package services

import (
	"donapresentes/models"
	"encoding/json"
	"testing"
	"time"
)

type artFinalRepoStub struct{ created models.ArtFinalTaskInput }

func (r *artFinalRepoStub) Dashboard(models.ArtFinalAccess, models.ArtFinalFilters) (map[string]interface{}, error) {
	return map[string]interface{}{}, nil
}
func (r *artFinalRepoStub) CreateTask(i models.ArtFinalTaskInput, _ int) (int64, error) {
	r.created = i
	return 1, nil
}
func (r *artFinalRepoStub) UpdateTask(_ int64, i models.ArtFinalTaskInput, _ int, _ models.ArtFinalAccess) error {
	r.created = i
	return nil
}
func (r *artFinalRepoStub) CreateStory(models.ArtFinalStoryInput, int) (int64, error) { return 1, nil }
func (r *artFinalRepoStub) CheckStory(int64, bool, int) error                         { return nil }
func (r *artFinalRepoStub) DeleteStory(int64, int) error                              { return nil }
func (r *artFinalRepoStub) ListLayoutRequests(models.ArtFinalAccess) (json.RawMessage, error) {
	return json.RawMessage(`[]`), nil
}
func (r *artFinalRepoStub) GetLayoutRequest(int64, models.ArtFinalAccess) (json.RawMessage, error) {
	return json.RawMessage(`{}`), nil
}
func (r *artFinalRepoStub) CreateLayoutRequest(models.LayoutRequestInput, int, models.ArtFinalAccess) (int64, error) {
	return 1, nil
}
func (r *artFinalRepoStub) TransitionLayoutRequest(int64, string, string, int) error { return nil }
func (r *artFinalRepoStub) AddLayoutVersion(int64, models.LayoutVersionInput, int) (int64, error) {
	return 1, nil
}
func (r *artFinalRepoStub) DecideLayoutVersion(int64, string, string, int, models.ArtFinalAccess) error {
	return nil
}
func (r *artFinalRepoStub) UpsertLayoutJob(string, int64, models.LayoutJobInput, int, models.ArtFinalAccess) error {
	return nil
}
func (r *artFinalRepoStub) ConfirmProductReceived(int64, bool, int, models.ArtFinalAccess) error {
	return nil
}
func (r *artFinalRepoStub) UpdateStoryLifecycle(int64, models.StoryLifecycleInput, int) error {
	return nil
}

func TestArtFinalTaskValidation(t *testing.T) {
	r := &artFinalRepoStub{}
	s := NewArtFinalServiceWithRepository(r)
	if _, err := s.CreateTask(models.ArtFinalTaskInput{Panel: "media", Category: "layout", Title: "x"}, 1); err == nil {
		t.Fatal("expected category/panel validation")
	}
	now := time.Now()
	if _, err := s.CreateTask(models.ArtFinalTaskInput{Panel: "media", Category: "video", Title: " Video ", Priority: 3, DueAt: &now, Channel: "Instagram", Format: "Story"}, 1); err != nil {
		t.Fatal(err)
	}
	if r.created.Title != "Video" || r.created.Status != "open" {
		t.Fatalf("unexpected normalization: %+v", r.created)
	}
}

func TestLayoutRequestValidationAndStateInputs(t *testing.T) {
	s := NewArtFinalServiceWithRepository(&artFinalRepoStub{})
	_, err := s.CreateLayoutRequest(models.LayoutRequestInput{SourceType: "sale", SourceID: 1, Title: "Layout", FileMode: "separate", Items: []models.LayoutRequestItemInput{{EntityType: "sale_item", ItemID: 2}}}, 1, models.ArtFinalAccess{Sales: true})
	if err == nil {
		t.Fatal("expected missing separate file")
	}
	if err := s.TransitionLayoutRequest(1, "published", "", 1); err == nil {
		t.Fatal("expected invalid transition target")
	}
	if err := s.DecideLayoutVersion(1, models.LayoutDecisionInput{Status: "changes_requested"}, 1, models.ArtFinalAccess{}); err == nil {
		t.Fatal("expected mandatory change reason")
	}
}

func TestLayoutJobsRequireFinalFile(t *testing.T) {
	s := NewArtFinalServiceWithRepository(&artFinalRepoStub{})
	if err := s.UpsertLayoutJob("corel", 1, models.LayoutJobInput{Status: "received"}, 1, models.ArtFinalAccess{}); err == nil {
		t.Fatal("expected corel file")
	}
	if err := s.UpsertLayoutJob("engraving", 1, models.LayoutJobInput{Status: "ready"}, 1, models.ArtFinalAccess{}); err == nil {
		t.Fatal("expected engraving file")
	}
}

func TestArtFinalDashboardRejectsInvalidFilter(t *testing.T) {
	s := NewArtFinalServiceWithRepository(&artFinalRepoStub{})
	if _, err := s.Dashboard(models.ArtFinalAccess{}, models.ArtFinalFilters{Category: "invalid"}); err == nil {
		t.Fatal("expected invalid category")
	}
}
