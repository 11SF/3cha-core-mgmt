package member

import (
	"errors"
	"net/http"

	"portal/backend/pkg/response"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

func (h *handler) Delete(c *gin.Context) {
	ctx := c.Request.Context()

	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.BadRequest(c, err)
		return
	}

	if err := h.memberStorage.SoftDelete(ctx, id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.NotFound(c)
			return
		}
		response.InternalError(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}
