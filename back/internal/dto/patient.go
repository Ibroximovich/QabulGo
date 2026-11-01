package dto

import (
	"time"

	"github.com/google/uuid"
)

type PatientProfileResponse struct {
	ID            uuid.UUID `json:"id" example:"a1b2c3d4-e5f6-7890-abcd-ef1234567890"`
	FirstName     string    `json:"first_name" example:"Ali"`
	LastName      string    `json:"last_name" example:"Valiyev"`
	Email         string    `json:"email" example:"ali@example.com"`
	Phone         string    `json:"phone" example:"+998901234567"`
	Role          string    `json:"role" example:"patient"`
	EmailVerified bool      `json:"email_verified" example:"true"`
	CreatedAt     time.Time `json:"created_at" example:"2026-10-30T10:00:00Z"`
}

type UpdatePatientProfileRequest struct {
	FirstName string `json:"first_name" binding:"required" example:"Ali"`
	LastName  string `json:"last_name" binding:"required" example:"Valiyev"`
	Phone     string `json:"phone" binding:"required" example:"+998901234567"`
}
