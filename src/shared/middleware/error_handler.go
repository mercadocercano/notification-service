package middleware

import (
	"net/http"
	"notification-service/src/shared/logger"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// ErrorResponse representa la estructura estándar de respuesta de error
type ErrorResponse struct {
	Success bool   `json:"success"`
	Error   Error  `json:"error"`
}

type Error struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}

// BusinessError representa errores de negocio que pueden ser manejados
type BusinessError struct {
	Code       string
	Message    string
	Details    string
	HTTPStatus int
}

func (e BusinessError) Error() string {
	return e.Message
}

// Errores de negocio predefinidos
var (
	ErrInvalidRequestFormat = BusinessError{
		Code:       "INVALID_REQUEST_FORMAT",
		Message:    "Formato de request inválido",
		HTTPStatus: http.StatusBadRequest,
	}
	
	ErrInvalidEmail = BusinessError{
		Code:       "INVALID_EMAIL",
		Message:    "El formato del email es inválido",
		HTTPStatus: http.StatusBadRequest,
	}
	
	ErrTemplateNotFound = BusinessError{
		Code:       "TEMPLATE_NOT_FOUND",
		Message:    "El template especificado no existe",
		HTTPStatus: http.StatusBadRequest,
	}
	
	ErrNotificationNotFound = BusinessError{
		Code:       "NOTIFICATION_NOT_FOUND",
		Message:    "La notificación no fue encontrada",
		HTTPStatus: http.StatusNotFound,
	}
	
	ErrInternalServer = BusinessError{
		Code:       "INTERNAL_SERVER_ERROR",
		Message:    "Error interno del servidor",
		HTTPStatus: http.StatusInternalServerError,
	}
)

// ErrorHandlerMiddleware maneja todos los errores de forma centralizada
func ErrorHandlerMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// Ejecutar el siguiente handler
		ctx.Next()

		// Si hay errores, manejarlos
		if len(ctx.Errors) > 0 {
			err := ctx.Errors.Last().Err
			handleError(ctx, err)
		}
	}
}

// handleError maneja diferentes tipos de errores
func handleError(ctx *gin.Context, err error) {
	log := logger.GetLogger()

	switch e := err.(type) {
	case BusinessError:
		// Error de negocio - respuesta controlada
		log.Warn("Business error occurred",
			zap.String("code", e.Code),
			zap.String("message", e.Message),
			zap.String("path", ctx.Request.URL.Path),
		)
		
		ctx.JSON(e.HTTPStatus, ErrorResponse{
			Success: false,
			Error: Error{
				Code:    e.Code,
				Message: e.Message,
				Details: e.Details,
			},
		})
		
	default:
		// Error no controlado - error interno
		log.Error("Unhandled error occurred",
			zap.Error(err),
			zap.String("path", ctx.Request.URL.Path),
			zap.String("method", ctx.Request.Method),
		)
		
		ctx.JSON(http.StatusInternalServerError, ErrorResponse{
			Success: false,
			Error: Error{
				Code:    "INTERNAL_SERVER_ERROR",
				Message: "Error interno del servidor",
				Details: err.Error(),
			},
		})
	}
}

// AbortWithBusinessError es un helper para abortar con error de negocio
func AbortWithBusinessError(ctx *gin.Context, err BusinessError) {
	ctx.Error(err)
	ctx.Abort()
}

// AbortWithError es un helper para abortar con error genérico
func AbortWithError(ctx *gin.Context, err error) {
	ctx.Error(err)
	ctx.Abort()
} 