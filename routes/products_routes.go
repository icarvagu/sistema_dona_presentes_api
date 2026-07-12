package routes

import (
	"net/http"

	"donapresentes/controllers"

	"github.com/gorilla/mux"
)

// RegisterProductsRoutes registers all product-related HTTP endpoints on the given router.
// Routes: GET /products (list), POST /products (create), GET /products/{id}, PUT /products/{id},
// DELETE /products/{id}, GET /products/pending, GET /products/financial-report, GET /products/groups,
// POST /products/approve-all, POST /products/{id}/approve, PATCH /products/{id}/last-cost.
func RegisterProductsRoutes(r *mux.Router) {
	r.HandleFunc("/products/pending", controllers.GetPendingProducts).Methods(http.MethodGet)
	r.HandleFunc("/products/financial-report", controllers.GetProductsFinancialReport).Methods(http.MethodGet)
	r.HandleFunc("/products/groups", controllers.GetProductGroups).Methods(http.MethodGet)
	r.HandleFunc("/products/approve-all", controllers.BulkApproveProducts).Methods(http.MethodPost)
	r.HandleFunc("/products/{id}/approve", controllers.ApproveProduct).Methods(http.MethodPost)
	r.HandleFunc("/products/{id}/last-cost", controllers.UpdateProductLastCost).Methods(http.MethodPatch)
	r.HandleFunc("/products/{id}", controllers.GetProduct).Methods(http.MethodGet)
	r.HandleFunc("/products", controllers.GetProducts).Methods(http.MethodGet)
	r.HandleFunc("/products", controllers.CreateProduct).Methods(http.MethodPost)
	r.HandleFunc("/products/{id}", controllers.UpdateProduct).Methods(http.MethodPut)
	r.HandleFunc("/products/{id}", controllers.DeleteProduct).Methods(http.MethodDelete)
}
