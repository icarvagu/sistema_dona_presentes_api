// Package jobs provides background job scheduling for the Dona Presentes API.
package jobs

import (
	"log"
	"time"

	"donapresentes/repositories"
	"donapresentes/services"

	cron "github.com/robfig/cron/v3"
)

// Scheduler manages background cron jobs such as periodic data
// synchronization with external services.
type Scheduler struct {
	cron       *cron.Cron
	syncSvc    *services.SyncService
	xbzService *services.XBZService
}

// NewScheduler creates a Scheduler configured with the given sync and
// XBZ services.
func NewScheduler(sync *services.SyncService, xbz *services.XBZService) *Scheduler {
	return &Scheduler{
		cron:       cron.New(),
		syncSvc:    sync,
		xbzService: xbz,
	}
}

// RegisterJobs registers all recurring jobs with the internal cron
// scheduler. Currently schedules a sync job that runs daily from
// Sunday through Friday at 23:22.
func (s *Scheduler) RegisterJobs() {
	log.Printf("Entrou no RegisterJobs do Scheduler")

	if _, err := s.cron.AddFunc("22 23 * * 0-5", func() {
		go func() {
			start := time.Now()
			log.Printf("[scheduler] iniciando sync job em %s", start.Format(time.RFC3339))
			result, err := s.syncSvc.Synchronize()
			duration := time.Since(start)
			if err != nil {
				log.Printf("[scheduler] erro na sincronização agendada: %v (duracao=%s)", err, duration)
				return
			}
			log.Printf("[scheduler] sincronização concluída: criados=%d atualizados=%d erros=%d total=%d (duracao=%s)", result.Criados, result.Atualizados, result.Erros, result.Total, duration)
		}()
	}); err != nil {
		log.Printf("[scheduler] falha ao agendar job de sync: %v", err)
	}

}

// Start begins executing all registered cron jobs.
func (s *Scheduler) Start() {
	log.Printf("[scheduler] Start() chamado - iniciando cron")
	s.cron.Start()
}

// Stop halts the cron scheduler, preventing any further job executions.
func (s *Scheduler) Stop() {
	s.cron.Stop()
}

// NewSyncServiceWithRepos is a convenience constructor that creates a
// SyncService wired with the provided repositories and audit service.
func NewSyncServiceWithRepos(xbz *services.XBZService, productRepo *repositories.ProductRepository, supplierRepo *repositories.SupplierRepository, auditService *services.AuditService) *services.SyncService {
	return services.NewSyncService(xbz, productRepo, supplierRepo, auditService)
}
