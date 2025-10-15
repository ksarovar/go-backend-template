package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
	"golang-backend/microservices/shared/database"
	"golang-backend/microservices/shared/models"
	"golang-backend/microservices/shared/utils"
)

// UpdateRoleRequest represents the request payload for updating user role
type UpdateRoleRequest struct {
	Role string `json:"role" example:"admin"`
}

// ListUsers retrieves all users (admin only)
// @Summary List all users
// @Description Get a list of all users in the system (admin only)
// @Tags admin
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {array} models.UserResponse
// @Failure 401 {string} string "Unauthorized"
// @Failure 403 {string} string "Admin access required"
// @Failure 500 {string} string "Internal server error"
// @Router /users [get]
func ListUsers(w http.ResponseWriter, r *http.Request) {
	collection := database.GetCollection("users")
	ctx := context.Background()

	// Find all users
	cursor, err := collection.Find(ctx, bson.M{}, options.Find())
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	defer cursor.Close(ctx)

	var users []models.User
	if err = cursor.All(ctx, &users); err != nil {
		http.Error(w, "Failed to decode users", http.StatusInternalServerError)
		return
	}

	// Convert to response format with decrypted emails
	var userResponses []models.UserResponse
	for _, user := range users {
		decryptedEmail, err := utils.Decrypt(user.Email, r.Context().Value("encryptionKey").(string))
		if err != nil {
			http.Error(w, "Failed to decrypt data", http.StatusInternalServerError)
			return
		}

		userResponse := models.UserResponse{
			ID:        user.ID.Hex(),
			Email:     decryptedEmail,
			Role:      user.Role,
			CreatedAt: user.CreatedAt,
			UpdatedAt: user.UpdatedAt,
		}
		userResponses = append(userResponses, userResponse)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(userResponses)
}

// DeleteUser deletes a user by ID (admin only)
// @Summary Delete user
// @Description Delete a user by their ID (admin only)
// @Tags admin
// @Accept json
// @Produce json
// @Param id path string true "User ID"
// @Security BearerAuth
// @Success 200 {object} map[string]string
// @Failure 400 {string} string "Invalid user ID"
// @Failure 401 {string} string "Unauthorized"
// @Failure 403 {string} string "Admin access required"
// @Failure 404 {string} string "User not found"
// @Failure 500 {string} string "Internal server error"
// @Router /users/{id} [delete]
func DeleteUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userIDStr := vars["id"]

	userID, err := primitive.ObjectIDFromHex(userIDStr)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	collection := database.GetCollection("users")
	ctx := context.Background()

	result, err := collection.DeleteOne(ctx, bson.M{"_id": userID})
	if err != nil {
		http.Error(w, "Failed to delete user", http.StatusInternalServerError)
		return
	}

	if result.DeletedCount == 0 {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "User deleted successfully"})
}

// UpdateUserRole updates a user's role (admin only)
// @Summary Update user role
// @Description Update a user's role by their ID (admin only)
// @Tags admin
// @Accept json
// @Produce json
// @Param id path string true "User ID"
// @Param request body UpdateRoleRequest true "Role update data"
// @Security BearerAuth
// @Success 200 {object} map[string]string
// @Failure 400 {string} string "Invalid request payload or user ID"
// @Failure 401 {string} string "Unauthorized"
// @Failure 403 {string} string "Admin access required"
// @Failure 404 {string} string "User not found"
// @Failure 500 {string} string "Internal server error"
// @Router /users/{id}/role [put]
func UpdateUserRole(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userIDStr := vars["id"]

	userID, err := primitive.ObjectIDFromHex(userIDStr)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	var req UpdateRoleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	// Validate role
	if req.Role != "user" && req.Role != "admin" {
		http.Error(w, "Invalid role. Must be 'user' or 'admin'", http.StatusBadRequest)
		return
	}

	collection := database.GetCollection("users")
	ctx := context.Background()

	update := bson.M{
		"$set": bson.M{
			"role":       req.Role,
			"updated_at": time.Now(),
		},
	}

	result, err := collection.UpdateOne(ctx, bson.M{"_id": userID}, update)
	if err != nil {
		http.Error(w, "Failed to update user role", http.StatusInternalServerError)
		return
	}

	if result.MatchedCount == 0 {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "User role updated successfully"})
}
