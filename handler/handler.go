package handler

import (
	"log"
	"net/http"
	"encoding/json"
	"github.com/gorilla/mux"
	"ipfs-identity/logger"
	"ipfs-identity/util"
)


// Global identity manager instance.
var im = util.NewIdentityManager()

// HTTP request types.
type userRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// addUserHandler handles POST /users to add a new user.
func AddUserHandler(w http.ResponseWriter, r *http.Request) {

	config := logger.NewConfigFromEnv()

	logInstance, err := logger.NewLogger(config)
	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}

	var req userRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logInstance.Error("Error decoding add user request: %v", err)
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	id, err := im.AddUser(req.Username, req.Password)
	if err != nil {
		logInstance.Error("Error adding user: %v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	logInstance.Info("User added with ID: %s", id)

	response := map[string]string{
		"message": "User added successfully",
		"id":      id,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// loginHandler handles POST /login to authenticate a user.
func LoginHandler(w http.ResponseWriter, r *http.Request) {

	config := logger.NewConfigFromEnv()

	logInstance, err := logger.NewLogger(config)
    if err != nil {
        log.Fatalf("Failed to initialize logger: %v", err)
    }

	var req userRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logInstance.Error("Error decoding login request: %v", err)
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	id, err := im.Login(req.Username, req.Password)
	if err != nil {
		logInstance.Warn("Failed login attempt for username: %s", req.Username)
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}
	logInstance.Info("User %s logged in successfully", req.Username)

	response := map[string]string{"id": id}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
// updateUserHandler handles PUT /users/{id} to update user information.
func UpdateUserHandler(w http.ResponseWriter, r *http.Request) {

	config := logger.NewConfigFromEnv()

	logInstance, err := logger.NewLogger(config)
	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}

	params := mux.Vars(r)
	id, ok := params["id"]
	if !ok {
		http.Error(w, "User id is missing", http.StatusBadRequest)
		return
	}

	var req userRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logInstance.Error("Error decoding update request: %v", err)
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	err = im.EditUser(id, req.Username, req.Password)
	if err != nil {
		logInstance.Error("Error updating user: %v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	logInstance.Info("User %s updated successfully", id)

	response := map[string]string{
		"message": "User updated successfully",
		"id":      id,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// deleteUserHandler handles DELETE /users/{id} to delete a user.
func DeleteUserHandler(w http.ResponseWriter, r *http.Request) {
	config := logger.NewConfigFromEnv()

	logInstance, err := logger.NewLogger(config)
	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	params := mux.Vars(r)
	id, ok := params["id"]
	if !ok {
		http.Error(w, "User id is missing", http.StatusBadRequest)
		return
	}

	err = im.DeleteUser(id)
	if err != nil {
		logInstance.Error("Error deleting user: %v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	logInstance.Info("User %s deleted successfully", id)

	response := map[string]string{
		"message": "User deleted successfully",
		"id":      id,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
