package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"
	"golang-backend/microservices/shared/config"
	"golang-backend/microservices/shared/database"
	"golang-backend/microservices/shared/models"
	"golang-backend/microservices/shared/utils"
)

// RegisterRequest represents the request payload for user registration
type RegisterRequest struct {
	Email    string `json:"email" example:"user@example.com"`
	Password string `json:"password" example:"password123"`
	Role     string `json:"role,omitempty" example:"user"`
}

// LoginRequest represents the request payload for user login
type LoginRequest struct {
	Email    string `json:"email" example:"user@example.com"`
	Password string `json:"password" example:"password123"`
}

// RegisterResponse represents the response for user registration
type RegisterResponse struct {
	Message string `json:"message" example:"User registered successfully"`
}

// LoginResponse represents the response for user login
type LoginResponse struct {
	Token string `json:"token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
	Role  string `json:"role" example:"user"`
}

// AdminRegisterRequest represents the request payload for admin user registration
type AdminRegisterRequest struct {
	Email    string `json:"email" example:"admin@example.com"`
	Password string `json:"password" example:"admin123"`
}

// AdminLoginRequest represents the request payload for admin login
type AdminLoginRequest struct {
	Email    string `json:"email" example:"admin@example.com"`
	Password string `json:"password" example:"admin123"`
}

// AdminLoginResponse represents the response for admin login
type AdminLoginResponse struct {
	Token string `json:"token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
	Role  string `json:"role" example:"admin"`
}

// Register handles user registration
// @Summary Register a new user
// @Description Register a new user with email and password
// @Tags auth
// @Accept json
// @Produce json
// @Param request body RegisterRequest true "User registration data"
// @Success 200 {object} RegisterResponse
// @Failure 400 {string} string "Invalid request payload"
// @Failure 409 {string} string "User already exists"
// @Failure 500 {string} string "Internal server error"
// @Router /register [post]
func Register(cfg *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req RegisterRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request payload", http.StatusBadRequest)
			return
		}

		collection := database.GetCollection("users")
		ctx := context.Background()

		// Check if user already exists
		var existingUser models.User
		err := collection.FindOne(ctx, bson.M{"email_hash": req.Email}).Decode(&existingUser)
		if err == nil {
			http.Error(w, "User already exists", http.StatusConflict)
			return
		} else if err != mongo.ErrNoDocuments {
			http.Error(w, "Database error", http.StatusInternalServerError)
			return
		}

		// Hash the password
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {
			http.Error(w, "Failed to hash password", http.StatusInternalServerError)
			return
		}

		// Encrypt email
		encryptedEmail, err := utils.Encrypt(req.Email, cfg.EncryptionKey)
		if err != nil {
			http.Error(w, "Failed to encrypt data", http.StatusInternalServerError)
			return
		}

		// Create email hash for lookup
		emailHash := req.Email

		// Determine role (default to "user" if not specified or invalid)
		role := "user"
		if req.Role == "admin" {
			role = "admin"
		}

		// Create new user
		now := time.Now()
		user := models.User{
			ID:        primitive.NewObjectID(),
			EmailHash: emailHash,
			Email:     encryptedEmail,
			Password:  string(hashedPassword),
			Role:      role,
			CreatedAt: now,
			UpdatedAt: now,
		}

		_, err = collection.InsertOne(ctx, user)
		if err != nil {
			http.Error(w, "Failed to create user", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"message": "User registered successfully"})
	}
}

// Login handles user login
// @Summary Login user
// @Description Login with email and password to get JWT token
// @Tags auth
// @Accept json
// @Produce json
// @Param request body LoginRequest true "User login data"
// @Success 200 {object} LoginResponse
// @Failure 400 {string} string "Invalid request payload"
// @Failure 401 {string} string "Invalid credentials"
// @Failure 500 {string} string "Internal server error"
// @Router /login [post]
func Login(cfg *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req LoginRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request payload", http.StatusBadRequest)
			return
		}

		collection := database.GetCollection("users")
		ctx := context.Background()

		// Find user by email hash
		var user models.User
		err := collection.FindOne(ctx, bson.M{"email_hash": req.Email}).Decode(&user)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				http.Error(w, "Invalid credentials", http.StatusUnauthorized)
			} else {
				http.Error(w, "Database error", http.StatusInternalServerError)
			}
			return
		}

		// Check password
		if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
			http.Error(w, "Invalid credentials", http.StatusUnauthorized)
			return
		}

		// Decrypt email for JWT
		decryptedEmail, err := utils.Decrypt(user.Email, cfg.EncryptionKey)
		if err != nil {
			http.Error(w, "Failed to decrypt data", http.StatusInternalServerError)
			return
		}

		// Generate JWT token
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
			"userID": user.ID.Hex(),
			"email":  decryptedEmail,
			"role":   user.Role,
			"exp":    time.Now().Add(time.Hour * 24).Unix(),
		})

		tokenString, err := token.SignedString([]byte(cfg.JWTSecret))
		if err != nil {
			http.Error(w, "Failed to generate token", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"token": tokenString,
			"role":  user.Role,
		})
	}
}

// AdminRegister handles admin user registration
// @Summary Register a new admin user
// @Description Register a new admin user with email and password
// @Tags admin
// @Accept json
// @Produce json
// @Param request body AdminRegisterRequest true "Admin registration data"
// @Success 200 {object} RegisterResponse
// @Failure 400 {string} string "Invalid request payload"
// @Failure 409 {string} string "Admin already exists"
// @Failure 500 {string} string "Internal server error"
// @Router /admin/register [post]
func AdminRegister(cfg *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req AdminRegisterRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request payload", http.StatusBadRequest)
			return
		}

		collection := database.GetCollection("users")
		ctx := context.Background()

		// Check if admin already exists
		var existingUser models.User
		err := collection.FindOne(ctx, bson.M{"email_hash": req.Email}).Decode(&existingUser)
		if err == nil {
			http.Error(w, "Admin already exists", http.StatusConflict)
			return
		} else if err != mongo.ErrNoDocuments {
			http.Error(w, "Database error", http.StatusInternalServerError)
			return
		}

		// Hash the password
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {
			http.Error(w, "Failed to hash password", http.StatusInternalServerError)
			return
		}

		// Encrypt email
		encryptedEmail, err := utils.Encrypt(req.Email, cfg.EncryptionKey)
		if err != nil {
			http.Error(w, "Failed to encrypt data", http.StatusInternalServerError)
			return
		}

		// Create email hash for lookup
		emailHash := req.Email

		// Create new admin user
		now := time.Now()
		user := models.User{
			ID:        primitive.NewObjectID(),
			EmailHash: emailHash,
			Email:     encryptedEmail,
			Password:  string(hashedPassword),
			Role:      "admin",
			CreatedAt: now,
			UpdatedAt: now,
		}

		_, err = collection.InsertOne(ctx, user)
		if err != nil {
			http.Error(w, "Failed to create admin", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"message": "Admin registered successfully"})
	}
}

// AdminLogin handles admin login
// @Summary Admin login
// @Description Login with admin email and password to get JWT token
// @Tags admin
// @Accept json
// @Produce json
// @Param request body AdminLoginRequest true "Admin login data"
// @Success 200 {object} AdminLoginResponse
// @Failure 400 {string} string "Invalid request payload"
// @Failure 401 {string} string "Invalid credentials"
// @Failure 403 {string} string "Access denied: Admin only"
// @Failure 500 {string} string "Internal server error"
// @Router /admin/login [post]
func AdminLogin(cfg *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req AdminLoginRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request payload", http.StatusBadRequest)
			return
		}

		collection := database.GetCollection("users")
		ctx := context.Background()

		// Find user by email hash
		var user models.User
		err := collection.FindOne(ctx, bson.M{"email_hash": req.Email}).Decode(&user)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				http.Error(w, "Invalid credentials", http.StatusUnauthorized)
			} else {
				http.Error(w, "Database error", http.StatusInternalServerError)
			}
			return
		}

		// Check if user is admin
		if user.Role != "admin" {
			http.Error(w, "Access denied: Admin only", http.StatusForbidden)
			return
		}

		// Check password
		if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
			http.Error(w, "Invalid credentials", http.StatusUnauthorized)
			return
		}

		// Decrypt email for JWT
		decryptedEmail, err := utils.Decrypt(user.Email, cfg.EncryptionKey)
		if err != nil {
			http.Error(w, "Failed to decrypt data", http.StatusInternalServerError)
			return
		}

		// Generate JWT token
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
			"userID": user.ID.Hex(),
			"email":  decryptedEmail,
			"role":   user.Role,
			"exp":    time.Now().Add(time.Hour * 24).Unix(),
		})

		tokenString, err := token.SignedString([]byte(cfg.JWTSecret))
		if err != nil {
			http.Error(w, "Failed to generate token", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"token": tokenString,
			"role":  user.Role,
		})
	}
}
