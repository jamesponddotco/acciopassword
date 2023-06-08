package middleware

import (
	"net/http"

	"git.sr.ht/~jamesponddotco/acciopassword/internal/cerrors"
	"go.uber.org/zap"
)

// UserAgent ensures that the request has a valid user agent.
func UserAgent(logger *zap.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.UserAgent() == "" {
			cerrors.JSON(w, logger, cerrors.ErrorResponse{
				Code:    http.StatusBadRequest,
				Message: "User agent is missing. Please provide a valid user agent.",
			})

			return
		}

		next.ServeHTTP(w, r)
	})
}
