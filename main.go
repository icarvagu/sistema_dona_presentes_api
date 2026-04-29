package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

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

	if err := config.Connect(dsn); err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}

	if err := goose.Up(config.DB, "./db/migrations"); err != nil {
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

	productRepo := repositories.NewProductRepository(config.DB)
	supplierRepo := repositories.NewSupplierRepository(config.DB)
	xbzService := services.NewXBZService(os.Getenv("XBZ_CNPJ"), os.Getenv("XBZ_TOKEN"))
	syncService := services.NewSyncService(xbzService, productRepo, supplierRepo)

	scheduler := jobs.NewScheduler(syncService, xbzService)
	scheduler.RegisterJobs()
	scheduler.Start()
	defer scheduler.Stop()

	r := mux.NewRouter()
	r.StrictSlash(true)

	r.Use(middleware.CORSMiddleware)
	middleware.ApplyCORSFallbackHandlers(r)
	r.Use(middleware.NormalizePathMiddleware)
	r.Use(middleware.LoggingMiddleware)
	r.Use(middleware.RecoveryMiddleware)

	routes.RegisterAuthRoutes(r)

	authService := services.NewAuthService(repositories.NewUserRepository(config.DB))
	protectedRouter := r.PathPrefix("").Subrouter()
	protectedRouter.Use(middleware.AuthMiddleware(authService))

	routes.RegisterSuppliersRoutes(protectedRouter)
	routes.RegisterCarriersRoutes(protectedRouter)
	routes.RegisterUsersRoutes(protectedRouter)
	routes.RegisterProductsRoutes(protectedRouter)
	routes.RegisterCustomersRoutes(protectedRouter)
	routes.RegisterSalesRoutes(protectedRouter)
	routes.RegisterQuotesRoutes(protectedRouter)
	routes.RegisterAuthProtectedRoutes(protectedRouter)

	adminRouter := protectedRouter.PathPrefix("").Subrouter()
	adminRouter.Use(middleware.AdminOnlyMiddleware())

	controllers.InitProductXBZController()

	adminRouter.HandleFunc("/users", controllers.CreateUser).Methods("POST", "OPTIONS")
	adminRouter.HandleFunc("/users/{id}", controllers.DeleteUser).Methods("DELETE", "OPTIONS")
	adminRouter.HandleFunc("/products-xbz/sync", controllers.SyncProductsFromXBZHandler).Methods("POST", "OPTIONS")

	log.Printf("Server starting on %s\n", port)
	log.Fatal(http.ListenAndServe(port, r))
}
