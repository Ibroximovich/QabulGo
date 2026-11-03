package utils

import (
	"errors"
	"net/mail"
	"regexp"
	"strings"
)

var (
	phoneRegex    = regexp.MustCompile(`^\+998[0-9]{9}$`)
	weakPasswords = map[string]struct{}{
		"12345678":      {},
		"123456789":     {},
		"1234567890":    {},
		"password":      {},
		"password123":   {},
		"qwertyui":      {},
		"qwerty123":     {},
		"admin123":      {},
		"administrator": {},
		"iloveyou":      {},
		"letmein123":    {},
		"welcome123":    {},
	}
)

// ValidateUzbekistanPhone checks if phone number follows +998XXXXXXXXX format
func ValidateUzbekistanPhone(phone string) bool {
	return phoneRegex.MatchString(phone)
}

// ValidateEmail checks if email has a valid syntax
func ValidateEmail(email string) bool {
	if len(email) < 5 || len(email) > 255 {
		return false
	}
	addr, err := mail.ParseAddress(email)
	if err != nil {
		return false
	}
	return addr.Address == email && strings.Contains(email, "@") && strings.Contains(email, ".")
}

// NormalizeEmail converts email to lowercase and trims whitespace
func NormalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

// ValidatePassword ensures password is 8-128 chars and not in the weak password list
func ValidatePassword(password string) error {
	if len(password) < 8 {
		return errors.New("parol kamida 8 ta belgidan iborat bo‘lishi kerak")
	}
	if len(password) > 128 {
		return errors.New("parol ko‘pi bilan 128 ta belgidan iborat bo‘lishi kerak")
	}
	if _, isWeak := weakPasswords[strings.ToLower(password)]; isWeak {
		return errors.New("kiritilgan parol juda zaif, iltimos murakkabroq parol tanlang")
	}
	return nil
}

// SanitizeString trims whitespace
func SanitizeString(s string) string {
	return strings.TrimSpace(s)
}

// MaskEmail masks email address for secure logging (e.g. al***@gmail.com)
func MaskEmail(email string) string {
	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		return "***"
	}
	name := parts[0]
	domain := parts[1]
	if len(name) <= 2 {
		return "***@" + domain
	}
	return name[:2] + "***@" + domain
}
