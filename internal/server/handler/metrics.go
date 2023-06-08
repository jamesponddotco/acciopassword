package handler

import (
	"encoding/json"
	"net/http"

	"git.sr.ht/~jamesponddotco/acciopassword/internal/cerrors"
	"git.sr.ht/~jamesponddotco/acciopassword/internal/database"
	"git.sr.ht/~jamesponddotco/acciopassword/internal/server/model"
	"git.sr.ht/~jamesponddotco/xstd-go/xnet/xhttp"
	"go.uber.org/zap"
)

// MetricsHandler is an HTTP handler for the /metrics endpoint.
type MetricsHandler struct {
	db     *database.DB
	logger *zap.Logger
}

// NewMetricsHandler creates a new MetricsHandler instance.
func NewMetricsHandler(db *database.DB, logger *zap.Logger) *MetricsHandler {
	return &MetricsHandler{
		db:     db,
		logger: logger,
	}
}

// ServeHTTP serves the /metrics endpoint.
func (h *MetricsHandler) ServeHTTP(w http.ResponseWriter, _ *http.Request) {
	var (
		countDiceware = h.db.Count(database.CounterTypeDiceware)
		countRandom   = h.db.Count(database.CounterTypeRandom)
		countPIN      = h.db.Count(database.CounterTypePIN)
		countTotal    = countDiceware + countRandom + countPIN
		counter       = model.NewMetrics(countRandom, countDiceware, countPIN, countTotal)
	)

	counterJSON, _ := json.Marshal(counter)

	w.Header().Set(xhttp.ContentType, xhttp.ApplicationJSON)

	_, err := w.Write(counterJSON)
	if err != nil {
		h.logger.Error("Failed to write access counter JSON to response", zap.Error(err))

		cerrors.JSON(w, h.logger, cerrors.ErrorResponse{
			Code:    http.StatusInternalServerError,
			Message: "Failed to write access counter JSON to response.",
		})

		return
	}
}
