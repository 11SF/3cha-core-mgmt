package queue

import (
	"errors"
	"net/http"

	queueaccess "portal/backend/app/queue/access"
	"portal/backend/pkg/response"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

func (h *handler) MarkDone(c *gin.Context) {
	ctx := c.Request.Context()

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, err)
		return
	}

	q, err := h.queueStorage.UpdateStatus(ctx, id, queueaccess.StatusDone)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.NotFound(c)
			return
		}
		response.InternalError(c, err)
		return
	}

	member, err := h.memberStorage.GetByID(ctx, q.MemberID)
	if err != nil {
		response.InternalError(c, err)
		return
	}

	response.Respond(c, response.Option[QueueWithMember]{
		HTTPStatus: http.StatusOK,
		Code:       response.CodeSuccess,
		Message:    response.MessageSuccess,
		Data: &QueueWithMember{
			ID:          q.ID.String(),
			QueueDate:   q.QueueDate.Format("2006-01-02"),
			Status:      q.Status,
			MemberID:    member.ID.String(),
			MemberName:  member.Name,
			AvatarColor: member.AvatarColor,
		},
	})
}
