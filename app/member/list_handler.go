package member

import (
	"portal/backend/pkg/response"

	"github.com/gin-gonic/gin"
)

func (h *handler) List(c *gin.Context) {
	ctx := c.Request.Context()

	members, err := h.memberStorage.List(ctx)
	if err != nil {
		response.InternalError(c, err)
		return
	}

	response.OK(c, members)
}
