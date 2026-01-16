package service

import (
	"errors"
	"strconv"

	"go-backend/internal/db"
	"go-backend/internal/domain"
)

var (
	ErrNotFound       = errors.New("resource not found")
	ErrInvalidStatus  = errors.New("invalid status")
	ErrInvalidUser    = errors.New("invalid user")
	ErrNoUpdateFields = errors.New("no fields to update")
)

type Store interface {
	ListUsers() []domain.User
	GetUserByID(id int) *domain.User
	CreateUser(user domain.User) (domain.User, error)

	ListTasks() []domain.Task
	CreateTask(task domain.Task) (domain.Task, error)
	UpdateTask(id int, title *string, status *string, userID *int) (*domain.Task, error)

	GetStats() domain.Stats
}

type Service struct {
	store Store
}

func New(store Store) *Service {
	return &Service{store: store}
}

func (s *Service) ListUsers() []domain.User {
	return s.store.ListUsers()
}

func (s *Service) GetUserByID(id int) (*domain.User, error) {
	user := s.store.GetUserByID(id)
	if user == nil {
		return nil, ErrNotFound
	}
	return user, nil
}

func (s *Service) CreateUser(user domain.User) (domain.User, error) {
	return s.store.CreateUser(user)
}

func (s *Service) ListTasks(status, userID string) []domain.Task {
	tasks := s.store.ListTasks()
	if status == "" && userID == "" {
		return tasks
	}

	var filtered []domain.Task
	var uid int
	var err error
	if userID != "" {
		uid, err = strconv.Atoi(userID)
		if err != nil {
			return filtered
		}
	}

	for _, task := range tasks {
		if status != "" && task.Status != status {
			continue
		}
		if userID != "" && task.UserID != uid {
			continue
		}
		filtered = append(filtered, task)
	}

	return filtered
}

func (s *Service) CreateTask(task domain.Task) (domain.Task, error) {
	if !isValidStatus(task.Status) {
		return domain.Task{}, ErrInvalidStatus
	}

	if s.store.GetUserByID(task.UserID) == nil {
		return domain.Task{}, ErrInvalidUser
	}

	return s.store.CreateTask(task)
}

type TaskUpdateInput struct {
	Title  *string
	Status *string
	UserID *int
}

func (s *Service) UpdateTask(id int, input TaskUpdateInput) (*domain.Task, error) {
	if input.Title == nil && input.Status == nil && input.UserID == nil {
		return nil, ErrNoUpdateFields
	}

	if input.Status != nil && !isValidStatus(*input.Status) {
		return nil, ErrInvalidStatus
	}

	if input.UserID != nil {
		if s.store.GetUserByID(*input.UserID) == nil {
			return nil, ErrInvalidUser
		}
	}

	updatedTask, err := s.store.UpdateTask(id, input.Title, input.Status, input.UserID)
	if err != nil {
		if errors.Is(err, db.ErrTaskNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	return updatedTask, nil
}

func (s *Service) GetStats() domain.Stats {
	return s.store.GetStats()
}

func isValidStatus(status string) bool {
	switch status {
	case "pending", "in-progress", "completed":
		return true
	default:
		return false
	}
}
