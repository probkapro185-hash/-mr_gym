package handler

import (
	"net/http"
	"time"

	"crm_gym/internal/dto"
	"crm_gym/internal/middleware"
	"crm_gym/internal/service"
)

type PaymentHandler struct {
	paymentSvc *service.PaymentService
}

func NewPaymentHandler(paymentSvc *service.PaymentService) *PaymentHandler {
	return &PaymentHandler{paymentSvc: paymentSvc}
}

// GET /api/v1/payments/my — история платежей текущего пользователя
func (h *PaymentHandler) GetMyPayments(w http.ResponseWriter, r *http.Request) {
	userID, _ := middleware.GetUserID(r.Context())
	from, to := parsePaymentDateRange(r)

	payments, err := h.paymentSvc.ListByUser(r.Context(), userID, from, to)
	if err != nil {
		handleServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, dto.ListResponse{Items: payments, Total: len(payments)})
}

// GET /api/v1/payments/user/{id} — история платежей пользователя (менеджер/админ)
func (h *PaymentHandler) GetUserPayments(w http.ResponseWriter, r *http.Request) {
	userID, err := pathParamInt64(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}
	from, to := parsePaymentDateRange(r)
	payments, err := h.paymentSvc.ListByUser(r.Context(), userID, from, to)
	if err != nil {
		handleServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, dto.ListResponse{Items: payments, Total: len(payments)})
}

// GET /api/v1/payments — все финансы (только админ)
func (h *PaymentHandler) GetAll(w http.ResponseWriter, r *http.Request) {
	from, to := parsePaymentDateRange(r)
	limit := queryInt(r, "limit", 50)
	offset := queryInt(r, "offset", 0)

	payments, total, err := h.paymentSvc.ListAll(r.Context(), from, to, limit, offset)
	if err != nil {
		handleServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, dto.ListResponse{Items: payments, Total: total})
}

// POST /api/v1/payments/deposit — пополнение баланса (менеджер/админ)
func (h *PaymentHandler) Deposit(w http.ResponseWriter, r *http.Request) {
	var req dto.DepositRequest
	if err := decode(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	payment, err := h.paymentSvc.Deposit(r.Context(), req)
	if err != nil {
		handleServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, payment)
}

// POST /api/v1/payments/charge — списание с баланса (только админ)
func (h *PaymentHandler) Charge(w http.ResponseWriter, r *http.Request) {
	var req dto.ChargeRequest
	if err := decode(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	payment, err := h.paymentSvc.Charge(r.Context(), req)
	if err != nil {
		handleServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, payment)
}

func parsePaymentDateRange(r *http.Request) (time.Time, time.Time) {
	now := time.Now()
	fromStr := queryString(r, "from")
	toStr := queryString(r, "to")

	parse := func(s string, fallback time.Time) time.Time {
		if s == "" {
			return fallback
		}
		if t, err := time.Parse("2006-01-02", s); err == nil {
			return t
		}
		if t, err := time.Parse(time.RFC3339, s); err == nil {
			return t
		}
		return fallback
	}

	from := parse(fromStr, time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location()))
	to := parse(toStr, from.AddDate(0, 1, 0))
	return from, to
}
