package handler

import (
	"net/http"
	"time"

	"crm_gym/internal/dto"
	"crm_gym/internal/middleware"
	"crm_gym/internal/service"
)

type SessionHandler struct {
	sessionSvc *service.SessionService
}

func NewSessionHandler(sessionSvc *service.SessionService) *SessionHandler {
	return &SessionHandler{sessionSvc: sessionSvc}
}

// GET /api/v1/sessions/my — занятия текущего клиента
func (h *SessionHandler) GetMySessions(w http.ResponseWriter, r *http.Request) {
	userID, _ := middleware.GetUserID(r.Context())
	from, to := parseDateRange(r)

	sessions, err := h.sessionSvc.GetForClient(r.Context(), userID, from, to)
	if err != nil {
		handleServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, dto.ListResponse{Items: sessions, Total: len(sessions)})
}

// GET /api/v1/sessions — все занятия (менеджер/админ)
func (h *SessionHandler) GetAll(w http.ResponseWriter, r *http.Request) {
	from, to := parseDateRange(r)
	sessions, err := h.sessionSvc.GetAll(r.Context(), from, to)
	if err != nil {
		handleServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, dto.ListResponse{Items: sessions, Total: len(sessions)})
}

// GET /api/v1/sessions/client/{id} — занятия конкретного клиента (менеджер/админ)
func (h *SessionHandler) GetByClientID(w http.ResponseWriter, r *http.Request) {
	clientID, err := pathParamInt64(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}
	from, to := parseDateRange(r)
	sessions, err := h.sessionSvc.GetForClient(r.Context(), clientID, from, to)
	if err != nil {
		handleServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, dto.ListResponse{Items: sessions, Total: len(sessions)})
}

// POST /api/v1/sessions — создание занятия (менеджер/админ)
func (h *SessionHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req dto.CreateSessionRequest
	if err := decode(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	session, err := h.sessionSvc.Create(r.Context(), req)
	if err != nil {
		handleServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, session)
}

// PATCH /api/v1/sessions/{id} — обновление занятия (время, дата, тренер) — менеджер/админ
func (h *SessionHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := pathParamInt64(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}
	var req dto.UpdateSessionRequest
	if err := decode(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	session, err := h.sessionSvc.Update(r.Context(), id, req)
	if err != nil {
		handleServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, session)
}

// DELETE /api/v1/sessions/{id} — удаление занятия (менеджер/админ)
func (h *SessionHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := pathParamInt64(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}
	if err := h.sessionSvc.Delete(r.Context(), id); err != nil {
		handleServiceError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// parseDateRange извлекает from/to из query параметров (формат RFC3339 или date YYYY-MM-DD)
// По умолчанию — текущий месяц
func parseDateRange(r *http.Request) (time.Time, time.Time) {
	now := time.Now()
	fromStr := queryString(r, "from")
	toStr := queryString(r, "to")

	parseDate := func(s string, fallback time.Time) time.Time {
		if s == "" {
			return fallback
		}
		if t, err := time.Parse(time.RFC3339, s); err == nil {
			return t
		}
		if t, err := time.Parse("2006-01-02", s); err == nil {
			return t
		}
		return fallback
	}

	from := parseDate(fromStr, time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location()))
	to := parseDate(toStr, from.AddDate(0, 1, 0))
	return from, to
}
