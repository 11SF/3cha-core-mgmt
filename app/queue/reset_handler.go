package queue

import (
	"log/slog"
	"net/http"
	"time"

	queueaccess "portal/backend/app/queue/access"
	"portal/backend/pkg/response"

	"github.com/gin-gonic/gin"
)

func (h *handler) Reset(c *gin.Context) {
	ctx := c.Request.Context()

	members, err := h.memberStorage.ListActive(ctx)
	if err != nil {
		response.InternalError(c, err)
		return
	}
	if len(members) == 0 {
		response.BadRequest(c, nil)
		return
	}

	// Delete today's entry so GetToday regenerates it fresh from members[0].
	today := time.Now().UTC().Truncate(24 * time.Hour)
	if err := h.queueStorage.DeleteByDate(ctx, today); err != nil {
		slog.Error("failed to delete today's queue entry on reset", "error", err)
		response.InternalError(c, err)
		return
	}

	// Point to the last member so roundRobinNext returns members[0].
	lastMember := members[len(members)-1]
	if err := h.queueStorage.UpsertConfig(ctx, queueaccess.QueueConfig{
		LastMemberID: &lastMember.ID,
	}); err != nil {
		slog.Error("failed to reset queue config", "error", err)
		response.InternalError(c, err)
		return
	}

	slog.Info("queue reset", "nextFirst", members[0].Name)
	response.Respond(c, response.Option[any]{
		HTTPStatus: http.StatusOK,
		Code:       response.CodeSuccess,
		Message:    "queue reset — " + members[0].Name + " จะเป็นคนแรกในรอบถัดไป",
	})
}
