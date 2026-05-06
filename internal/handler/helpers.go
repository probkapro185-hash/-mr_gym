package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"crm_gym/internal/domain"
	"crm_gym/internal/dto"
	"crm_gym/internal/validator"
)

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string, details ...interface{}) {
	resp := dto.ErrorResponse{Error: msg}
	if len(details) > 0 {
		resp.Details = details[0]
	}
	writeJSON(w, status, resp)
}

func decode(r *http.Request, v interface{}) error {
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	return dec.Decode(v)
}

func handleServiceError(w http.ResponseWriter, err error) {
	var valErrs validator.ValidationErrors
	if errors.As(err, &valErrs) {
		writeError(w, http.StatusBadRequest, "validation error", valErrs)
		return
	}
	switch {
	case errors.Is(err, domain.ErrNotFound):
		writeError(w, http.StatusNotFound, "not found")
	case errors.Is(err, domain.ErrAlreadyExists):
		writeError(w, http.StatusConflict, "already exists")
	case errors.Is(err, domain.ErrInvalidCredentials):
		writeError(w, http.StatusUnauthorized, "invalid credentials")
	case errors.Is(err, domain.ErrForbidden):
		writeError(w, http.StatusForbidden, "forbidden")
	case errors.Is(err, domain.ErrInsufficientBalance):
		writeError(w, http.StatusPaymentRequired, "insufficient balance")
	default:
		writeError(w, http.StatusInternalServerError, err.Error())
	}
}

func pathParamInt64(r *http.Request, key string) (int64, error) {
	val := r.PathValue(key)
	return strconv.ParseInt(val, 10, 64)
}

func queryInt(r *http.Request, key string, def int) int {
	v := r.URL.Query().Get(key)
	if v == "" {
		return def
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return def
	}
	return n
}

func queryString(r *http.Request, key string) string {
	return r.URL.Query().Get(key)
}
