package data

import "time"

type UserRole string

const (
	RAdmin    UserRole = "admin"
	REmployee UserRole = "employee"
)

type User struct {
	Id         int
	Fullname   string
	Username   string
	Email      string
	Password   string
	Role       UserRole
	BaseSalary int
	JoinDate   time.Time

	CreatedAt time.Time
	UpdatedAt time.Time
}
