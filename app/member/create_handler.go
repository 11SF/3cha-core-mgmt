package member

import (
	"net/http"

	"portal/backend/app/member/access"
	"portal/backend/pkg/response"

	"github.com/gin-gonic/gin"
)

type CreateMemberRequest struct {
	Name        string `json:"name" binding:"required,min=1,max=100"`
	AvatarColor string `json:"avatarColor" binding:"omitempty,len=7"`
}

func (h *handler) Create(c *gin.Context) {
	ctx := c.Request.Context()

	var req CreateMemberRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err)
		return
	}

	color := req.AvatarColor
	if color == "" {
		color = "#6366f1"
	}

	m, err := h.memberStorage.Create(ctx, access.Member{
		Name:        req.Name,
		AvatarColor: color,
	})
	if err != nil {
		response.InternalError(c, err)
		return
	}

	response.Respond(c, response.Option[access.Member]{
		HTTPStatus: http.StatusCreated,
		Code:       response.CodeSuccess,
		Message:    response.MessageSuccess,
		Data:       &m,
	})
}
