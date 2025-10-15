package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"golang-backend/microservices/shared/database"
	"golang-backend/microservices/shared/models"
	"golang-backend/microservices/shared/utils"
)

// UpdateProfileRequest represents the request payload for updating user profile
type UpdateProfileRequest struct {
	Email string `json:"email" example:"newemail@example.com"`
}

// GetUserProfile retrieves the current user's profile
// @Summary Get user profile
// @Description Get the current authenticated user's profile information
// @Tags user
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} models.UserResponse
// @Failure 401 {string} string "Unauthorized"
// @Failure 404 {string} string "User not found"
// @Failure 500 {string} string "Internal server error"
// @Router /profile [get]
func GetUserProfile(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context (set by middleware)
	userIDStr := r.Context().Value("userID").(string)
	userID, err := primitive.ObjectIDFromHex(userIDStr)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	collection := database.GetCollection("users")
	ctx := context.Background()

	var user models.User
	err = collection.FindOne(ctx, bson.M{"_id": userID}).Decode(&user)
	if err != nil {
		if err.Error() == "mongo: no documents in result" {
			http.Error(w, "User not found", http.StatusNotFound)
		} else {
			http.Error(w, "Database error", http.StatusInternalServerError)
		}
		return
	}

	// Decrypt email
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

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(userResponse)
}

// UpdateUserProfile updates the current user's profile
// @Summary Update user profile
// @Description Update the current authenticated user's profile information
// @Tags user
// @Accept json
// @Produce json
// @Param request body UpdateProfileRequest true "Profile update data"
// @Security BearerAuth
// @Success 200 {object} map[string]string
// @Failure 400 {string} string "Invalid request payload"
// @Failure 401 {string} string "Unauthorized"
// @Failure 404 {string} string "User not found"
// @Failure 500 {string} string "Internal server error"
// @Router /profile [put]
func UpdateUserProfile(w http.ResponseWriter, r *http.Request) {
	var req UpdateProfileRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	// Get user ID from context (set by middleware)
	userIDStr := r.Context().Value("userID").(string)
	userID, err := primitive.ObjectIDFromHex(userIDStr)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	collection := database.GetCollection("users")
	ctx := context.Background()

	// Encrypt new email
	encryptedEmail, err := utils.Encrypt(req.Email, r.Context().Value("encryptionKey").(string))
	if err != nil {
		http.Error(w, "Failed to encrypt data", http.StatusInternalServerError)
		return
	}

	// Update user
	update := bson.M{
		"$set": bson.M{
			"email":      encryptedEmail,
			"email_hash": req.Email,
			"updated_at": time.Now(),
		},
	}

	result, err := collection.UpdateOne(ctx, bson.M{"_id": userID}, update)
	if err != nil {
		http.Error(w, "Failed to update user", http.StatusInternalServerError)
		return
	}

	if result.MatchedCount == 0 {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Profile updated successfully"})
}
