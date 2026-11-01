package dto

import (
	"time"

	"github.com/google/uuid"
)

type PatientRegisterRequest struct {
	FirstName       string `json:"first_name" binding:"required" example:"Ali"`
	LastName        string `json:"last_name" binding:"required" example:"Valiyev"`
	Email           string `json:"email" binding:"required" example:"ali@example.com"`
	Phone           string `json:"phone" binding:"required" example:"+998901234567"`
	Password        string `json:"password" binding:"required" example:"StrongPass123!"`
	ConfirmPassword string `json:"confirm_password" binding:"required" example:"StrongPass123!"`
	AcceptTerms     bool   `json:"accept_terms" example:"true"`
}

type PatientRegisterData struct {
	ID        uuid.UUID `json:"id" example:"a1b2c3d4-e5f6-7890-abcd-ef1234567890"`
	FirstName string    `json:"first_name" example:"Ali"`
	LastName  string    `json:"last_name" example:"Valiyev"`
	Email     string    `json:"email" example:"ali@example.com"`
	Phone     string    `json:"phone" example:"+998901234567"`
	Role      string    `json:"role" example:"patient"`
	CreatedAt time.Time `json:"created_at" example:"2026-10-30T10:00:00Z"`
}

type VerifyEmailRequest struct {
	Email string `json:"email" binding:"required" example:"ali@example.com"`
	Code  string `json:"code" binding:"required" example:"123456"`
}

type ResendVerificationRequest struct {
	Email string `json:"email" binding:"required" example:"ali@example.com"`
}

type PatientLoginRequest struct {
	Email    string `json:"email" binding:"required" example:"ali@example.com"`
	Password string `json:"password" binding:"required" example:"StrongPass123!"`
}

type UserSummary struct {
	ID        uuid.UUID `json:"id" example:"a1b2c3d4-e5f6-7890-abcd-ef1234567890"`
	FirstName string    `json:"first_name" example:"Ali"`
	LastName  string    `json:"last_name" example:"Valiyev"`
	Email     string    `json:"email" example:"ali@example.com"`
	Phone     string    `json:"phone" example:"+998901234567"`
	Role      string    `json:"role" example:"patient"`
}

type LoginData struct {
	AccessToken string      `json:"access_token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
	TokenType   string      `json:"token_type" example:"Bearer"`
	ExpiresIn   int64       `json:"expires_in" example:"900"` // soniyada (15 daqiqa)
	User        UserSummary `json:"user"`
}

type ForgotPasswordRequest struct {
	Email string `json:"email" binding:"required" example:"ali@example.com"`
}

type ResetPasswordRequest struct {
	Email           string `json:"email" binding:"required" example:"ali@example.com"`
	Code            string `json:"code" binding:"required" example:"123456"`
	NewPassword     string `json:"new_password" binding:"required" example:"NewStrongPass123!"`
	ConfirmPassword string `json:"confirm_password" binding:"required" example:"NewStrongPass123!"`
}
