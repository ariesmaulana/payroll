package lib

import (
	"context"

	"github.com/ariesmaulana/payroll/lib/contextutil"
)

type ServiceInterface interface {
	Login(ctx context.Context, in *LoginIn) *LoginOut

	UserSalary(ctx context.Context, in *UserSalaryIn) *UserSalaryOut
}

type LoginIn struct {
	Trace    *contextutil.Trace
	UserName string
	Password string
}

type LoginOut struct {
	Success bool
	Message string

	Token string
}

type UserSalaryIn struct {
	Trace *contextutil.Trace
}

type UserSalaryOut struct {
	Success bool
	Message string

	// Result key is userId and value is baseSalary
	Result map[int]int
}
