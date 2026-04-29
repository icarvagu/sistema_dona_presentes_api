package main

import (
	"fmt"
	"log"
	"net/http"

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

const (
	dsn  = "postgres://postgres:123456@postgres:5432/crud?sslmode=disable"
	port = ":8080"
)

func init() {
	var err error
	err = config.Connect(dsn)
	if err != nil {
		panic(fmt.Sprintf("Failed to connect to database: %v", err))
	}

	if err := goose.Up(config.DB, "./db/migrations"); err != nil {
		panic(fmt.Sprintf("Failed to run migrations: %v", err))
	}
	log.Println("Connected to the database successfully!")
}

func main() {
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
	xbzService := services.NewXBZService("36168035000181", "X142AA979C")
	syncService := services.NewSyncService(xbzService, productRepo, supplierRepo)

	scheduler := jobs.NewScheduler(syncService, xbzService)
	scheduler.RegisterJobs()
	scheduler.Start()
	defer scheduler.Stop()

	r := mux.NewRouter()
	r.StrictSlash(true)

	r.Use(middleware.CORSMiddleware)
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

	adminRouter := protectedRouter.PathPrefix("").Subrouter()
	adminRouter.Use(middleware.AdminOnlyMiddleware())

	controllers.InitProductXBZController()

	adminRouter.HandleFunc("/users", controllers.CreateUser).Methods("POST", "OPTIONS")
	adminRouter.HandleFunc("/users/{id}", controllers.DeleteUser).Methods("DELETE", "OPTIONS")
	adminRouter.HandleFunc("/products-xbz/sync", controllers.SyncProductsFromXBZ).Methods("POST", "OPTIONS")

	log.Printf("Server starting on %s\n", port)
	log.Fatal(http.ListenAndServe(port, r))
}
