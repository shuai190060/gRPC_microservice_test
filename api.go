package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	jwt "github.com/golang-jwt/jwt/v4"
	"github.com/gorilla/mux"
)

type APIServer struct {
	listenAddr string
	store      Storage
}

func NewApiServer(listenAddr string, store Storage) *APIServer {
	return &APIServer{
		listenAddr: listenAddr,
		store:      store,
	}
}

func (s *APIServer) Run() {
	router := mux.NewRouter()
	router.HandleFunc("/account", makeHttpHandleFunc(s.handleAccount))

	// take the id as variable for http request
	router.HandleFunc("/account/{id}", withJWTAuth(makeHttpHandleFunc(s.handleGetAccountByID), s.store))

	// transfer
	router.HandleFunc("/transfer", makeHttpHandleFunc(s.handleTransfer))

	log.Println("Json api server running on port:", s.listenAddr)
	http.ListenAndServe(s.listenAddr, router)

}

func (s *APIServer) handleAccount(w http.ResponseWriter, r *http.Request) error {
	switch r.Method {
	case "GET":
		return s.handleGetAccount(w, r)
	case "POST":
		return s.handleCreateAccount(w, r)
	// case "DELETE":
	// 	return s.handleDeleteAccount(w, r)
	default:
		return fmt.Errorf("method not allowed %s", r.Method)
	}
}

// get all accounts
func (s *APIServer) handleGetAccount(w http.ResponseWriter, r *http.Request) error {
	accounts, err := s.store.GetAccounts()
	if err != nil {
		return err
	}
	return WriteJSON(w, http.StatusOK, accounts)
}

// get account by ID
func (s *APIServer) handleGetAccountByID(w http.ResponseWriter, r *http.Request) error {
	// idstring := mux.Vars(r)["id"]
	// // fmt.Println(mux.Vars(r))
	// id, err := strconv.Atoi(idstring)
	// if err != nil {
	// 	return fmt.Errorf("invalid id given %s", idstring)

	// }
	if r.Method == "GET" {
		id, err := getID(r)
		if err != nil {
			return err
		}
		account, err := s.store.GetAccountByID(id)
		if err != nil {
			return err
		}

		return WriteJSON(w, http.StatusOK, account)
	}
	if r.Method == "DELETE" {
		return s.handleDeleteAccount(w, r)
	}
	return fmt.Errorf("method not allowed %s", r.Method)
}

func (s *APIServer) handleCreateAccount(w http.ResponseWriter, r *http.Request) error {
	createAccountReq := new(CreateAccountRequest)
	if err := json.NewDecoder(r.Body).Decode(createAccountReq); err != nil {
		return err
	}

	account := NewAccount(createAccountReq.FirstName, createAccountReq.LastName)
	if err := s.store.CreateAccount(account); err != nil {
		return err
	}

	tokenString, err := createJWT(account)
	if err != nil {
		return err
	}
	fmt.Println("jwt-token is", tokenString)

	return WriteJSON(w, http.StatusOK, account)
}

func (s *APIServer) handleDeleteAccount(w http.ResponseWriter, r *http.Request) error {

	id, err := getID(r)
	if err != nil {
		return err
	}

	if err := s.store.DeleteAccount(id); err != nil {
		return err
	}

	return WriteJSON(w, http.StatusOK, map[string]int{"deleted": id})
}

func (s *APIServer) handleTransfer(w http.ResponseWriter, r *http.Request) error {
	transferReq := new(TransferRequest)
	if err := json.NewDecoder(r.Body).Decode(transferReq); err != nil {
		return err
	}
	defer r.Body.Close()

	return WriteJSON(w, http.StatusOK, transferReq)
}

// helper functions

func WriteJSON(w http.ResponseWriter, status int, v any) error {

	w.Header().Add("Content-Type", "application/json") // need to add instead of set
	w.WriteHeader(status)
	return json.NewEncoder(w).Encode(v)
}

// decrorator for httphandlerFunc
type apiFunc func(http.ResponseWriter, *http.Request) error

type ApiError struct {
	Error string `json:"error"`
}

func makeHttpHandleFunc(f apiFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := f(w, r); err != nil {
			WriteJSON(w, http.StatusBadRequest, ApiError{Error: err.Error()})
		}
	}

}

func getID(r *http.Request) (int, error) {
	idstring := mux.Vars(r)["id"]
	// fmt.Println(mux.Vars(r))
	id, err := strconv.Atoi(idstring)
	if err != nil {
		return id, fmt.Errorf("invalid id given %s", idstring)

	}
	return id, nil
}

// middleware to add jwt
func withJWTAuth(handlerFunc http.HandlerFunc, s Storage) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("calling JWT auth middleware")
		tokenString := r.Header.Get("x-jwt-token")
		// if len(jwt.ErrTokenNotValidYet.Error())
		token, err := validateJWT(tokenString)
		if err != nil {
			WriteJSON(w, http.StatusForbidden, ApiError{Error: "invalid token-0"})
			return
		}

		if !token.Valid {
			WriteJSON(w, http.StatusForbidden, ApiError{Error: "invalid token-1"})
			return
		}
		// -------------------------------------
		// when the userID match with the claim number, then it can be returning the handlerFunc
		// so it restrict  user-A can check only user-A's info, not others
		userID, err := getID(r)
		if err != nil {
			WriteJSON(w, http.StatusForbidden, ApiError{Error: "invalid token-2"})
			return
		}
		account, err := s.GetAccountByID(userID)
		if err != nil {
			WriteJSON(w, http.StatusForbidden, ApiError{Error: "invalid token-3"})
			return
		}
		claims := token.Claims.(jwt.MapClaims)
		// panic(reflect.TypeOf(claims["accountNumber"]))
		// fmt.Println(claims, int64(claims["accountNumber"].(float64)), account.Number)
		if account.Number != int64(claims["accountNumber"].(float64)) {
			WriteJSON(w, http.StatusForbidden, ApiError{Error: "invalid token-4"})
			return
		}
		// -------------------------------------
		handlerFunc(w, r)
	}
}

func validateJWT(tokenString string) (*jwt.Token, error) {
	secret := os.Getenv("JWT_SECRET")
	return jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Don't forget to validate the alg is what you expect:
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}

		// hmacSampleSecret is a []byte containing your secret, e.g. []byte("my_secret_key")
		return []byte(secret), nil
	})
}

func createJWT(account *Account) (string, error) {
	// Create the Claims
	claims := &jwt.MapClaims{
		"expiresAt":     15000,
		"accountNumber": account.Number,
	}

	secret := os.Getenv("JWT_SECRET")
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret)) // pass the secret in as signing key

}
