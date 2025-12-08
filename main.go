package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"

	"donapresentes/controllers"
	"donapresentes/controllers/config"
	"donapresentes/routes"

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
}

func main() {
	defer config.DB.Close()

	// Initialize repositories after database is connected
	controllers.InitFornecedorRepository()
	controllers.InitTransportadoraRepository()
	controllers.InitFuncionarioRepository()
	controllers.InitProdutoRepository()
	controllers.InitClienteRepository()

	r := mux.NewRouter()

	// Register fornecedores routes
	routes.RegisterFornecedoresRoutes(r)

	// Register transportadora routes
	routes.RegisterTransportadoraRoutes(r)

	// Register funcionarios routes
	routes.RegisterFuncionariosRoutes(r)

	// Register produtos routes
	routes.RegisterProdutosRoutes(r)

	// Register clientes routes
	routes.RegisterClientesRoutes(r)

	log.Printf("Server starting on %s\n", port)
	log.Fatal(http.ListenAndServe(port, r))
}
