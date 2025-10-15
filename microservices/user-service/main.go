package main

import (
	"log"
	"net/http"

	"github.com/gorilla/mux"
	httpSwagger "github.com/swaggo/http-swagger"
	_ "golang-backend/microservices/user-service/docs"
	"golang-backend/microservices/shared/config"
	"golang-backend/microservices/shared/database"
	"golang-backend/microservices/user-service/handlers"
	"golang-backend/microservices/user-service/middleware"
)

// @title User Service API
// @version 1.0
// @description User service for profile management
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host localhost:8082
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

	// Apply authentication middleware to all routes
	r.Use(middleware.JWTAuthMiddleware(cfg))

	// User routes
	r.HandleFunc("/profile", handlers.GetUserProfile).Methods("GET")
	r.HandleFunc("/profile", handlers.UpdateUserProfile).Methods("PUT")

	// Health check
	r.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("User Service is healthy"))
	}).Methods("GET")

	// Swagger route
	r.PathPrefix("/swagger/").Handler(httpSwagger.WrapHandler)

	log.Println("User Service starting on :8082")
	log.Fatal(http.ListenAndServe(":8082", r))
}
