package main

import (
	"encoding/json"
	"fmt"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

type User struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Role     string `json:"role"`
	Password string `json:"password"` // Store securely (hashed in a real app)
	Status   string `json:"status"`   // "online" or "offline"
}

type SmartContract struct {
	contractapi.Contract
}

// CreateUser - Registers a new user
func (s *SmartContract) CreateUser(ctx contractapi.TransactionContextInterface, id, name, username, email, role, password string) error {
	existingUser, _ := ctx.GetStub().GetState(id)
	if existingUser != nil {
		return fmt.Errorf("user already exists")
	}

	user := User{
		ID:       id,
		Name:     name,
		Username: username,
		Email:    email,
		Role:     role,
		Password: password, // In production, store hashed password
		Status:   "offline",
	}

	userBytes, _ := json.Marshal(user)
	return ctx.GetStub().PutState(id, userBytes)
}

// ReadUser - Fetches user details
func (s *SmartContract) ReadUser(ctx contractapi.TransactionContextInterface, id string) (*User, error) {
	userBytes, err := ctx.GetStub().GetState(id)
	if err != nil || userBytes == nil {
		return nil, fmt.Errorf("user not found")
	}

	var user User
	_ = json.Unmarshal(userBytes, &user)
	return &user, nil
}

// UpdateUser - Updates user details (excluding password)
func (s *SmartContract) UpdateUser(ctx contractapi.TransactionContextInterface, id, name, username, email, role string) error {
	user, err := s.ReadUser(ctx, id)
	if err != nil {
		return err
	}

	user.Name = name
	user.Username = username
	user.Email = email
	user.Role = role

	userBytes, _ := json.Marshal(user)
	return ctx.GetStub().PutState(id, userBytes)
}

// DeleteUser - Removes a user
func (s *SmartContract) DeleteUser(ctx contractapi.TransactionContextInterface, id string) error {
	_, err := s.ReadUser(ctx, id)
	if err != nil {
		return err
	}
	return ctx.GetStub().DelState(id)
}

// Login - Authenticates user & updates status to "online"
func (s *SmartContract) Login(ctx contractapi.TransactionContextInterface, username, password string) (string, error) {
	resultsIterator, err := ctx.GetStub().GetStateByRange("", "")
	if err != nil {
		return "", err
	}
	defer resultsIterator.Close()

	for resultsIterator.HasNext() {
		queryResponse, _ := resultsIterator.Next()
		var user User
		_ = json.Unmarshal(queryResponse.Value, &user)

		if user.Username == username && user.Password == password {
			if user.Status == "online" {
				return "", fmt.Errorf("user is already online")
			}
			user.Status = "online"
			userBytes, _ := json.Marshal(user)
			_ = ctx.GetStub().PutState(user.ID, userBytes)
			return "Login successful", nil
		}
	}
	return "", fmt.Errorf("invalid username or password")
}

// Logout - Updates user status to "offline"
func (s *SmartContract) Logout(ctx contractapi.TransactionContextInterface, id string) error {
	user, err := s.ReadUser(ctx, id)
	if err != nil {
		return err
	}

	user.Status = "offline"
	userBytes, _ := json.Marshal(user)
	return ctx.GetStub().PutState(id, userBytes)
}

func main() {
	chaincode, err := contractapi.NewChaincode(new(SmartContract))
	if err != nil {
		fmt.Println("Error creating chaincode:", err)
	}

	if err := chaincode.Start(); err != nil {
		fmt.Println("Error starting chaincode:", err)
	}
}
