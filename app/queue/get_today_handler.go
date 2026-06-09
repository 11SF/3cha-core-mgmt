package queue

import (
	"errors"
	"net/http"
	"time"

	queueaccess "portal/backend/app/queue/access"
	"portal/backend/pkg/response"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type QueueWithMember struct {
	ID          string                  `json:"id"`
	QueueDate   string                  `json:"queueDate"`
	Status      queueaccess.QueueStatus `json:"status"`
	MemberID    string                  `json:"memberId"`
	MemberName  string                  `json:"memberName"`
	AvatarColor string                  `json:"avatarColor"`
}

func (h *handler) GetToday(c *gin.Context) {
	ctx := c.Request.Context()
	today := time.Now().UTC().Truncate(24 * time.Hour)

	q, err := h.queueStorage.GetByDate(ctx, today)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		response.Respond(c, response.Option[any]{
			HTTPStatus: http.StatusOK,
			Code:       "NO_MEETING_TODAY",
			Message:    "วันนี้ไม่มี daily meeting (วันหยุด)",
		})
		return
	}
	if err != nil {
		response.InternalError(c, err)
		return
	}

	member, err := h.memberStorage.GetByID(ctx, q.MemberID)
	if err != nil {
		response.InternalError(c, err)
		return
	}

	response.OK(c, toResponse(q, member.ID.String(), member.Name, member.AvatarColor))
}

func toResponse(q queueaccess.DailyQueue, memberID, name, color string) QueueWithMember {
	return QueueWithMember{
		ID:          q.ID.String(),
		QueueDate:   q.QueueDate.Format("2006-01-02"),
		Status:      q.Status,
		MemberID:    memberID,
		MemberName:  name,
		AvatarColor: color,
	}
}
