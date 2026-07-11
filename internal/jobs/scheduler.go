package jobs

import (
	"log"
	"time"

	"donapresentes/repositories"
	"donapresentes/services"

	cron "github.com/robfig/cron/v3"
)

type Scheduler struct {
	cron       *cron.Cron
	syncSvc    *services.SyncService
	xbzService *services.XBZService
}

func NewScheduler(sync *services.SyncService, xbz *services.XBZService) *Scheduler {
	return &Scheduler{
		cron:       cron.New(),
		syncSvc:    sync,
		xbzService: xbz,
	}
}

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

func (s *Scheduler) Start() {
	log.Printf("[scheduler] Start() chamado - iniciando cron")
	s.cron.Start()
}

func (s *Scheduler) Stop() {
	s.cron.Stop()
}

func NewSyncServiceWithRepos(xbz *services.XBZService, productRepo *repositories.ProductRepository, supplierRepo *repositories.SupplierRepository, auditService *services.AuditService) *services.SyncService {
	return services.NewSyncService(xbz, productRepo, supplierRepo, auditService)
}
