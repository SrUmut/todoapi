package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/mux"
	"golang.org/x/crypto/bcrypt"
)

var secret = []byte("supersecretpassword")

type APIServer struct {
	listenAddr string
	store      Storage
}

func newAPIServer(listenAddr string, db Storage) *APIServer {
	return &APIServer{
		listenAddr: listenAddr,
		store:      db,
	}
}

func (s *APIServer) Start() error {
	router := mux.NewRouter()
	router.HandleFunc("/account", getHandlerFunc(s.handleAccount))
	router.HandleFunc("/account/{id:[0-9]+}", getHandlerFunc(s.handleAccountByID))
	router.HandleFunc("/login/{id:[0-9]+}", getHandlerFunc(s.handleLogin))

	log.Println("Starting API server on", s.listenAddr)
	if err := http.ListenAndServe(s.listenAddr, router); err != nil {
		return err
	}

	return nil
}

func (s *APIServer) handleAccount(w http.ResponseWriter, r *http.Request) error {
	switch r.Method {
	case "GET":
		return s.handleGetAccount(w, r)
	case "POST":
		return s.handleCreateAccount(w, r)
	default:
		return writeJSON(w, http.StatusMethodNotAllowed, APIError{"not allowed"})
	}
}

func (s *APIServer) handleAccountByID(w http.ResponseWriter, r *http.Request) error {
	switch r.Method {
	case "GET":
		return s.handleGetTodoByID(w, r)
	case "POST":
		return s.handleModifyTodo(w, r)
	case "DELETE":
		return s.handleDeleteAccountByID(w, r)
	default:
		return writeJSON(w, http.StatusMethodNotAllowed, APIError{"not allowed"})
	}
}

func (s *APIServer) handleLogin(w http.ResponseWriter, r *http.Request) error {
	if r.Method != "POST" {
		return writeJSON(w, http.StatusMethodNotAllowed, APIError{"not allowed"})
	}
	idStr := mux.Vars(r)["id"]
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return err
	}

	if err := s.comparePassword(r, id); err != nil {
		return err
	}

	tokenString, err := createJWT(idStr)
	if err != nil {
		return err
	}

	return writeJSON(w, http.StatusOK, map[string]string{"token": tokenString})
}

func (s *APIServer) handleGetAccount(w http.ResponseWriter, r *http.Request) error {
	accounts, err := s.store.GetAccount()
	if err != nil {
		return err
	}
	return writeJSON(w, http.StatusOK, accounts)
}

func (s *APIServer) handleCreateAccount(w http.ResponseWriter, r *http.Request) error {
	req := new(CreateAccountReq)
	if err := json.NewDecoder(r.Body).Decode(req); err != nil {
		return err
	}

	if req.FirtName == "" || req.LastName == "" || req.Password == "" {
		return fmt.Errorf("failed")
	}

	account, err := NewAccount(req.FirtName, req.LastName, req.Password)
	if err != nil {
		return err
	}

	if err := s.store.CreateAccount(account); err != nil {
		return err
	}

	return writeJSON(w, http.StatusOK, account)
}

func (s *APIServer) handleGetTodoByID(w http.ResponseWriter, r *http.Request) error {
	idStr := mux.Vars(r)["id"]
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return err
	}

	if err := s.validateJWT(r, id); err != nil {
		return err
	}

	todos, err := s.store.GetTodoByID(id)
	if err != nil {
		return err
	}
	return writeJSON(w, http.StatusOK, todos)
}

func (s *APIServer) handleModifyTodo(w http.ResponseWriter, r *http.Request) error {
	idStr := mux.Vars(r)["id"]
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return err
	}
	req := new(TodoModifyReq)
	if err := json.NewDecoder(r.Body).Decode(req); err != nil {
		return err
	}

	if err := s.store.AddTodoWithID(id, req.Add); err != nil {
		return err
	}
	if err := s.store.StatusChangeTodoWithID(id, req.StatusChange); err != nil {
		return err
	}
	if err := s.store.DeleteTodoWithID(id, req.Delete); err != nil {
		return err
	}

	writeJSON(w, http.StatusOK, "done")

	return nil
}

func (s *APIServer) handleDeleteAccountByID(w http.ResponseWriter, r *http.Request) error {
	idStr := mux.Vars(r)["id"]
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return err
	}

	if err := s.validateJWT(r, id); err != nil {
		return err
	}

	if err := s.comparePassword(r, id); err != nil {
		return err
	}

	if err := s.store.DeleteAccountByID(id); err != nil {
		return err
	}

	writeJSON(w, http.StatusOK, "done")

	return nil
}

type APIFunc func(w http.ResponseWriter, r *http.Request) error

func getHandlerFunc(f APIFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := f(w, r); err != nil {
			writeJSON(w, http.StatusBadRequest, APIError{err.Error()})
		}
	}
}

func writeJSON(w http.ResponseWriter, status int, v any) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	return json.NewEncoder(w).Encode(v)
}

type APIError struct {
	Error string `json:"error"`
}

func (s *APIServer) comparePassword(r *http.Request, id int64) error {
	// get the password from request body
	req := new(PasswordReq)
	if err := json.NewDecoder(r.Body).Decode(req); err != nil {
		return nil
	}

	if req.Password == "" {
		return fmt.Errorf("failed")
	}

	// get the enc password of the account
	enc_pass, err := s.store.GetAccountByID(id)
	if err != nil {
		return err
	}

	// compare password and enc password
	if err := bcrypt.CompareHashAndPassword([]byte(*enc_pass), []byte(req.Password)); err != nil {
		return fmt.Errorf("failed")
	}

	return nil
}

func (s *APIServer) validateJWT(r *http.Request, id int64) error {
	tokenStr := r.Header.Get("jwt-token")

	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		return secret, nil
	})
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return fmt.Errorf("your token is expired, login again")
		}
		return err
	}

	// check if JWT subject and id same
	idStr, err := token.Claims.GetSubject()
	if err != nil {
		return err
	}
	if idStr != strconv.FormatInt(id, 10) {
		return fmt.Errorf("failed")
	}

	return nil
}

func createJWT(idStr string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": idStr,
		"exp": time.Now().Add(time.Second).Unix(),
	})
	tokenString, err := token.SignedString(secret)
	if err != nil {
		return "", err
	}
	return tokenString, nil
}
