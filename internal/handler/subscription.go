package handler

import (
	"net/http"

	"crm_gym/internal/dto"
	"crm_gym/internal/middleware"
	"crm_gym/internal/service"
)

type SubscriptionHandler struct {
	subSvc *service.SubscriptionService
}

func NewSubscriptionHandler(subSvc *service.SubscriptionService) *SubscriptionHandler {
	return &SubscriptionHandler{subSvc: subSvc}
}

// GET /api/v1/subscriptions/my — абонементы текущего пользователя
func (h *SubscriptionHandler) GetMy(w http.ResponseWriter, r *http.Request) {
	userID, _ := middleware.GetUserID(r.Context())
	subs, err := h.subSvc.ListByUser(r.Context(), userID)
	if err != nil {
		handleServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, dto.ListResponse{Items: subs, Total: len(subs)})
}

// GET /api/v1/subscriptions/user/{id} — абонементы клиента (менеджер/админ)
func (h *SubscriptionHandler) GetByUserID(w http.ResponseWriter, r *http.Request) {
	userID, err := pathParamInt64(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}
	subs, err := h.subSvc.ListByUser(r.Context(), userID)
	if err != nil {
		handleServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, dto.ListResponse{Items: subs, Total: len(subs)})
}

// POST /api/v1/subscriptions — создание абонемента (менеджер/админ)
func (h *SubscriptionHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req dto.CreateSubscriptionRequest
	if err := decode(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	sub, err := h.subSvc.Create(r.Context(), req)
	if err != nil {
		handleServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, sub)
}

// DELETE /api/v1/subscriptions/{id} — удаление абонемента (только админ)
func (h *SubscriptionHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := pathParamInt64(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}
	if err := h.subSvc.Delete(r.Context(), id); err != nil {
		handleServiceError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
