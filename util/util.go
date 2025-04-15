package util

import (

	"errors"
	"time"
	"os"
	"io/ioutil"
	"encoding/json"
	"strings"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"ipfs-identity/logger"
	"github.com/joho/godotenv"
	ipfsapi "github.com/ipfs/go-ipfs-api"
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

