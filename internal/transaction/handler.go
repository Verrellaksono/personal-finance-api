package transaction

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
)

type ServiceContract interface {
	CreateTransaction(ctx context.Context, input CreateTransactionInput) (*Transaction, error)
}

type Handler struct {
	service ServiceContract
}

func NewHandler(service ServiceContract) *Handler {
	return &Handler{service: service}
}

func WriteJSON(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(data)
}

func (h *Handler) CreateTransaction(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		WriteJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "Method not allowed."})
		return
	}

	var req CreateTransactionInput
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		WriteJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	result, err := h.service.CreateTransaction(r.Context(), req)
	if err != nil {
		switch {
		case errors.Is(err, ErrInvalidAmount), errors.Is(err, ErrInvalidType):
			WriteJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		case errors.Is(err, ErrAccountNotFounrd):
			WriteJSON(w, http.StatusNotFound, map[string]string{"error": err.Error()})
		case errors.Is(err, ErrInsufficientBalance):
			WriteJSON(w, http.StatusUnprocessableEntity, map[string]string{"error": err.Error()})
		default:
			log.Printf("[ERROR] CreateTransaction failed: %v", err)
			WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": "Internal Server Error."})
		}
		return
	}

	WriteJSON(w, http.StatusCreated, result)
}
