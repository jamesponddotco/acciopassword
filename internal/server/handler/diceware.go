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
	// DefaultDicewareSeparator is the default separator for diceware passwords.
	DefaultDicewareSeparator string = "-"

	// MaxDicewareLength is the maximum length of a diceware password.
	MaxDicewareLength int = 64
)

// DicewareHandler is an HTTP handler for the /diceware endpoint.
type DicewareHandler struct {
	db     *database.DB
	logger *zap.Logger
}

// NewDicewareHandler returns a new DicewareHandler instance.
func NewDicewareHandler(db *database.DB, logger *zap.Logger) *DicewareHandler {
	return &DicewareHandler{
		db:     db,
		logger: logger,
	}
}

// ServeHTTP handles HTTP requests for the /diceware endpoint.
func (h *DicewareHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var (
		length = acopw.DefaultDicewareLength
		err    error
	)

	if r.URL.Query().Get("length") != "" {
		length, err = strconv.Atoi(r.URL.Query().Get("length"))
		if err != nil {
			h.logger.Error("error parsing diceware length", zap.Error(err))

			cerrors.JSON(w, h.logger, cerrors.ErrorResponse{
				Code:    http.StatusInternalServerError,
				Message: "Cannot parse the given diceware length. Please provide a valid integer.",
			})

			return
		}

		if length < 1 {
			length = acopw.DefaultDicewareLength
		}

		if length > MaxDicewareLength {
			h.logger.Error("password length is too long", zap.Int("length", length))

			cerrors.JSON(w, h.logger, cerrors.ErrorResponse{
				Code:    http.StatusBadRequest,
				Message: "The given password length is too long. Please provide a length less than or equal to " + strconv.Itoa(MaxDicewareLength) + ".",
			})

			return
		}
	}

	var (
		capitalize = r.URL.Query().Get("capitalize") == "true"
		separator  = r.URL.Query().Get("separator")
	)

	if separator == "" {
		separator = DefaultDicewareSeparator
	}

	diceware := &acopw.Diceware{
		Separator:  separator,
		Capitalize: capitalize,
		Length:     length,
	}

	password, err := diceware.Generate()
	if err != nil {
		h.logger.Error("error generating diceware password", zap.Error(err))

		cerrors.JSON(w, h.logger, cerrors.ErrorResponse{
			Code:    http.StatusInternalServerError,
			Message: "Cannot generate diceware password. Please try again later.",
		})

		return
	}

	contentType := r.Header.Get(xhttp.ContentType)

	if contentType == xhttp.ApplicationJSON {
		w.Header().Set(xhttp.ContentType, xhttp.ApplicationJSON)

		var (
			passwordModel   = model.NewDicewarePassword(password)
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

		_, err = w.Write([]byte(password))
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
		if err := h.db.Increment(database.CounterTypeDiceware); err != nil {
			h.logger.Error("Failed to increment access counter", zap.Error(err))
		}
	}()
}
