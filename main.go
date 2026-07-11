package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"donapresentes/controllers"
	"donapresentes/controllers/config"
	"donapresentes/internal/jobs"
	"donapresentes/middleware"
	"donapresentes/repositories"
	"donapresentes/routes"
	"donapresentes/services"

	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
	"github.com/pressly/goose/v3"
)

const port = ":8080"

func bootstrap() error {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		return fmt.Errorf("DATABASE_URL environment variable is required")
	}

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		return fmt.Errorf("JWT_SECRET environment variable is required")
	}

	if err := config.Connect(dsn); err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}

	if err := goose.Up(config.DB, "./db/migrations", goose.WithAllowMissing()); err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}
	log.Println("Connected to the database successfully!")
	return nil
}

func main() {
	if err := bootstrap(); err != nil {
		log.Fatalf("server bootstrap failed: %v", err)
	}

	defer config.DB.Close()

	controllers.InitAuthService()
	controllers.InitSupplierService()
	controllers.InitCarrierService()
	controllers.InitUserService()
	controllers.InitProductService()
	controllers.InitCustomerService()
	controllers.InitSaleService()
	controllers.InitQuoteService()
	controllers.InitSalesWorkflowService()
	controllers.InitArtFinalService()
	controllers.InitProductionService()
	controllers.InitDashboardService(services.NewDashboardService(repositories.NewDashboardRepository(config.DB)))
	controllers.InitPurchaseService()

	productRepo := repositories.NewProductRepository(config.DB)
	supplierRepo := repositories.NewSupplierRepository(config.DB)
	xbzService := services.NewXBZService(os.Getenv("XBZ_CNPJ"), os.Getenv("XBZ_TOKEN"))
	syncService := services.NewSyncService(xbzService, productRepo, supplierRepo, services.NewAuditService(config.DB))

	scheduler := jobs.NewScheduler(syncService, xbzService)
	scheduler.RegisterJobs()
	scheduler.Start()
	defer scheduler.Stop()

	requestLogRepo := repositories.NewRequestLogRepository(config.DB)
	requestLoggerSvc := services.NewRequestLoggerService(requestLogRepo)
	defer requestLoggerSvc.Shutdown()
	middleware.InitRequestLogger(requestLoggerSvc)

	errorLogRepo := repositories.NewErrorLogRepository(config.DB)
	middleware.InitErrorLogRepo(errorLogRepo)

	config.DB.Exec("SELECT cleanup_old_logs()")

	r := mux.NewRouter()
	r.StrictSlash(true)

	r.Use(middleware.TracingMiddleware)
	r.Use(middleware.CORSMiddleware)
	middleware.ApplyCORSFallbackHandlers(r)
	r.Use(middleware.SecurityHeadersMiddleware)
	r.Use(middleware.InputValidationMiddleware)
	r.Use(middleware.RateLimitMiddlewarePerRoute)
	r.Use(middleware.TimeoutMiddleware(30 * time.Second))
	r.Use(middleware.NormalizePathMiddleware)
	r.Use(middleware.StructuredLoggingMiddleware)
	r.Use(middleware.RecoveryMiddleware)

	r.HandleFunc("/health", controllers.HealthCheck).Methods("GET", "OPTIONS")
	routes.RegisterAuthRoutes(r)

	authService := services.NewAuthService(
		repositories.NewUserRepository(config.DB),
		repositories.NewRefreshTokenRepository(config.DB),
		services.NewAuditService(config.DB),
	)
	protectedRouter := r.PathPrefix("").Subrouter()
	protectedRouter.Use(middleware.AuthMiddleware(authService))

	routes.RegisterSuppliersRoutes(protectedRouter)
	routes.RegisterCarriersRoutes(protectedRouter)
	routes.RegisterUsersRoutes(protectedRouter)
	routes.RegisterProductsRoutes(protectedRouter)
	routes.RegisterCustomersRoutes(protectedRouter)
	routes.RegisterSalesRoutes(protectedRouter)
	routes.RegisterQuotesRoutes(protectedRouter)
	routes.RegisterSalesWorkflowRoutes(protectedRouter)
	routes.RegisterArtFinalRoutes(protectedRouter)
	routes.RegisterProductionRoutes(protectedRouter)
	routes.RegisterDashboardRoutes(protectedRouter)
	routes.RegisterPurchasesRoutes(protectedRouter)
	routes.RegisterAuthProtectedRoutes(protectedRouter)

	adminRouter := protectedRouter.PathPrefix("").Subrouter()
	adminRouter.Use(middleware.AdminOnlyMiddleware())

	controllers.InitProductXBZController()

	adminRouter.HandleFunc("/users", controllers.CreateUser).Methods("POST", "OPTIONS")
	adminRouter.HandleFunc("/users/{id}", controllers.DeleteUser).Methods("DELETE", "OPTIONS")
	adminRouter.HandleFunc("/products-xbz/sync", controllers.SyncProductsFromXBZHandler).Methods("POST", "OPTIONS")

	tlsCert := os.Getenv("TLS_CERT_FILE")
	tlsKey := os.Getenv("TLS_KEY_FILE")

	if tlsCert != "" && tlsKey != "" {
		log.Printf("Server starting on %s with TLS\n", port)
		log.Fatal(http.ListenAndServeTLS(port, tlsCert, tlsKey, r))
	} else {
		log.Printf("Server starting on %s (plain HTTP - configure TLS for production)\n", port)
		log.Fatal(http.ListenAndServe(port, r))
	}
}
