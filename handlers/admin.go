package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"golang.org/x/crypto/bcrypt"
	"golang-backend/config"
	"golang-backend/database"
	"golang-backend/models"
	"golang-backend/utils"
)

// ListUsersRequest represents the request for listing users
type ListUsersRequest struct {
	Page  int `json:"page,omitempty"`
	Limit int `json:"limit,omitempty"`
}

// ListUsersResponse represents the response for listing users
type ListUsersResponse struct {
	Users      []UserResponse `json:"users"`
	Total      int            `json:"total"`
	Page       int            `json:"page"`
	Limit      int            `json:"limit"`
	TotalPages int            `json:"total_pages"`
}

// UserResponse represents a user in the response
type UserResponse struct {
	ID        string    `json:"id"`
	Email     string    `json:"email"`
	Role      string    `json:"role"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// DeleteUserRequest represents the request for deleting a user
type DeleteUserRequest struct {
	UserID string `json:"user_id"`
}

// UpdateUserRoleRequest represents the request for updating user role
type UpdateUserRoleRequest struct {
	UserID string `json:"user_id"`
	Role   string `json:"role"`
}

// @Summary List all users
// @Description Get a paginated list of all users (Admin only)
// @Tags admin
// @Accept json
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(10)
// @Security BearerAuth
// @Success 200 {object} ListUsersResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /admin/users [get]
func ListUsers(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Get user claims from context
	claims := r.Context().Value("claims").(jwt.MapClaims)
	userRole := claims["role"].(string)

	if userRole != "admin" {
		http.Error(w, `{"error": "Forbidden: Admin access required"}`, http.StatusForbidden)
		return
	}

	// Parse query parameters
	page := 1
	limit := 10

	if p := r.URL.Query().Get("page"); p != "" {
		if parsed, err := strconv.Atoi(p); err == nil && parsed > 0 {
			page = parsed
		}
	}

	if l := r.URL.Query().Get("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 && parsed <= 100 {
			limit = parsed
		}
	}

	skip := (page - 1) * limit

	// Get users from database
	collection := database.DB.Collection("users")
	ctx := context.Background()

	// Count total users
	total, err := collection.CountDocuments(ctx, bson.M{})
	if err != nil {
		http.Error(w, `{"error": "Failed to count users"}`, http.StatusInternalServerError)
		return
	}

	// Find users with pagination
	opts := options.Find().SetSkip(int64(skip)).SetLimit(int64(limit)).SetSort(bson.M{"created_at": -1})
	cursor, err := collection.Find(ctx, bson.M{}, opts)
	if err != nil {
		http.Error(w, `{"error": "Failed to fetch users"}`, http.StatusInternalServerError)
		return
	}
	defer cursor.Close(ctx)

	var users []models.User
	if err = cursor.All(ctx, &users); err != nil {
		http.Error(w, `{"error": "Failed to decode users"}`, http.StatusInternalServerError)
		return
	}

	// Decrypt emails and prepare response
	var userResponses []UserResponse
	for _, user := range users {
		decryptedEmail, err := utils.Decrypt(user.Email, config.Load().EncryptionKey)
		if err != nil {
			http.Error(w, `{"error": "Failed to decrypt user data"}`, http.StatusInternalServerError)
			return
		}

		userResponses = append(userResponses, UserResponse{
			ID:        user.ID.Hex(),
			Email:     decryptedEmail,
			Role:      user.Role,
			CreatedAt: user.CreatedAt,
			UpdatedAt: user.UpdatedAt,
		})
	}

	totalPages := (int(total) + limit - 1) / limit

	response := ListUsersResponse{
		Users:      userResponses,
		Total:      int(total),
		Page:       page,
		Limit:      limit,
		TotalPages: totalPages,
	}

	json.NewEncoder(w).Encode(response)
}

// @Summary Delete a user
// @Description Delete a user by ID (Admin only)
// @Tags admin
// @Accept json
// @Produce json
// @Param request body DeleteUserRequest true "User deletion request"
// @Security BearerAuth
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /admin/users/delete [post]
func DeleteUser(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Get user claims from context
	claims := r.Context().Value("claims").(jwt.MapClaims)
	userRole := claims["role"].(string)

	if userRole != "admin" {
		http.Error(w, `{"error": "Forbidden: Admin access required"}`, http.StatusForbidden)
		return
	}

	var req DeleteUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error": "Invalid request body"}`, http.StatusBadRequest)
		return
	}

	if req.UserID == "" {
		http.Error(w, `{"error": "User ID is required"}`, http.StatusBadRequest)
		return
	}

	userID, err := primitive.ObjectIDFromHex(req.UserID)
	if err != nil {
		http.Error(w, `{"error": "Invalid user ID format"}`, http.StatusBadRequest)
		return
	}

	collection := database.DB.Collection("users")
	ctx := context.Background()

	result, err := collection.DeleteOne(ctx, bson.M{"_id": userID})
	if err != nil {
		http.Error(w, `{"error": "Failed to delete user"}`, http.StatusInternalServerError)
		return
	}

	if result.DeletedCount == 0 {
		http.Error(w, `{"error": "User not found"}`, http.StatusNotFound)
		return
	}

	json.NewEncoder(w).Encode(SuccessResponse{Message: "User deleted successfully"})
}

// @Summary Update user role
// @Description Update a user's role (Admin only)
// @Tags admin
// @Accept json
// @Produce json
// @Param request body UpdateUserRoleRequest true "User role update request"
// @Security BearerAuth
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /admin/users/role [put]
func UpdateUserRole(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Get user claims from context
	claims := r.Context().Value("claims").(jwt.MapClaims)
	userRole := claims["role"].(string)

	if userRole != "admin" {
		http.Error(w, `{"error": "Forbidden: Admin access required"}`, http.StatusForbidden)
		return
	}

	var req UpdateUserRoleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error": "Invalid request body"}`, http.StatusBadRequest)
		return
	}

	if req.UserID == "" || req.Role == "" {
		http.Error(w, `{"error": "User ID and role are required"}`, http.StatusBadRequest)
		return
	}

	if req.Role != "user" && req.Role != "admin" {
		http.Error(w, `{"error": "Invalid role. Must be 'user' or 'admin'"}`, http.StatusBadRequest)
		return
	}

	userID, err := primitive.ObjectIDFromHex(req.UserID)
	if err != nil {
		http.Error(w, `{"error": "Invalid user ID format"}`, http.StatusBadRequest)
		return
	}

	collection := database.DB.Collection("users")
	ctx := context.Background()

	update := bson.M{
		"$set": bson.M{
			"role":       req.Role,
			"updated_at": time.Now(),
		},
	}

	result, err := collection.UpdateOne(ctx, bson.M{"_id": userID}, update)
	if err != nil {
		http.Error(w, `{"error": "Failed to update user role"}`, http.StatusInternalServerError)
		return
	}

	if result.MatchedCount == 0 {
		http.Error(w, `{"error": "User not found"}`, http.StatusNotFound)
		return
	}

	json.NewEncoder(w).Encode(SuccessResponse{Message: "User role updated successfully"})
}

// @Summary Get user profile
// @Description Get current user's profile information
// @Tags user
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} UserResponse
// @Failure 401 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /user/profile [get]
func GetUserProfile(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Get user claims from context
	claims := r.Context().Value("claims").(jwt.MapClaims)
	userIDStr := claims["userID"].(string)

	userID, err := primitive.ObjectIDFromHex(userIDStr)
	if err != nil {
		http.Error(w, `{"error": "Invalid user ID"}`, http.StatusBadRequest)
		return
	}

	collection := database.DB.Collection("users")
	ctx := context.Background()

	var user models.User
	err = collection.FindOne(ctx, bson.M{"_id": userID}).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			http.Error(w, `{"error": "User not found"}`, http.StatusNotFound)
			return
		}
		http.Error(w, `{"error": "Failed to fetch user"}`, http.StatusInternalServerError)
		return
	}

	// Decrypt email
	decryptedEmail, err := utils.Decrypt(user.Email, config.Load().EncryptionKey)
	if err != nil {
		http.Error(w, `{"error": "Failed to decrypt user data"}`, http.StatusInternalServerError)
		return
	}

	response := UserResponse{
		ID:        user.ID.Hex(),
		Email:     decryptedEmail,
		Role:      user.Role,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}

	json.NewEncoder(w).Encode(response)
}

// @Summary Update user profile
// @Description Update current user's profile information
// @Tags user
// @Accept json
// @Produce json
// @Param request body UpdateProfileRequest true "Profile update request"
// @Security BearerAuth
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 409 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /user/profile [put]
func UpdateUserProfile(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Get user claims from context
	claims := r.Context().Value("claims").(jwt.MapClaims)
	userIDStr := claims["userID"].(string)

	userID, err := primitive.ObjectIDFromHex(userIDStr)
	if err != nil {
		http.Error(w, `{"error": "Invalid user ID"}`, http.StatusBadRequest)
		return
	}

	var req UpdateProfileRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error": "Invalid request body"}`, http.StatusBadRequest)
		return
	}

	collection := database.DB.Collection("users")
	ctx := context.Background()

	update := bson.M{
		"$set": bson.M{
			"updated_at": time.Now(),
		},
	}

	// Update email if provided
	if req.Email != "" {
		// Check if email is already taken by another user
		emailHash := utils.HashEmail(req.Email)
		cfg := config.Load()
		encryptedEmail, err := utils.Encrypt(req.Email, cfg.EncryptionKey)
		if err != nil {
			http.Error(w, `{"error": "Failed to encrypt email"}`, http.StatusInternalServerError)
			return
		}

		count, err := collection.CountDocuments(ctx, bson.M{"email_hash": emailHash, "_id": bson.M{"$ne": userID}})
		if err != nil {
			http.Error(w, `{"error": "Failed to check email availability"}`, http.StatusInternalServerError)
			return
		}

		if count > 0 {
			http.Error(w, `{"error": "Email already in use"}`, http.StatusConflict)
			return
		}

		update["$set"].(bson.M)["email"] = encryptedEmail
		update["$set"].(bson.M)["email_hash"] = emailHash
	}

	// Update password if provided
	if req.Password != "" {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {
			http.Error(w, `{"error": "Failed to hash password"}`, http.StatusInternalServerError)
			return
		}
		update["$set"].(bson.M)["password"] = string(hashedPassword)
	}

	result, err := collection.UpdateOne(ctx, bson.M{"_id": userID}, update)
	if err != nil {
		http.Error(w, `{"error": "Failed to update profile"}`, http.StatusInternalServerError)
		return
	}

	if result.MatchedCount == 0 {
		http.Error(w, `{"error": "User not found"}`, http.StatusNotFound)
		return
	}

	json.NewEncoder(w).Encode(SuccessResponse{Message: "Profile updated successfully"})
}

// UpdateProfileRequest represents the request for updating user profile
type UpdateProfileRequest struct {
	Email    string `json:"email,omitempty"`
	Password string `json:"password,omitempty"`
}

// SuccessResponse represents a success response
type SuccessResponse struct {
	Message string `json:"message"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error string `json:"error"`
}
