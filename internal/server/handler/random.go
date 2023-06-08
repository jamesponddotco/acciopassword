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
	// MaxRandomLength is the maximum length of a diceware password.
	MaxRandomLength int = 256
)

// RandomHandler is an HTTP handler for the /diceware endpoint.
type RandomHandler struct {
	db     *database.DB
	logger *zap.Logger
}

// NewRandomHandler returns a new RandomHandler instance.
func NewRandomHandler(db *database.DB, logger *zap.Logger) *RandomHandler {
	return &RandomHandler{
		db:     db,
		logger: logger,
	}
}

// ServeHTTP handles HTTP requests for the /diceware endpoint.
func (h *RandomHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var (
		length       = acopw.DefaultRandomLength
		useLowercase = true
		useUppercase = true
		useNumbers   = true
		useSymbols   = true
		err          error
	)

	if r.URL.Query().Get("length") != "" {
		length, err = strconv.Atoi(r.URL.Query().Get("length"))
		if err != nil {
			h.logger.Error("error parsing password length", zap.Error(err))

			cerrors.JSON(w, h.logger, cerrors.ErrorResponse{
				Code:    http.StatusBadRequest,
				Message: "Cannot parse the given password length. Please provide a valid integer.",
			})

			return
		}

		if length < 1 {
			length = acopw.DefaultRandomLength
		}

		if length > MaxRandomLength {
			h.logger.Error("password length is too long", zap.Int("length", length))

			cerrors.JSON(w, h.logger, cerrors.ErrorResponse{
				Code:    http.StatusBadRequest,
				Message: "The given password length is too long. Please provide a length less than or equal to " + strconv.Itoa(MaxRandomLength) + ".",
			})

			return
		}
	}

	if r.URL.Query().Get("lowercase") != "" {
		useLowercase, err = strconv.ParseBool(r.URL.Query().Get("lowercase"))
		if err != nil {
			h.logger.Error("error parsing lowercase flag", zap.Error(err))

			cerrors.JSON(w, h.logger, cerrors.ErrorResponse{
				Code:    http.StatusBadRequest,
				Message: "Cannot parse the given lowercase flag. Please provide a valid boolean.",
			})

			return
		}
	}

	if r.URL.Query().Get("uppercase") != "" {
		useUppercase, err = strconv.ParseBool(r.URL.Query().Get("uppercase"))
		if err != nil {
			h.logger.Error("error parsing uppercase flag", zap.Error(err))

			cerrors.JSON(w, h.logger, cerrors.ErrorResponse{
				Code:    http.StatusBadRequest,
				Message: "Cannot parse the given uppercase flag. Please provide a valid boolean.",
			})

			return
		}
	}

	if r.URL.Query().Get("numbers") != "" {
		useNumbers, err = strconv.ParseBool(r.URL.Query().Get("numbers"))
		if err != nil {
			h.logger.Error("error parsing numbers flag", zap.Error(err))

			cerrors.JSON(w, h.logger, cerrors.ErrorResponse{
				Code:    http.StatusBadRequest,
				Message: "Cannot parse the given numbers flag. Please provide a valid boolean.",
			})

			return
		}
	}

	if r.URL.Query().Get("symbols") != "" {
		useSymbols, err = strconv.ParseBool(r.URL.Query().Get("symbols"))
		if err != nil {
			h.logger.Error("error parsing symbols flag", zap.Error(err))

			cerrors.JSON(w, h.logger, cerrors.ErrorResponse{
				Code:    http.StatusBadRequest,
				Message: "Cannot parse the given symbols flag. Please provide a valid boolean.",
			})

			return
		}
	}

	var (
		password = &acopw.Random{
			Length:     length,
			UseLower:   useLowercase,
			UseUpper:   useUppercase,
			UseNumbers: useNumbers,
			UseSymbols: useSymbols,
		}
		contentType = r.Header.Get(xhttp.ContentType)
	)

	if contentType == xhttp.ApplicationJSON {
		w.Header().Set(xhttp.ContentType, xhttp.ApplicationJSON)

		var (
			passwordModel   = model.NewRandomPassword(password.Generate())
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

		_, err = w.Write([]byte(password.Generate()))
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
		if err := h.db.Increment(database.CounterTypeRandom); err != nil {
			h.logger.Error("Failed to increment access counter", zap.Error(err))
		}
	}()
}
