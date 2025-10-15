package main

import (
	"log"
	"net/http"

	"github.com/gorilla/mux"
	httpSwagger "github.com/swaggo/http-swagger"
	_ "golang-backend/microservices/auth-service/docs"
	"golang-backend/microservices/shared/config"
	"golang-backend/microservices/shared/database"
	"golang-backend/microservices/auth-service/handlers"
)

// @title Auth Service API
// @version 1.0
// @description Authentication service for user registration and login
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host localhost:8081
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

	// Auth routes
	r.HandleFunc("/register", handlers.Register(cfg)).Methods("POST")
	r.HandleFunc("/login", handlers.Login(cfg)).Methods("POST")
	r.HandleFunc("/admin/register", handlers.AdminRegister(cfg)).Methods("POST")
	r.HandleFunc("/admin/login", handlers.AdminLogin(cfg)).Methods("POST")

	// Health check
	r.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Auth Service is healthy"))
	}).Methods("GET")

	// Swagger route
	r.PathPrefix("/swagger/").Handler(httpSwagger.WrapHandler)

	log.Println("Auth Service starting on :8081")
	log.Fatal(http.ListenAndServe(":8081", r))
}
