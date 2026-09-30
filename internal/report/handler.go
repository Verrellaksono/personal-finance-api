package report

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"personal-finance/internal/auth"
)

type ServiceContract interface {
	GenerateMonthlySummary(ctx context.Context, userID int64, monthStr string) (*MonthlySummaryResponse, error)
}

type Handler struct {
	service ServiceContract
}

func NewHandler(service ServiceContract) *Handler {
	return &Handler{
		service: service,
	}
}

func writeJSON(w http.ResponseWriter, code int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(data)
}

func (h *Handler) MonthlySummary(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
		return
	}

	userID, ok := auth.GetUserIDFromContext(r.Context())
	if !ok {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	monthParam := r.URL.Query().Get("month")
	if monthParam == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "query parameter 'month' is required (format: YYYY-MM)"})
		return
	}

	summary, err := h.service.GenerateMonthlySummary(r.Context(), userID, monthParam)
	if err != nil {
		if errors.Is(err, ErrInvalidMonth) {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
			return
		} else {
			log.Printf("[ERROR] MonthlySummary failed: %v", err)
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal server error"})
			return
		}
	}

	writeJSON(w, http.StatusOK, summary)
}
