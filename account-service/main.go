package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
)

// Account represents a bank account
type Account struct {
	ID           int     `json:"id"`
	CustomerID   int     `json:"customer_id"`
	AccountType  string  `json:"account_type"`
	Balance      float64 `json:"balance"`
	CurrencyCode string  `json:"currency_code"`
	Status       string  `json:"status"`
	CreatedAt    string  `json:"created_at"`
	UpdatedAt    string  `json:"updated_at"`
}

var db *sql.DB

func main() {
	// Initialize database connection
	initDB()
	defer db.Close()

	// Create router
	router := mux.NewRouter()

	// Define routes
	router.HandleFunc("/health", healthCheck).Methods("GET")
	router.HandleFunc("/accounts", getAccounts).Methods("GET")
	router.HandleFunc("/accounts/{id}", getAccount).Methods("GET")
	router.HandleFunc("/accounts", createAccount).Methods("POST")
	router.HandleFunc("/accounts/{id}", updateAccount).Methods("PUT")
	router.HandleFunc("/accounts/{id}/balance", getBalance).Methods("GET")
	router.HandleFunc("/accounts/{id}/deposit", depositFunds).Methods("POST")
	router.HandleFunc("/accounts/{id}/withdraw", withdrawFunds).Methods("POST")

	// Start server
	port := getEnv("PORT", "8080")
	log.Printf("Account service starting on port %s...", port)
	log.Fatal(http.ListenAndServe(":"+port, router))
}

func initDB() {
	// Get database connection parameters from environment variables
	host := getEnv("DB_HOST", "localhost")
	port := getEnv("DB_PORT", "5432")
	user := getEnv("DB_USER", "postgres")
	password := getEnv("DB_PASSWORD", "postgres")
	dbname := getEnv("DB_NAME", "bankdb")

	// Create connection string
	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)

	// Open database connection
	var err error
	db, err = sql.Open("postgres", connStr)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Check connection
	err = db.Ping()
	if err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}

	log.Println("Successfully connected to database")

	// Create accounts table if it doesn't exist
	createTableSQL := `
	CREATE TABLE IF NOT EXISTS accounts (
		id SERIAL PRIMARY KEY,
		customer_id INTEGER NOT NULL,
		account_type VARCHAR(50) NOT NULL,
		balance DECIMAL(15,2) NOT NULL DEFAULT 0.00,
		currency_code VARCHAR(3) NOT NULL DEFAULT 'USD',
		status VARCHAR(20) NOT NULL DEFAULT 'active',
		created_at TIMESTAMP NOT NULL DEFAULT NOW(),
		updated_at TIMESTAMP NOT NULL DEFAULT NOW()
	);`

	_, err = db.Exec(createTableSQL)
	if err != nil {
		log.Fatalf("Failed to create accounts table: %v", err)
	}
}

func healthCheck(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(map[string]bool{"status": true})
}

func getAccounts(w http.ResponseWriter, r *http.Request) {
	// Get query parameters for pagination
	limit := r.URL.Query().Get("limit")
	offset := r.URL.Query().Get("offset")
	
	if limit == "" {
		limit = "100" // Default limit
	}
	
	if offset == "" {
		offset = "0" // Default offset
	}

	// Query accounts with pagination
	query := `SELECT id, customer_id, account_type, balance, currency_code, status, 
			  created_at, updated_at FROM accounts LIMIT $1 OFFSET $2`
	
	rows, err := db.Query(query, limit, offset)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	accounts := []Account{}
	for rows.Next() {
		var a Account
		err := rows.Scan(&a.ID, &a.CustomerID, &a.AccountType, &a.Balance, 
						&a.CurrencyCode, &a.Status, &a.CreatedAt, &a.UpdatedAt)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		accounts = append(accounts, a)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(accounts)
}

func getAccount(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id := params["id"]

	var account Account
	query := `SELECT id, customer_id, account_type, balance, currency_code, status, 
			  created_at, updated_at FROM accounts WHERE id = $1`
	
	err := db.QueryRow(query, id).Scan(&account.ID, &account.CustomerID, &account.AccountType, 
									  &account.Balance, &account.CurrencyCode, &account.Status, 
									  &account.CreatedAt, &account.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Account not found", http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(account)
}

func createAccount(w http.ResponseWriter, r *http.Request) {
	var account Account
	err := json.NewDecoder(r.Body).Decode(&account)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Validate required fields
	if account.CustomerID == 0 || account.AccountType == "" {
		http.Error(w, "Customer ID and account type are required", http.StatusBadRequest)
		return
	}

	// Insert new account
	query := `INSERT INTO accounts (customer_id, account_type, balance, currency_code, status) 
			  VALUES ($1, $2, $3, $4, $5) RETURNING id, created_at, updated_at`
	
	err = db.QueryRow(query, account.CustomerID, account.AccountType, account.Balance, 
					 account.CurrencyCode, account.Status).Scan(&account.ID, &account.CreatedAt, &account.UpdatedAt)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(account)
}

func updateAccount(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id := params["id"]

	var account Account
	err := json.NewDecoder(r.Body).Decode(&account)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Update account
	query := `UPDATE accounts SET account_type = $1, status = $2, updated_at = NOW() 
			  WHERE id = $3 RETURNING id, customer_id, account_type, balance, currency_code, status, created_at, updated_at`
	
	err = db.QueryRow(query, account.AccountType, account.Status, id).Scan(&account.ID, &account.CustomerID, 
																		 &account.AccountType, &account.Balance, 
																		 &account.CurrencyCode, &account.Status, 
																		 &account.CreatedAt, &account.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Account not found", http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(account)
}

func getBalance(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id := params["id"]

	var balance float64
	var currencyCode string
	query := `SELECT balance, currency_code FROM accounts WHERE id = $1`
	
	err := db.QueryRow(query, id).Scan(&balance, &currencyCode)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Account not found", http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"account_id": id,
		"balance": balance,
		"currency_code": currencyCode,
	})
}

func depositFunds(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id := params["id"]

	// Parse request body
	var requestBody struct {
		Amount float64 `json:"amount"`
	}
	
	err := json.NewDecoder(r.Body).Decode(&requestBody)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Validate amount
	if requestBody.Amount <= 0 {
		http.Error(w, "Amount must be positive", http.StatusBadRequest)
		return
	}

	// Begin transaction
	tx, err := db.Begin()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()

	// Update balance
	query := `UPDATE accounts SET balance = balance + $1, updated_at = NOW() 
			  WHERE id = $2 RETURNING balance, currency_code`
	
	var newBalance float64
	var currencyCode string
	err = tx.QueryRow(query, requestBody.Amount, id).Scan(&newBalance, &currencyCode)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Account not found", http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	// Commit transaction
	err = tx.Commit()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Return updated balance
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"account_id": id,
		"balance": newBalance,
		"currency_code": currencyCode,
		"message": fmt.Sprintf("Successfully deposited %.2f", requestBody.Amount),
	})
}

func withdrawFunds(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id := params["id"]

	// Parse request body
	var requestBody struct {
		Amount float64 `json:"amount"`
	}
	
	err := json.NewDecoder(r.Body).Decode(&requestBody)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Validate amount
	if requestBody.Amount <= 0 {
		http.Error(w, "Amount must be positive", http.StatusBadRequest)
		return
	}

	// Begin transaction
	tx, err := db.Begin()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()

	// Check if account has sufficient funds
	var currentBalance float64
	err = tx.QueryRow("SELECT balance FROM accounts WHERE id = $1", id).Scan(&currentBalance)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Account not found", http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	if currentBalance < requestBody.Amount {
		http.Error(w, "Insufficient funds", http.StatusBadRequest)
		return
	}

	// Update balance
	query := `UPDATE accounts SET balance = balance - $1, updated_at = NOW() 
			  WHERE id = $2 RETURNING balance, currency_code`
	
	var newBalance float64
	var currencyCode string
	err = tx.QueryRow(query, requestBody.Amount, id).Scan(&newBalance, &currencyCode)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Commit transaction
	err = tx.Commit()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Return updated balance
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"account_id": id,
		"balance": newBalance,
		"currency_code": currencyCode,
		"message": fmt.Sprintf("Successfully withdrew %.2f", requestBody.Amount),
	})
}

// Helper function to get environment variable with default value
func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}