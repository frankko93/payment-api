package http

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/franco/payment-api/internal/application/command"
)

// PaymentHandler handles HTTP requests for payments
type PaymentHandler struct {
	createPaymentService *command.CreatePaymentService
}

// NewPaymentHandler creates a new PaymentHandler
func NewPaymentHandler(createPaymentService *command.CreatePaymentService) *PaymentHandler {
	return &PaymentHandler{
		createPaymentService: createPaymentService,
	}
}

// CreatePaymentRequest represents the HTTP request body
type CreatePaymentRequest struct {
	UserID         string  `json:"userId"`
	Amount         float64 `json:"amount"`
	Currency       string  `json:"currency"`
	ServiceID      string  `json:"serviceId"`
	IdempotencyKey string  `json:"idempotencyKey"`
	ClientID       string  `json:"clientId"`
}

// CreatePaymentResponse represents the HTTP response body
type CreatePaymentResponse struct {
	PaymentID string `json:"paymentId"`
	Status    string `json:"status"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error string `json:"error"`
}

// HandleCreatePayment handles POST /payments requests
func (h *PaymentHandler) HandleCreatePayment(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.respondError(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req CreatePaymentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondError(w, "invalid request body", http.StatusBadRequest)
		return
	}

	// Execute service
	result, err := h.createPaymentService.Execute(r.Context(), command.CreatePaymentRequest{
		UserID:         req.UserID,
		Amount:         req.Amount,
		Currency:       req.Currency,
		ServiceID:      req.ServiceID,
		IdempotencyKey: req.IdempotencyKey,
		ClientID:       req.ClientID,
	})

	if err != nil {
		log.Printf("Error creating payment: %v", err)
		h.respondError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.respondJSON(w, CreatePaymentResponse{
		PaymentID: result.PaymentID,
		Status:    result.Status,
	}, http.StatusOK)
}

func (h *PaymentHandler) respondJSON(w http.ResponseWriter, data interface{}, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(data)
}

func (h *PaymentHandler) respondError(w http.ResponseWriter, message string, statusCode int) {
	h.respondJSON(w, ErrorResponse{Error: message}, statusCode)
}
