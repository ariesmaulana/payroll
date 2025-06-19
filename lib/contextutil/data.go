package contextutil

import (
	"context"

	"github.com/ariesmaulana/payroll/data"
)

// Trace struct to hold trace information
type Trace struct {
	TraceID string
	Method  string
	Path    string
	Headers map[string][]string
	Body    string
}

// AuthUser untuk menyimpan info user hasil verifikasi JWT
type AuthUser struct {
	Id       int
	Username string
	Role     data.UserRole
}

//
// AUTH USER UTILITY
//

// Inject user login ke dalam context (dipakai di middleware auth)
func WithUser(ctx context.Context, user *AuthUser) context.Context {
	return context.WithValue(ctx, authUserKey, user)
}

// Ambil user login dari context (dipakai di handler)
func GetUser(ctx context.Context) (*AuthUser, bool) {
	user, ok := ctx.Value(authUserKey).(*AuthUser)
	return user, ok
}
