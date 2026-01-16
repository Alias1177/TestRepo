package domain

type User struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
	Role  string `json:"role"`
}

type Task struct {
	ID     int    `json:"id"`
	Title  string `json:"title"`
	Status string `json:"status"`
	UserID int    `json:"userId"`
}

type Stats struct {
	Users UsersStats `json:"users"`
	Tasks TasksStats `json:"tasks"`
}

type UsersStats struct {
	Total int `json:"total"`
}

type TasksStats struct {
	Total      int `json:"total"`
	Pending    int `json:"pending"`
	InProgress int `json:"inProgress"`
	Completed  int `json:"completed"`
}
