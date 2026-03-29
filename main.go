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

	// Initialize services after database is connected
	controllers.InitAuthService()
	controllers.InitSupplierService()
	controllers.InitCarrierService()
	controllers.InitUserService()
	controllers.InitProductService()
	controllers.InitCustomerService()
	controllers.InitSaleService()
	controllers.InitQuoteService()

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
	r.StrictSlash(true)

	// Aplicar middlewares globais (CORS deve ser o primeiro)
	r.Use(middleware.CORSMiddleware)
	r.Use(middleware.NormalizePathMiddleware)
	r.Use(middleware.LoggingMiddleware)
	r.Use(middleware.RecoveryMiddleware)

	// Rotas públicas (sem autenticação)
	routes.RegisterAuthRoutes(r)

	// Criar router protegido com middleware de autenticação
	authService := services.NewAuthService(repositories.NewUserRepository(config.DB))
	protectedRouter := r.PathPrefix("").Subrouter()
	protectedRouter.Use(middleware.AuthMiddleware(authService))

	// Rotas protegidas (requerem autenticação)
	routes.RegisterSuppliersRoutes(protectedRouter)

	// Register carriers routes
	routes.RegisterCarriersRoutes(protectedRouter)

	// Register users routes
	routes.RegisterUsersRoutes(protectedRouter)

	// Register products routes
	routes.RegisterProductsRoutes(protectedRouter)

	// Register customers routes
	routes.RegisterCustomersRoutes(protectedRouter)

	// Register sales routes
	routes.RegisterSalesRoutes(protectedRouter)

	// Register quotes routes
	routes.RegisterQuotesRoutes(protectedRouter)

	// Criar router para rotas administrativas (apenas admin)
	adminRouter := protectedRouter.PathPrefix("").Subrouter()
	adminRouter.Use(middleware.AdminOnlyMiddleware())
	
	// Inicializar controller XBZ
	controllers.InitProductXBZController()
	
	// Rotas administrativas
	adminRouter.HandleFunc("/users", controllers.CreateUser).Methods("POST", "OPTIONS")
	adminRouter.HandleFunc("/users/{id}", controllers.DeleteUser).Methods("DELETE", "OPTIONS")
	adminRouter.HandleFunc("/products-xbz/sync", controllers.SyncProductsFromXBZ).Methods("POST", "OPTIONS")

	// Wrapper HTTP handler para garantir CORS em todas as rotas, incluindo 404 e rotas não encontradas
	handler := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		// Aplicar CORS antes de processar qualquer requisição
		origin := req.Header.Get("Origin")
		if origin != "" {
			w.Header().Set("Access-Control-Allow-Origin", origin)
		} else {
			w.Header().Set("Access-Control-Allow-Origin", "*")
		}

		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, PATCH, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With, Accept, Origin, X-CSRF-Token")
		w.Header().Set("Access-Control-Expose-Headers", "Content-Length, Content-Type, Location")
		w.Header().Set("Access-Control-Max-Age", "3600")
		w.Header().Set("Access-Control-Allow-Credentials", "false")

		// Responder preflight imediatamente (deve ser antes de qualquer processamento)
		if req.Method == "OPTIONS" {
			log.Printf("[CORS] Preflight request para %s", req.URL.Path)
			w.WriteHeader(http.StatusNoContent)
			return
		}

		r.ServeHTTP(w, req)
	})

	log.Printf("Server starting on %s\n", port)
	log.Fatal(http.ListenAndServe(port, handler))
}
