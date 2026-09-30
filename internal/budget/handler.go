package budget

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"personal-finance/internal/auth"
)

type ServiceContract interface {
	SetBudget(ctx context.Context, input SetBudgetInput) (*Budget, error)
	CheckBudgetProgress(ctx context.Context, userID int64, monthStr string) ([]BudgetProgress, error)
}

type Handler struct {
	service ServiceContract
}

func NewHandler(service ServiceContract) *Handler {
	return &Handler{service: service}
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

func (h *Handler) SetBudget(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
		return
	}

	userID, ok := auth.GetUserIDFromContext(r.Context())
	if !ok {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unautorized"})
		return
	}

	var input SetBudgetInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid json body: " + err.Error()})
		return
	}

	input.UserID = userID

	b, err := h.service.SetBudget(r.Context(), input)
	if err != nil {
		switch {
		case errors.Is(err, ErrInvalidAmount), errors.Is(err, ErrInvalidMonth):
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		default:
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal server error"})
		}
		return
	}

	writeJSON(w, http.StatusOK, b)
}

func (h *Handler) GetBudgetProgress(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
		return
	}

	userID, ok := auth.GetUserIDFromContext(r.Context())
	if !ok {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unautorized"})
		return
	}

	monthParam := r.URL.Query().Get("month")
	if monthParam == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "query parameter 'month' is required (format: YYYY-MM)"})
		return
	}

	progressList, err := h.service.CheckBudgetProgress(r.Context(), userID, monthParam)
	if err != nil {
		if errors.Is(err, ErrInvalidMonth) {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		} else {
			log.Printf("[ERROR] SetBudget failed: %v", err)
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal server error"})
		}
		return
	}

	writeJSON(w, http.StatusOK, progressList)
}
