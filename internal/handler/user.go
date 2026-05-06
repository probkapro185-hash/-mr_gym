package handler

import (
	"net/http"

	"crm_gym/internal/dto"
	"crm_gym/internal/middleware"
	"crm_gym/internal/repository"
	"crm_gym/internal/service"
)

type UserHandler struct {
	userSvc *service.UserService
}

func NewUserHandler(userSvc *service.UserService) *UserHandler {
	return &UserHandler{userSvc: userSvc}
}

// GET /api/v1/users/me — профиль текущего пользователя
func (h *UserHandler) GetMe(w http.ResponseWriter, r *http.Request) {
	userID, _ := middleware.GetUserID(r.Context())
	u, err := h.userSvc.GetMe(r.Context(), userID)
	if err != nil {
		handleServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, u)
}

// GET /api/v1/users — список пользователей (менеджер/админ)
func (h *UserHandler) List(w http.ResponseWriter, r *http.Request) {
	f := repository.UserFilter{
		Role:   queryString(r, "role"),
		Status: queryString(r, "status"),
		Search: queryString(r, "search"),
		Limit:  queryInt(r, "limit", 50),
		Offset: queryInt(r, "offset", 0),
	}
	users, total, err := h.userSvc.List(r.Context(), f)
	if err != nil {
		handleServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, dto.ListResponse{Items: users, Total: total})
}

// GET /api/v1/users/{id} — пользователь по ID (менеджер/админ)
func (h *UserHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id, err := pathParamInt64(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}
	u, err := h.userSvc.GetByID(r.Context(), id)
	if err != nil {
		handleServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, u)
}

// POST /api/v1/users — создание пользователя (только админ)
func (h *UserHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req dto.CreateUserRequest
	if err := decode(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	u, err := h.userSvc.CreateByAdmin(r.Context(), req)
	if err != nil {
		handleServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, u)
}

// PATCH /api/v1/users/{id} — обновление данных (менеджер/админ)
func (h *UserHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := pathParamInt64(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}
	var req dto.UpdateUserRequest
	if err := decode(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	u, err := h.userSvc.Update(r.Context(), id, req)
	if err != nil {
		handleServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, u)
}

// POST /api/v1/users/{id}/approve — одобрение заявки (менеджер/админ)
func (h *UserHandler) Approve(w http.ResponseWriter, r *http.Request) {
	id, err := pathParamInt64(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}
	var req dto.SetPasswordRequest
	if err := decode(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if err := h.userSvc.ApproveApplication(r.Context(), id, req); err != nil {
		handleServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, dto.SuccessResponse{Message: "Заявка одобрена"})
}

// POST /api/v1/users/{id}/reject — отклонение заявки (менеджер/админ)
func (h *UserHandler) Reject(w http.ResponseWriter, r *http.Request) {
	id, err := pathParamInt64(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}
	if err := h.userSvc.RejectApplication(r.Context(), id); err != nil {
		handleServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, dto.SuccessResponse{Message: "Заявка отклонена"})
}

// DELETE /api/v1/users/{id} — удаление пользователя (только админ)
func (h *UserHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := pathParamInt64(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}
	if err := h.userSvc.Delete(r.Context(), id); err != nil {
		handleServiceError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
