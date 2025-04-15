package util

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/joho/godotenv"
	ipfsapi "github.com/ipfs/go-ipfs-api"
	"golang.org/x/crypto/bcrypt"

	"ipfs-identity/logger"
)

// User represents the user identity structure.
type User struct {
	ID        string    `json:"id"`
	Username  string    `json:"username"`
	Password  string    `json:"password"` // Hashed password
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// IdentityManager handles identity operations.
// It includes a mutex for protecting concurrent access to the user data and CID.
type IdentityManager struct {
	ipfs *ipfsapi.Shell
	cid  string // Content ID of the user database
	mu   sync.RWMutex
	log  logger.Logger
}

// NewIdentityManager initializes the IdentityManager.
func NewIdentityManager() *IdentityManager {
	// Load configuration for logger.
	config := logger.NewConfigFromEnv() // Adjust as necessary
	log, err := logger.NewLogger(config)
	if err != nil {
		fmt.Printf("Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}

	// Load .env file
	if err := godotenv.Load(); err != nil {
		log.Error("Error loading .env file: %v", err)
		os.Exit(1)
	}

	ipfsNode := os.Getenv("IPFS_NODE")
	if ipfsNode == "" {
		log.Error("IPFS_NODE environment variable not set")
		os.Exit(1)
	}

	shell := ipfsapi.NewShell(ipfsNode)
	return &IdentityManager{
		ipfs: shell,
		log:  log,
	}
}

// loadUsers retrieves users from IPFS.
func (im *IdentityManager) loadUsers() (map[string]User, error) {
	im.mu.RLock()
	defer im.mu.RUnlock()

	// If no CID is set, return an empty map.
	if im.cid == "" {
		return make(map[string]User), nil
	}

	reader, err := im.ipfs.Cat(im.cid)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch data from IPFS: %w", err)
	}
	defer reader.Close()

	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("failed reading data: %w", err)
	}

	var users map[string]User
	if err := json.Unmarshal(data, &users); err != nil {
		return nil, fmt.Errorf("failed to unmarshal user data: %w", err)
	}
	return users, nil
}

// saveUsers saves users to IPFS.
func (im *IdentityManager) saveUsers(users map[string]User) error {
	// First marshal the data to JSON.
	data, err := json.Marshal(users)
	if err != nil {
		return fmt.Errorf("failed to marshal users: %w", err)
	}

	// Save the JSON data to IPFS.
	cid, err := im.ipfs.Add(strings.NewReader(string(data)))
	if err != nil {
		return fmt.Errorf("failed to add data to IPFS: %w", err)
	}

	// Lock for writing the new CID.
	im.mu.Lock()
	im.cid = cid
	im.mu.Unlock()

	return nil
}

// AddUser creates a new user.
func (im *IdentityManager) AddUser(username, password string) (string, error) {
	// Load the current users.
	users, err := im.loadUsers()
	if err != nil {
		return "", err
	}

	// Check if the username already exists.
	for _, user := range users {
		if user.Username == username {
			return "", errors.New("username already exists")
		}
	}

	// Generate a hashed password.
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("failed to hash password: %w", err)
	}

	// Create a new user.
	id := uuid.New().String()
	newUser := User{
		ID:        id,
		Username:  username,
		Password:  string(hashedPassword),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Update the users map.
	users[id] = newUser
	if err := im.saveUsers(users); err != nil {
		return "", err
	}

	im.log.Info("User added successfully with ID: %s", id)
	return id, nil
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
			return fmt.Errorf("failed to hash new password: %w", err)
		}
		user.Password = string(hashedPassword)
	}

	user.UpdatedAt = time.Now()
	users[id] = user

	if err := im.saveUsers(users); err != nil {
		return err
	}

	im.log.Info("User with ID %s updated successfully", id)
	return nil
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
	if err := im.saveUsers(users); err != nil {
		return err
	}

	im.log.Info("User with ID %s deleted successfully", id)
	return nil
}

// Login authenticates a user.
func (im *IdentityManager) Login(username, password string) (string, error) {
	users, err := im.loadUsers()
	if err != nil {
		return "", err
	}

	for id, user := range users {
		if user.Username == username {
			if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err == nil {
				im.log.Info("User %s authenticated successfully", username)
				return id, nil
			}
			return "", errors.New("invalid password")
		}
	}
	return "", errors.New("user not found")
}

