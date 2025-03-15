package main

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	ipfsapi "github.com/ipfs/go-ipfs-api"
	"github.com/joho/godotenv"
	"golang.org/x/crypto/bcrypt"

	"ipfs-identity/logger" // adjust the import path based on your module name
)

var logg = logger.Init(true)

// User represents the user identity structure.
type User struct {
	ID        string    `json:"id"`
	Username  string    `json:"username"`
	Password  string    `json:"password"` // Hashed password
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// IdentityManager handles identity operations.
type IdentityManager struct {
	ipfs *ipfsapi.Shell
	cid  string // Content ID of the user database
}

// NewIdentityManager initializes the IdentityManager.
func NewIdentityManager() *IdentityManager {
	// Load .env file
	err := godotenv.Load()
	if err != nil {
		logg.Error("Error loading .env file: %v", err)
		os.Exit(1)
	}

	ipfsNode := os.Getenv("IPFS_NODE")
	if ipfsNode == "" {
		logg.Error("IPFS_NODE environment variable not set")
		os.Exit(1)
	}

	shell := ipfsapi.NewShell(ipfsNode)
	return &IdentityManager{ipfs: shell}
}

// loadUsers retrieves users from IPFS.
func (im *IdentityManager) loadUsers() (map[string]User, error) {
	if im.cid == "" {
		return make(map[string]User), nil
	}

	reader, err := im.ipfs.Cat(im.cid)
	if err != nil {
		return nil, err
	}
	defer reader.Close()

	data, err := ioutil.ReadAll(reader)
	if err != nil {
		return nil, err
	}

	var users map[string]User
	err = json.Unmarshal(data, &users)
	return users, err
}

// saveUsers saves users to IPFS.
func (im *IdentityManager) saveUsers(users map[string]User) error {
	data, err := json.Marshal(users)
	if err != nil {
		return err
	}

	cid, err := im.ipfs.Add(strings.NewReader(string(data)))
	if err != nil {
		return err
	}
	im.cid = cid
	return nil
}

// AddUser creates a new user.
func (im *IdentityManager) AddUser(username, password string) (string, error) {
	users, err := im.loadUsers()
	if err != nil {
		return "", err
	}

	for _, user := range users {
		if user.Username == username {
			return "", errors.New("username already exists")
		}
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}

	id := uuid.New().String()
	user := User{
		ID:        id,
		Username:  username,
		Password:  string(hashedPassword),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	users[id] = user
	err = im.saveUsers(users)
	return id, err
}

// EditUser updates an existing user.
func (im *IdentityManager) EditUser(id, newUsername, newPassword string) error {
	users, err := im.loadUsers()
	if err != nil {
		return err
	}

	user, exists := users[id]
	if !exists {
		return errors.New("user not found")
	}

	if newUsername != "" {
		user.Username = newUsername
	}
	if newPassword != "" {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
		if err != nil {
			return err
		}
		user.Password = string(hashedPassword)
	}
	user.UpdatedAt = time.Now()

	users[id] = user
	return im.saveUsers(users)
}

// DeleteUser removes a user.
func (im *IdentityManager) DeleteUser(id string) error {
	users, err := im.loadUsers()
	if err != nil {
		return err
	}

	if _, exists := users[id]; !exists {
		return errors.New("user not found")
	}

	delete(users, id)
	return im.saveUsers(users)
}

// Login authenticates a user.
func (im *IdentityManager) Login(username, password string) (string, error) {
	users, err := im.loadUsers()
	if err != nil {
		return "", err
	}

	for id, user := range users {
		if user.Username == username {
			err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
			if err == nil {
				return id, nil
			}
			return "", errors.New("invalid password")
		}
	}
	return "", errors.New("user not found")
}

// Global identity manager instance.
var im = NewIdentityManager()

// HTTP request types.
type userRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// addUserHandler handles POST /users to add a new user.
func addUserHandler(w http.ResponseWriter, r *http.Request) {
	var req userRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logg.Error("Error decoding add user request: %v", err)
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	id, err := im.AddUser(req.Username, req.Password)
	if err != nil {
		logg.Error("Error adding user: %v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	logg.Info("User added with ID: %s", id)

	response := map[string]string{"id": id}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// loginHandler handles POST /login to authenticate a user.
func loginHandler(w http.ResponseWriter, r *http.Request) {
	var req userRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logg.Error("Error decoding login request: %v", err)
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	id, err := im.Login(req.Username, req.Password)
	if err != nil {
		logg.Warning("Failed login attempt for username: %s", req.Username)
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}
	logg.Info("User %s logged in successfully", req.Username)

	response := map[string]string{"id": id}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// updateUserHandler handles PUT /users/{id} to update user information.
func updateUserHandler(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id, ok := params["id"]
	if !ok {
		http.Error(w, "User id is missing", http.StatusBadRequest)
		return
	}

	var req userRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logg.Error("Error decoding update request: %v", err)
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	err := im.EditUser(id, req.Username, req.Password)
	if err != nil {
		logg.Error("Error updating user: %v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	logg.Info("User %s updated successfully", id)
	w.WriteHeader(http.StatusNoContent)
}

// deleteUserHandler handles DELETE /users/{id} to delete a user.
func deleteUserHandler(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id, ok := params["id"]
	if !ok {
		http.Error(w, "User id is missing", http.StatusBadRequest)
		return
	}

	err := im.DeleteUser(id)
	if err != nil {
		logg.Error("Error deleting user: %v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	logg.Info("User %s deleted successfully", id)
	w.WriteHeader(http.StatusNoContent)
}

func main() {
	r := mux.NewRouter()

	// Define API endpoints.
	r.HandleFunc("/users", addUserHandler).Methods("POST")
	r.HandleFunc("/users/{id}", updateUserHandler).Methods("PUT")
	r.HandleFunc("/users/{id}", deleteUserHandler).Methods("DELETE")
	r.HandleFunc("/login", loginHandler).Methods("POST")

	// Optional: You can add a root handler.
	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		msg := "Welcome to the Identity API"
		logg.Info("Root accessed: %s", r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"message": msg})
	}).Methods("GET")

	addr := os.Getenv("SERVER_ADDR")
	if addr == "" {
		logg.Error("SERVER_ADDR environment variable not set")
		os.Exit(1)
	}
	if !strings.Contains(addr, ":") {
		addr += ":6501"
	}
	logg.Info("Starting server on %s", addr)
	srv := &http.Server{
		Handler:      r,
		Addr:         addr,
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}
	if err := srv.ListenAndServe(); err != nil {
		logg.Error("Server failed to start: %v", err)
		log.Fatal(err)
	}
}
