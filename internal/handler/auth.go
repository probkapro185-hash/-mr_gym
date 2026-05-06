package handler

import (
	"net/http"

	"crm_gym/internal/dto"
	"crm_gym/internal/service"
)

type AuthHandler struct {
	authSvc *service.AuthService
}

func NewAuthHandler(authSvc *service.AuthService) *AuthHandler {
	return &AuthHandler{authSvc: authSvc}
}

// POST /api/v1/auth/apply — подача заявки (публичный эндпоинт)
func (h *AuthHandler) Apply(w http.ResponseWriter, r *http.Request) {
	var req dto.ApplicationRequest
	if err := decode(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if err := h.authSvc.SubmitApplication(r.Context(), req); err != nil {
		handleServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, dto.SuccessResponse{
		Message: "Ваша заявка принята. Ожидайте одобрения менеджером.",
	})
}

// POST /api/v1/auth/login — вход
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req dto.LoginRequest
	if err := decode(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	tokens, err := h.authSvc.Login(r.Context(), req)
	if err != nil {
		handleServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, tokens)
}

// POST /api/v1/auth/refresh — обновление токена
func (h *AuthHandler) Refresh(w http.ResponseWriter, r *http.Request) {
	var req dto.RefreshRequest
	if err := decode(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	tokens, err := h.authSvc.Refresh(r.Context(), req)
	if err != nil {
		handleServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, tokens)
}
