package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"ipfs-identity/handler" 
	"ipfs-identity/logger"
)



func main() {
	r := mux.NewRouter()

	config := logger.NewConfigFromEnv()

	// Initialize logger
	log, err := logger.NewLogger(config)
	if err != nil {
		fmt.Printf("Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}
	defer log.Sync()

	// Define API endpoints.
	r.HandleFunc("/addusers", handler.AddUserHandler).Methods("POST")
	r.HandleFunc("/users/{id}", handler.UpdateUserHandler).Methods("PUT")
	r.HandleFunc("/users/{id}", handler.DeleteUserHandler).Methods("DELETE")
	r.HandleFunc("/login", handler.LoginHandler).Methods("POST")

	// Optional: You can add a root handler.
	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		msg := "Welcome to the Identity API"
		log.Info(fmt.Sprintf("Root accessed: %s", r.URL.Path))
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"message": msg})
	}).Methods("GET")

	addr := os.Getenv("SERVER_ADDR")
	if addr == "" {
		log.Error("SERVER_ADDR environment variable not set")
		os.Exit(1)
	}
	if !strings.Contains(addr, ":") {
		addr += ":6501"
	}

	ipfsNode := os.Getenv("IPFS_NODE")
	if ipfsNode == "" {
		log.Error("IPFS_NODE environment variable not set")
		os.Exit(1)
	}
	log.Info(fmt.Sprintf("IPFS node is running at %s", ipfsNode))

	log.Info(fmt.Sprintf("Starting server on %s", addr))
	srv := &http.Server{
		Handler:      r,
		Addr:         addr,
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}
	if err := srv.ListenAndServe(); err != nil {
		log.Error(fmt.Sprintf("Server failed to start: %v", err))
		log.Fatal(err.Error())
	}
}
