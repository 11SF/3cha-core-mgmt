package queue

import (
	"log/slog"
	"time"

	"portal/backend/pkg/response"

	"github.com/gin-gonic/gin"
)

func (h *handler) List(c *gin.Context) {
	ctx := c.Request.Context()

	to := time.Now().UTC().Truncate(24 * time.Hour)
	from := to.AddDate(0, 0, -30)

	queues, err := h.queueStorage.List(ctx, from, to)
	if err != nil {
		response.InternalError(c, err)
		return
	}

	type queueItem struct {
		ID          string `json:"id"`
		QueueDate   string `json:"queueDate"`
		Status      string `json:"status"`
		MemberID    string `json:"memberId"`
		MemberName  string `json:"memberName"`
		AvatarColor string `json:"avatarColor"`
	}

	items := make([]queueItem, 0, len(queues))
	for _, q := range queues {
		m, err := h.memberStorage.GetByID(ctx, q.MemberID)
		if err != nil {
			slog.Warn("member not found for queue entry", "memberId", q.MemberID, "queueId", q.ID)
			continue
		}
		items = append(items, queueItem{
			ID:          q.ID.String(),
			QueueDate:   q.QueueDate.Format("2006-01-02"),
			Status:      string(q.Status),
			MemberID:    m.ID.String(),
			MemberName:  m.Name,
			AvatarColor: m.AvatarColor,
		})
	}

	response.OK(c, items)
}
