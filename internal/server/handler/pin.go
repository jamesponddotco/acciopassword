package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"git.sr.ht/~jamesponddotco/acciopassword/internal/cerrors"
	"git.sr.ht/~jamesponddotco/acciopassword/internal/database"
	"git.sr.ht/~jamesponddotco/acciopassword/internal/server/model"
	"git.sr.ht/~jamesponddotco/acopw-go"
	"git.sr.ht/~jamesponddotco/xstd-go/xnet/xhttp"
	"go.uber.org/zap"
)

const (
	// MaxPINLength is the maximum length of a PIN.
	MaxPINLength int = 64
)

// PINHandler is an HTTP handler for the /pin endpoint.
type PINHandler struct {
	db     *database.DB
	logger *zap.Logger
}

// NewPINHandler returns a new PINHandler instance.
func NewPINHandler(db *database.DB, logger *zap.Logger) *PINHandler {
	return &PINHandler{
		db:     db,
		logger: logger,
	}
}

// ServeHTTP handles HTTP requests for the /pin endpoint.
func (h *PINHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var (
		length = acopw.DefaultPINLength
		err    error
	)

	if r.URL.Query().Get("length") != "" {
		length, err = strconv.Atoi(r.URL.Query().Get("length"))
		if err != nil {
			h.logger.Error("error parsing PIN length", zap.Error(err))

			cerrors.JSON(w, h.logger, cerrors.ErrorResponse{
				Code:    http.StatusInternalServerError,
				Message: "Cannot parse the given PIN length. Please provide a valid integer.",
			})

			return
		}

		if length < 1 {
			length = acopw.DefaultPINLength
		}

		if length > MaxPINLength {
			h.logger.Error("PIN length is too long", zap.Int("length", length))

			cerrors.JSON(w, h.logger, cerrors.ErrorResponse{
				Code:    http.StatusBadRequest,
				Message: "The given PIN length is too long. Please provide a length less than or equal to " + strconv.Itoa(MaxPINLength) + ".",
			})

			return
		}
	}

	var (
		pin = &acopw.PIN{
			Length: length,
		}
		password    = pin.Generate()
		contentType = r.Header.Get(xhttp.ContentType)
	)

	if contentType == xhttp.ApplicationJSON {
		w.Header().Set(xhttp.ContentType, xhttp.ApplicationJSON)

		var (
			passwordModel   = model.NewPIN(password)
			passwordJSON, _ = json.Marshal(passwordModel)
		)

		_, err = w.Write(passwordJSON)
		if err != nil {
			h.logger.Error("error writing response", zap.Error(err))

			cerrors.JSON(w, h.logger, cerrors.ErrorResponse{
				Code:    http.StatusInternalServerError,
				Message: "Cannot write response. Please try again later.",
			})

			return
		}
	} else {
		w.Header().Set(xhttp.ContentType, xhttp.TextPlain)

		_, err := w.Write([]byte(password))
		if err != nil {
			h.logger.Error("error writing response", zap.Error(err))

			cerrors.JSON(w, h.logger, cerrors.ErrorResponse{
				Code:    http.StatusInternalServerError,
				Message: "Cannot write response. Please try again later.",
			})

			return
		}
	}

	go func() {
		if err := h.db.Increment(database.CounterTypePIN); err != nil {
			h.logger.Error("Failed to increment access counter", zap.Error(err))
		}
	}()
}
