package server

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"
	"strings"

	"go-backend/config"
	"go-backend/internal/domain"
	"go-backend/internal/middleware"
	"go-backend/internal/service"
)

type HealthResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

type UsersResponse struct {
	Users []domain.User `json:"users"`
	Count int           `json:"count"`
}

type TasksResponse struct {
	Tasks []domain.Task `json:"tasks"`
	Count int           `json:"count"`
}

type Server struct {
	service *service.Service
	config  *config.Config
}

func New(svc *service.Service, cfg *config.Config) *Server {
	return &Server{
		service: svc,
		config:  cfg,
	}
}

func (s *Server) Start() error {
	wrap := s.registerRoutes()

	log.Printf("Go backend server starting on http://%s:%s", s.config.Host, s.config.Port)
	return http.ListenAndServe(s.config.Host+":"+s.config.Port, wrap)
}

func (s *Server) registerRoutes() http.Handler {
	mux := http.NewServeMux()
	wrap := middleware.Logger(mux)

	mux.HandleFunc("/health", s.handleHealth)
	mux.HandleFunc("/api/users", s.handleUsers)
	mux.HandleFunc("/api/users/", s.handleUserByID)
	mux.HandleFunc("/api/tasks", s.handleTasks)
	mux.HandleFunc("/api/tasks/", s.handleTaskByID)
	mux.HandleFunc("/api/stats", s.handleStats)

	return wrap
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	respondJSON(w, http.StatusOK, HealthResponse{Status: "ok", Message: "Go backend is running"})
}

func (s *Server) handleUsers(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	switch r.Method {
	case http.MethodGet:
		users := s.service.ListUsers()
		respondJSON(w, http.StatusOK, UsersResponse{Users: users, Count: len(users)})
	case http.MethodPost:
		var user domain.User
		if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}
		if user.Name == "" || user.Email == "" || user.Role == "" || !strings.Contains(user.Email, "@") {
			http.Error(w, "Missing required fields", http.StatusBadRequest)
			return
		}

		users := s.service.ListUsers()
		for _, u := range users {
			if u.Email == user.Email {
				http.Error(w, "Email already exists", http.StatusBadRequest)
				return
			}
		}

		created, err := s.service.CreateUser(user)
		if err != nil {
			http.Error(w, "Failed to create user", http.StatusInternalServerError)
			return
		}

		respondJSON(w, http.StatusCreated, created)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *Server) handleUserByID(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	idStr := strings.TrimPrefix(r.URL.Path, "/api/users/")
	userID, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	user, err := s.service.GetUserByID(userID)
	if err != nil {
		status := http.StatusInternalServerError
		if errors.Is(err, service.ErrNotFound) {
			status = http.StatusNotFound
		}
		http.Error(w, "User not found", status)
		return
	}

	respondJSON(w, http.StatusOK, user)
}

func (s *Server) handleTasks(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	switch r.Method {
	case http.MethodGet:
		status := r.URL.Query().Get("status")
		userID := r.URL.Query().Get("userId")
		tasks := s.service.ListTasks(status, userID)
		respondJSON(w, http.StatusOK, TasksResponse{Tasks: tasks, Count: len(tasks)})
	case http.MethodPost:
		var task domain.Task
		if err := json.NewDecoder(r.Body).Decode(&task); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}
		if task.Title == "" || task.Status == "" || task.UserID == 0 {
			http.Error(w, "Missing required fields", http.StatusBadRequest)
			return
		}

		created, err := s.service.CreateTask(task)
		if err != nil {
			status := http.StatusInternalServerError
			switch err {
			case service.ErrInvalidStatus, service.ErrInvalidUser:
				status = http.StatusBadRequest
			}
			http.Error(w, err.Error(), status)
			return
		}

		respondJSON(w, http.StatusCreated, created)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *Server) handleTaskByID(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	idStr := strings.TrimPrefix(r.URL.Path, "/api/tasks/")
	taskID, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid task ID", http.StatusBadRequest)
		return
	}

	var payload struct {
		Title  *string `json:"title"`
		Status *string `json:"status"`
		UserID *int    `json:"userId"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	updated, err := s.service.UpdateTask(taskID, service.TaskUpdateInput{
		Title:  payload.Title,
		Status: payload.Status,
		UserID: payload.UserID,
	})
	if err != nil {
		status := http.StatusInternalServerError
		switch err {
		case service.ErrNotFound:
			status = http.StatusNotFound
		case service.ErrInvalidStatus, service.ErrInvalidUser, service.ErrNoUpdateFields:
			status = http.StatusBadRequest
		}
		http.Error(w, err.Error(), status)
		return
	}

	respondJSON(w, http.StatusOK, updated)
}

func (s *Server) handleStats(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	stats := s.service.GetStats()
	respondJSON(w, http.StatusOK, stats)
}

func respondJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}
