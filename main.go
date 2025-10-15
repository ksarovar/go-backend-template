package main

import (
	"log"
	"net/http"

	"github.com/gorilla/mux"
	httpSwagger "github.com/swaggo/http-swagger"
	_ "golang-backend/docs"
	"golang-backend/config"
	"golang-backend/database"
	"golang-backend/handlers"
	"golang-backend/middleware"
)

// @title Golang Backend API
// @version 1.0
// @description A Golang backend with user authentication, encryption, and MongoDB
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host localhost:8080
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

	// Admin auth routes
	r.HandleFunc("/admin/register", handlers.AdminRegister(cfg)).Methods("POST")
	r.HandleFunc("/admin/login", handlers.AdminLogin(cfg)).Methods("POST")

	// Protected routes
	protected := r.PathPrefix("/").Subrouter()
	protected.Use(middleware.JWTAuthMiddleware(cfg))

	// User routes
	protected.HandleFunc("/user/profile", handlers.GetUserProfile).Methods("GET")
	protected.HandleFunc("/user/profile", handlers.UpdateUserProfile).Methods("PUT")

	// Admin routes
	admin := r.PathPrefix("/admin").Subrouter()
	admin.Use(middleware.JWTAuthMiddleware(cfg))
	admin.HandleFunc("/users", handlers.ListUsers).Methods("GET")
	admin.HandleFunc("/users/delete", handlers.DeleteUser).Methods("POST")
	admin.HandleFunc("/users/role", handlers.UpdateUserRole).Methods("PUT")

	// Swagger route
	r.PathPrefix("/swagger/").Handler(httpSwagger.WrapHandler)

	log.Println("Server starting on :8080")
	log.Fatal(http.ListenAndServe(":8080", r))
}
