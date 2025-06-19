package user

import "github.com/go-chi/chi/v5"

func RegisterRoutes(r chi.Router, h *Handler) {
	r.Route("/users", func(r chi.Router) {
		r.Post("/login", h.Login)
	})
}
