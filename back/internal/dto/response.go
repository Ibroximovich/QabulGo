package dto

type SuccessResponse struct {
	Success bool        `json:"success" example:"true"`
	Message string      `json:"message" example:"Amal muvaffaqiyatli bajarildi"`
	Data    interface{} `json:"data,omitempty"`
}

type ErrorResponse struct {
	Success bool              `json:"success" example:"false"`
	Message string            `json:"message" example:"Kiritilgan ma'lumotlarda xatolik mavjud"`
	Errors  map[string]string `json:"errors,omitempty"`
}

func NewSuccessResponse(message string, data interface{}) SuccessResponse {
	return SuccessResponse{
		Success: true,
		Message: message,
		Data:    data,
	}
}

func NewErrorResponse(message string, errors map[string]string) ErrorResponse {
	return ErrorResponse{
		Success: false,
		Message: message,
		Errors:  errors,
	}
}
