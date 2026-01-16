package db

import (
	"errors"
	"sync"

	"go-backend/internal/domain"
)

var ErrTaskNotFound = errors.New("task not found")

type DataStore struct {
	mu    sync.RWMutex
	users []domain.User
	tasks []domain.Task
}

func NewDataStore() *DataStore {
	return &DataStore{
		users: []domain.User{
			{ID: 1, Name: "John Doe", Email: "john@example.com", Role: "developer"},
			{ID: 2, Name: "Jane Smith", Email: "jane@example.com", Role: "designer"},
			{ID: 3, Name: "Bob Johnson", Email: "bob@example.com", Role: "manager"},
		},
		tasks: []domain.Task{
			{ID: 1, Title: "Implement authentication", Status: "pending", UserID: 1},
			{ID: 2, Title: "Design user interface", Status: "in-progress", UserID: 2},
			{ID: 3, Title: "Review code changes", Status: "completed", UserID: 3},
		},
	}
}

func (ds *DataStore) ListUsers() []domain.User {
	ds.mu.RLock()
	defer ds.mu.RUnlock()

	users := make([]domain.User, len(ds.users))
	copy(users, ds.users)
	return users
}

func (ds *DataStore) GetUserByID(id int) *domain.User {
	ds.mu.RLock()
	defer ds.mu.RUnlock()

	for i := range ds.users {
		if ds.users[i].ID == id {
			userCopy := ds.users[i]
			return &userCopy
		}
	}

	return nil
}

func (ds *DataStore) CreateUser(user domain.User) (domain.User, error) {
	ds.mu.Lock()
	defer ds.mu.Unlock()

	maxID := 0
	for _, existing := range ds.users {
		if existing.ID > maxID {
			maxID = existing.ID
		}
	}

	user.ID = maxID + 1
	ds.users = append(ds.users, user)

	return user, nil
}

func (ds *DataStore) ListTasks() []domain.Task {
	ds.mu.RLock()
	defer ds.mu.RUnlock()

	tasks := make([]domain.Task, len(ds.tasks))
	copy(tasks, ds.tasks)
	return tasks
}

func (ds *DataStore) CreateTask(task domain.Task) (domain.Task, error) {
	ds.mu.Lock()
	defer ds.mu.Unlock()

	maxID := 0
	for _, existing := range ds.tasks {
		if existing.ID > maxID {
			maxID = existing.ID
		}
	}

	task.ID = maxID + 1
	ds.tasks = append(ds.tasks, task)

	return task, nil
}

func (ds *DataStore) UpdateTask(id int, title *string, status *string, userID *int) (*domain.Task, error) {
	ds.mu.Lock()
	defer ds.mu.Unlock()

	for i := range ds.tasks {
		if ds.tasks[i].ID != id {
			continue
		}

		if title != nil {
			ds.tasks[i].Title = *title
		}
		if status != nil {
			ds.tasks[i].Status = *status
		}
		if userID != nil {
			ds.tasks[i].UserID = *userID
		}

		taskCopy := ds.tasks[i]
		return &taskCopy, nil
	}

	return nil, ErrTaskNotFound
}

func (ds *DataStore) GetStats() domain.Stats {
	ds.mu.RLock()
	defer ds.mu.RUnlock()

	stats := domain.Stats{}
	stats.Users.Total = len(ds.users)
	stats.Tasks.Total = len(ds.tasks)

	for _, task := range ds.tasks {
		switch task.Status {
		case "pending":
			stats.Tasks.Pending++
		case "in-progress":
			stats.Tasks.InProgress++
		case "completed":
			stats.Tasks.Completed++
		}
	}

	return stats
}
