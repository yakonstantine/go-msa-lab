package middleware

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	validator "github.com/go-playground/validator/v10"
	"github.com/yakonstantine/go-msa-lab/services/user-service/internal/dto"
	"github.com/yakonstantine/go-msa-lab/services/user-service/internal/entity"
)

func ErrorMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		if len(c.Errors) == 0 {
			return
		}

		err := c.Errors.Last().Err

		var valErr *entity.ValidationError
		var syntaxErr *json.SyntaxError
		var validatorErr validator.ValidationErrors
		switch {
		case errors.Is(err, entity.ErrNotFound):
			c.JSON(http.StatusNotFound, dto.ErrorDetailsDTO{
				Status: http.StatusNotFound,
				Error:  err.Error(),
			})
		case errors.Is(err, entity.ErrAlreadyExists):
			c.JSON(http.StatusConflict, dto.ErrorDetailsDTO{
				Status: http.StatusConflict,
				Error:  err.Error(),
			})
		case errors.As(err, &valErr):
			c.JSON(http.StatusBadRequest, dto.ErrorDetailsDTO{
				Status:  http.StatusBadRequest,
				Error:   valErr.Error(),
				Details: dto.FieldErrorsFromMap(valErr.Fields),
			})
		case errors.As(err, &syntaxErr):
			c.JSON(http.StatusBadRequest, dto.ErrorDetailsDTO{
				Status: http.StatusBadRequest,
				Error:  syntaxErr.Error(),
			})
		case errors.As(err, &validatorErr):
			c.JSON(http.StatusBadRequest, dto.ErrorDetailsDTO{
				Status:  http.StatusBadRequest,
				Error:   "invalid request body",
				Details: fieldErrorsFromValidationErrors(validatorErr),
			})
		default:
			slog.Error("unhandled error", "error", err)
			c.JSON(http.StatusInternalServerError, dto.ErrorDetailsDTO{
				Status: http.StatusInternalServerError,
				Error:  "internal server error",
			})
		}
	}
}

func fieldErrorsFromValidationErrors(fe []validator.FieldError) []dto.FieldErrorDTO {
	res := make([]dto.FieldErrorDTO, 0, len(fe))
	for _, err := range fe {
		res = append(res, dto.FieldErrorDTO{
			Field:   err.Field(),
			Message: fmt.Sprintf("field validation failed on the '%s' tag", err.Tag()),
		})
	}
	return res
}
