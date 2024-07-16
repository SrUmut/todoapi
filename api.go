package main

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"golang.org/x/crypto/bcrypt"
)

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
	router.HandleFunc("/account/{id}", getHandlerFunc(s.handleAccountByID))

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

	// get the password from request body
	req := new(PasswordReq)
	if err := json.NewDecoder(r.Body).Decode(req); err != nil {
		return nil
	}

	// get the enc password of the account
	enc_pass, err := s.store.GetAccountByID(id)
	if err != nil {
		return err
	}

	// compare password and enc password
	if err := bcrypt.CompareHashAndPassword([]byte(*enc_pass), []byte(req.Password)); err != nil {
		return errors.New("failed")
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
