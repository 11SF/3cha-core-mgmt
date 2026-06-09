package holiday

import (
	"net/http"
	"time"

	"portal/backend/app/holiday/access"
	"portal/backend/pkg/response"

	"github.com/gin-gonic/gin"
)

type CreateHolidayRequest struct {
	HolidayDate string `json:"holidayDate" binding:"required"` // YYYY-MM-DD
	Name        string `json:"name" binding:"required,min=1,max=200"`
}

func (h *handler) Create(c *gin.Context) {
	var req CreateHolidayRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err)
		return
	}

	date, err := time.Parse("2006-01-02", req.HolidayDate)
	if err != nil {
		response.BadRequest(c, err)
		return
	}

	holiday, err := h.holidayStorage.Create(c.Request.Context(), access.Holiday{
		HolidayDate: date,
		Name:        req.Name,
	})
	if err != nil {
		response.InternalError(c, err)
		return
	}

	response.Respond(c, response.Option[access.Holiday]{
		HTTPStatus: http.StatusCreated,
		Code:       response.CodeSuccess,
		Message:    response.MessageSuccess,
		Data:       &holiday,
	})
}
