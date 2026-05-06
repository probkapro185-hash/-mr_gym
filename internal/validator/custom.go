package validator

import (
	"fmt"
	"regexp"
	"strings"
)

var (
	// Российские номера: +7XXXXXXXXXX или 8XXXXXXXXXX, также форматы с пробелами/дефисами
	ruPhoneRe = regexp.MustCompile(`^(\+7|8)[\s\-]?\(?\d{3}\)?[\s\-]?\d{3}[\s\-]?\d{2}[\s\-]?\d{2}$`)
	emailRe   = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@(gmail\.com|mail\.ru|yandex\.ru|inbox\.ru|list\.ru|bk\.ru)$`)
)

type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

func (e ValidationError) Error() string {
	return fmt.Sprintf("%s: %s", e.Field, e.Message)
}

type ValidationErrors []ValidationError

func (ve ValidationErrors) Error() string {
	msgs := make([]string, len(ve))
	for i, e := range ve {
		msgs[i] = e.Error()
	}
	return strings.Join(msgs, "; ")
}

// ValidateApplicationRequest проверяет заявку от клиента
func ValidateApplicationRequest(fullName, phone, email string) ValidationErrors {
	var errs ValidationErrors

	if strings.TrimSpace(fullName) == "" {
		errs = append(errs, ValidationError{Field: "full_name", Message: "ФИО обязательно"})
	} else if len(strings.Fields(strings.TrimSpace(fullName))) < 2 {
		errs = append(errs, ValidationError{Field: "full_name", Message: "Укажите имя и фамилию"})
	}

	cleanPhone := strings.ReplaceAll(phone, " ", "")
	if !ruPhoneRe.MatchString(cleanPhone) {
		errs = append(errs, ValidationError{Field: "phone", Message: "Введите корректный российский номер телефона"})
	}

	if !emailRe.MatchString(strings.ToLower(strings.TrimSpace(email))) {
		errs = append(errs, ValidationError{Field: "email", Message: "Введите корректный email (gmail.com, mail.ru, yandex.ru и др.)"})
	}

	return errs
}

// ValidatePassword проверяет пароль
func ValidatePassword(password string) *ValidationError {
	if len(password) < 8 {
		return &ValidationError{Field: "password", Message: "Пароль должен содержать минимум 8 символов"}
	}
	return nil
}

// NormalizePhone приводит телефон к формату +7XXXXXXXXXX
func NormalizePhone(phone string) string {
	digits := regexp.MustCompile(`\D`).ReplaceAllString(phone, "")
	if len(digits) == 11 && digits[0] == '8' {
		digits = "7" + digits[1:]
	}
	if len(digits) == 11 {
		return "+" + digits
	}
	return phone
}
