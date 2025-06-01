package response

import (
	"net/http"
	"time"
	"notification-service/src/shared/middleware"
)

type SendNotificationResponse struct {
	ID        string    `json:"id"`
	Message   string    `json:"message"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}

// Result encapsula el resultado de una operación con manejo de errores de negocio
type SendNotificationResult struct {
	Data       *SendNotificationResponse `json:"data,omitempty"`
	Error      *BusinessError            `json:"error,omitempty"`
	Success    bool                      `json:"success"`
	HTTPStatus int                       `json:"-"` // No se serializa en JSON
}

// BusinessError representa errores de negocio con códigos específicos
type BusinessError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}

// Métodos factory para crear resultados
func NewSendNotificationSuccess(data *SendNotificationResponse) *SendNotificationResult {
	return &SendNotificationResult{
		Data:       data,
		Success:    true,
		HTTPStatus: http.StatusOK,
	}
}

func NewSendNotificationError(code, message, details string, httpStatus int) *SendNotificationResult {
	return &SendNotificationResult{
		Error: &BusinessError{
			Code:    code,
			Message: message,
			Details: details,
		},
		Success:    false,
		HTTPStatus: httpStatus,
	}
}

// Métodos de conveniencia para errores comunes
func NewInvalidEmailError() *SendNotificationResult {
	return NewSendNotificationError(
		"INVALID_EMAIL",
		"El formato del email es inválido",
		"",
		http.StatusBadRequest,
	)
}

func NewTemplateNotFoundError() *SendNotificationResult {
	return NewSendNotificationError(
		"TEMPLATE_NOT_FOUND",
		"El template especificado no existe",
		"",
		http.StatusBadRequest,
	)
}

func NewInvalidNotificationTypeError() *SendNotificationResult {
	return NewSendNotificationError(
		"INVALID_NOTIFICATION_TYPE",
		"El tipo de notificación no es válido",
		"",
		http.StatusBadRequest,
	)
}

func NewInternalServerError(details string) *SendNotificationResult {
	return NewSendNotificationError(
		"INTERNAL_SERVER_ERROR",
		"Error interno del servidor",
		details,
		http.StatusInternalServerError,
	)
}

// ToMiddlewareError convierte el BusinessError del resultado a BusinessError del middleware
func (r *SendNotificationResult) ToMiddlewareError() middleware.BusinessError {
	if r.Success || r.Error == nil {
		return middleware.BusinessError{}
	}
	
	return middleware.BusinessError{
		Code:       r.Error.Code,
		Message:    r.Error.Message,
		Details:    r.Error.Details,
		HTTPStatus: r.HTTPStatus,
	}
} 