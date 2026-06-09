package holiday

import (
	"portal/backend/pkg/response"

	"github.com/gin-gonic/gin"
)

func (h *handler) List(c *gin.Context) {
	list, err := h.holidayStorage.List(c.Request.Context())
	if err != nil {
		response.InternalError(c, err)
		return
	}
	response.OK(c, list)
}
