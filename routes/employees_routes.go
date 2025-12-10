package routes

import (
	"donapresentes/controllers"

	"github.com/gorilla/mux"
)

func RegisterEmployeesRoutes(r *mux.Router) {
	r.HandleFunc("/employees", controllers.GetEmployees).Methods("GET")
	r.HandleFunc("/employees/{id}", controllers.GetEmployee).Methods("GET")
	r.HandleFunc("/employees", controllers.CreateEmployee).Methods("POST")
	r.HandleFunc("/employees/{id}", controllers.UpdateEmployee).Methods("PUT")
	r.HandleFunc("/employees/{id}", controllers.DeleteEmployee).Methods("DELETE")
}
