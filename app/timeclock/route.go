package timeclock

import (
	"github.com/ariesmaulana/payroll/lib/middleware"
	"github.com/go-chi/chi/v5"
)

func RegisterRoutes(r chi.Router, handler *Handler) {
	r.Route("/timeclock", func(r chi.Router) {

		// Private endpoint - require auth middleware
		r.Group(func(r chi.Router) {
			r.Use(middleware.AuthMiddleware)

			// (attendance)
			r.Post("/add-period", handler.AddAttendancePeriod)
			r.Post("/clock-in", handler.SubmitAttendance)
			r.Post("/clock-out", handler.CheckoutAttendance)

			//  (overtime)
			r.Post("/overtime", handler.AddOvertime)

			//reimbursement
			r.Post("/reimbursement", handler.SubmitReimbursement)

			// (payroll)
			r.Post("/payroll/run", handler.RunPayroll)

			// (payslip)
			r.Get("/payslip/self", handler.GenerateSelfPaySlip)
			r.Get("/payslip/all", handler.GenerateAllPaySlips)
		})
	})
}
