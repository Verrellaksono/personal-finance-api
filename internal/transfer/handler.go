package transfer

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"personal-finance/internal/auth"
	"strings"
)

type ServiceContract interface {
	Processtransfer(ctx context.Context, input CreateTransferInput) (*TransferResult, error)
}

type Handler struct {
	service ServiceContract
}

func NewHandler(service ServiceContract) *Handler {
	return &Handler{service: service}
}

func (h *Handler) Transfer(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "Method not allowed."})
		return
	}

	userID, ok := auth.GetUserIDFromContext(r.Context())
	if !ok {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
	}

	// Read Idempotency Key from Header
	idempotencyKey := strings.TrimSpace(r.Header.Get("Idempotency-Key"))
	if idempotencyKey == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Idempotency-Key header is required"})
		return
	}

	var input CreateTransferInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid JSON body: " + err.Error()})
		return
	}

	input.UserID = userID
	input.IdempotencyKey = idempotencyKey

	res, err := h.service.Processtransfer(r.Context(), input)
	if err != nil {
		switch {
		case errors.Is(err, ErrInvalidAmount), errors.Is(err, ErrSameAccount), errors.Is(err, ErrMissingKey):
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		case errors.Is(err, ErrAccountNotFound):
			writeJSON(w, http.StatusNotFound, map[string]string{"error": err.Error()})
		case errors.Is(err, ErrInsufficientBalance):
			writeJSON(w, http.StatusUnprocessableEntity, map[string]string{"error": err.Error()})
		case errors.Is(err, ErrIdempotencyConflict):
			writeJSON(w, http.StatusConflict, map[string]string{"error": err.Error()})
		case strings.Contains(err.Error(), "unauthorized"):
			writeJSON(w, http.StatusForbidden, map[string]string{"error": err.Error()})
		default:
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal server error"})
		}
		return
	}

	writeJSON(w, http.StatusOK, res)
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}
