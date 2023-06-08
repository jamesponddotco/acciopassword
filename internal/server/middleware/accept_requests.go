package middleware

import (
	"fmt"
	"net/http"

	"git.sr.ht/~jamesponddotco/acciopassword/internal/cerrors"
	"go.uber.org/zap"
)

func AcceptRequests(logger *zap.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet && r.Method != http.MethodHead && r.Method != http.MethodOptions {
			cerrors.JSON(w, logger, cerrors.ErrorResponse{
				Code:    http.StatusMethodNotAllowed,
				Message: fmt.Sprintf("Method %s not allowed. Must be GET, HEAD, or OPTIONS.", r.Method),
			})

			return
		}

		next.ServeHTTP(w, r)
	})
}
