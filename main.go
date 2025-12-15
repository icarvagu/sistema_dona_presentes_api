package main

import (
	"database/sql"
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

var db *sql.DB

func init() {
	// Initialize database connection
	var err error
	err = config.Connect(dsn)
	if err != nil {
		panic(fmt.Sprintf("Failed to connect to database: %v", err))
	}

	// Run migrations
	if err := goose.Up(config.DB, "./db/migrations"); err != nil {
		panic(fmt.Sprintf("Failed to run migrations: %v", err))
	}
	fmt.Println("Connected to the database successfully!")
	log.Printf("Connected to the database successfully! log")

}

func main() {
	defer config.DB.Close()

	// Initialize repositories after database is connected
	controllers.InitSupplierRepository()
	controllers.InitCarrierRepository()
	controllers.InitEmployeeRepository()
	controllers.InitProductRepository()
	controllers.InitCustomerRepository()
	controllers.InitSaleRepository()

	// Initialize services and scheduler
	productRepo := repositories.NewProductRepository(config.DB)
	supplierRepo := repositories.NewSupplierRepository(config.DB)
	xbzService := services.NewXBZService("36168035000181", "X142AA979C")
	syncService := services.NewSyncService(xbzService, productRepo, supplierRepo)

	scheduler := jobs.NewScheduler(syncService, xbzService)
	scheduler.RegisterJobs()
	scheduler.Start()
	defer scheduler.Stop()

	r := mux.NewRouter()

	// Aplicar middlewares globais
	r.Use(middleware.LoggingMiddleware)
	r.Use(middleware.RecoveryMiddleware)

	// Register suppliers routes
	routes.RegisterSuppliersRoutes(r)

	// Register carriers routes
	routes.RegisterCarriersRoutes(r)

	// Register employees routes
	routes.RegisterEmployeesRoutes(r)

	// Register products routes
	routes.RegisterProductsRoutes(r)

	// Register customers routes
	routes.RegisterCustomersRoutes(r)

	// Register sales routes
	routes.RegisterSalesRoutes(r)

	// Register products XBZ routes
	routes.RegisterProductsXBZRoutes(r)

	log.Printf("Server starting on %s\n", port)
	log.Fatal(http.ListenAndServe(port, r))
}
