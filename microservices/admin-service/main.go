package main

import (
	"log"
	"net/http"

	"github.com/gorilla/mux"
	httpSwagger "github.com/swaggo/http-swagger"
	_ "golang-backend/microservices/admin-service/docs"
	"golang-backend/microservices/shared/config"
	"golang-backend/microservices/shared/database"
	"golang-backend/microservices/admin-service/handlers"
	"golang-backend/microservices/admin-service/middleware"
)

// @title Admin Service API
// @version 1.0
// @description Admin service for user management operations
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host localhost:8083
// @BasePath /

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
func main() {
	// Load configuration
	cfg := config.Load()

	// Connect to database
	database.Connect(cfg.MongoURI)

	// Create router
	r := mux.NewRouter()

	// Apply authentication and admin middleware to all routes
	r.Use(middleware.JWTAuthMiddleware(cfg))
	r.Use(middleware.AdminOnlyMiddleware)

	// Admin routes
	r.HandleFunc("/users", handlers.ListUsers).Methods("GET")
	r.HandleFunc("/users/{id}", handlers.DeleteUser).Methods("DELETE")
	r.HandleFunc("/users/{id}/role", handlers.UpdateUserRole).Methods("PUT")

	// Health check
	r.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Admin Service is healthy"))
	}).Methods("GET")

	// Swagger route
	r.PathPrefix("/swagger/").Handler(httpSwagger.WrapHandler)

	log.Println("Admin Service starting on :8083")
	log.Fatal(http.ListenAndServe(":8083", r))
}
