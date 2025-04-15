package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"ipfs-identity/handler" 
)

func customLog(level, format string, args ...interface{}) {
	logMessage := fmt.Sprintf(format, args...)
	switch level {
	case "INFO":
		fmt.Printf("[INFO] %s\n", logMessage)
	case "WARNING":
		fmt.Printf("[WARNING] %s\n", logMessage)
	case "ERROR":
		fmt.Printf("[ERROR] %s\n", logMessage)
	default:
		fmt.Printf("[INFO] %s\n", logMessage)
	}
}

func main() {
	r := mux.NewRouter()

	// Define API endpoints.
	r.HandleFunc("/users", handler.AddUserHandler).Methods("POST")
	r.HandleFunc("/users/{id}", handler.UpdateUserHandler).Methods("PUT")
	r.HandleFunc("/users/{id}", handler.DeleteUserHandler).Methods("DELETE")
	r.HandleFunc("/login", handler.LoginHandler).Methods("POST")

	// Optional: You can add a root handler.
	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		msg := "Welcome to the Identity API"
		customLog("INFO", "Root accessed: %s", r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"message": msg})
	}).Methods("GET")

	addr := os.Getenv("SERVER_ADDR")
	if addr == "" {
		customLog("ERROR", "SERVER_ADDR environment variable not set")
		os.Exit(1)
	}
	if !strings.Contains(addr, ":") {
		addr += ":6501"
	}
	customLog("INFO", "Starting server on %s", addr)
	srv := &http.Server{
		Handler:      r,
		Addr:         addr,
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}
	if err := srv.ListenAndServe(); err != nil {
		customLog("ERROR", "Server failed to start: %v", err)
		log.Fatal(err)
	}
}
