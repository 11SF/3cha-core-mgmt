package response

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
)

const (
	CodeSuccess       = "SUCCESS"
	CodeBadRequest    = "BAD_REQUEST"
	CodeNotFound      = "NOT_FOUND"
	CodeConflict      = "CONFLICT"
	CodeInternalError = "INTERNAL_ERROR"

	MessageSuccess       = "success"
	MessageBadRequest    = "bad request"
	MessageNotFound      = "not found"
	MessageConflict      = "conflict"
	MessageInternalError = "internal server error"
)

type body[T any] struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Data    *T     `json:"data,omitempty"`
}

type Option[T any] struct {
	HTTPStatus int
	Code       string
	Message    string
	Data       *T
	Err        error
}

func Respond[T any](c *gin.Context, opt Option[T]) {
	if opt.Err != nil {
		slog.Error("handler error", "error", opt.Err)
	}
	c.JSON(opt.HTTPStatus, body[T]{
		Code:    opt.Code,
		Message: opt.Message,
		Data:    opt.Data,
	})
}

func OK[T any](c *gin.Context, data T) {
	Respond(c, Option[T]{
		HTTPStatus: http.StatusOK,
		Code:       CodeSuccess,
		Message:    MessageSuccess,
		Data:       &data,
	})
}

func BadRequest(c *gin.Context, err error) {
	Respond(c, Option[any]{
		HTTPStatus: http.StatusBadRequest,
		Code:       CodeBadRequest,
		Message:    MessageBadRequest,
		Err:        err,
	})
}

func NotFound(c *gin.Context) {
	Respond(c, Option[any]{
		HTTPStatus: http.StatusNotFound,
		Code:       CodeNotFound,
		Message:    MessageNotFound,
	})
}

func InternalError(c *gin.Context, err error) {
	Respond(c, Option[any]{
		HTTPStatus: http.StatusInternalServerError,
		Code:       CodeInternalError,
		Message:    MessageInternalError,
		Err:        err,
	})
}
