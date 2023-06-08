package handler

import (
	"net/http"

	"git.sr.ht/~jamesponddotco/acciopassword/internal/cerrors"
	"git.sr.ht/~jamesponddotco/xstd-go/xnet/xhttp"
	"go.uber.org/zap"
)

const pong string = "pong"

// PingHandler is an HTTP handler for the /heartbeat endpoint.
type PingHandler struct {
	logger *zap.Logger
}

// NewPingHandler creates a new PingHandler instance.
func NewPingHandler(logger *zap.Logger) *PingHandler {
	return &PingHandler{
		logger: logger,
	}
}

// ServeHTTP serves the /heartbeat endpoint.
func (h *PingHandler) ServeHTTP(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set(xhttp.ContentType, xhttp.TextPlain)

	_, err := w.Write([]byte(pong))
	if err != nil {
		h.logger.Error("Failed to write response", zap.Error(err))

		cerrors.JSON(w, h.logger, cerrors.ErrorResponse{
			Code:    http.StatusInternalServerError,
			Message: "Failed to write response. Please try again later.",
		})

		return
	}
}
