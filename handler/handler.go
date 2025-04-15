package handler

import (
	"fmt"
	"net/http"
	"encoding/json"
	"github.com/gorilla/mux"
	"ipfs-identity/logger"
	"ipfs-identity/util"
)

var logg = logger.Init(true)

// Custom logging function
func customLog(level, format string, args ...any) {
	logMessage := fmt.Sprintf(format, args...)
	switch level {
	case "INFO":
		logg.Info("%s", logMessage)
	case "WARNING":
		logg.Warning("%s", logMessage)
	case "ERROR":
		logg.Error("%s", logMessage)
	default:
		logg.Info("%s", logMessage)
	}
}

// Global identity manager instance.
var im = util.NewIdentityManager()

// HTTP request types.
type userRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// addUserHandler handles POST /users to add a new user.
func AddUserHandler(w http.ResponseWriter, r *http.Request) {
	var req userRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		customLog("ERROR", "Error decoding add user request: %v", err)
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	id, err := im.AddUser(req.Username, req.Password)
	if err != nil {
		customLog("ERROR", "Error adding user: %v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	customLog("INFO", "User added with ID: %s", id)

	response := map[string]string{"id": id}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// loginHandler handles POST /login to authenticate a user.
func LoginHandler(w http.ResponseWriter, r *http.Request) {
	var req userRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		customLog("ERROR", "Error decoding login request: %v", err)
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	id, err := im.Login(req.Username, req.Password)
	if err != nil {
		customLog("WARNING", "Failed login attempt for username: %s", req.Username)
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}
	customLog("INFO", "User %s logged in successfully", req.Username)

	response := map[string]string{"id": id}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// updateUserHandler handles PUT /users/{id} to update user information.
func UpdateUserHandler(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id, ok := params["id"]
	if !ok {
		http.Error(w, "User id is missing", http.StatusBadRequest)
		return
	}

	var req userRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		customLog("ERROR", "Error decoding update request: %v", err)
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	err := im.EditUser(id, req.Username, req.Password)
	if err != nil {
		customLog("ERROR", "Error updating user: %v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	customLog("INFO", "User %s updated successfully", id)
	w.WriteHeader(http.StatusNoContent)
}

// deleteUserHandler handles DELETE /users/{id} to delete a user.
func DeleteUserHandler(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id, ok := params["id"]
	if !ok {
		http.Error(w, "User id is missing", http.StatusBadRequest)
		return
	}

	err := im.DeleteUser(id)
	if err != nil {
		customLog("ERROR", "Error deleting user: %v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	customLog("INFO", "User %s deleted successfully", id)
	w.WriteHeader(http.StatusNoContent)
}
