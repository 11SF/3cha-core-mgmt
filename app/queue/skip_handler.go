package queue

import (
	"errors"
	"log/slog"

	queueaccess "portal/backend/app/queue/access"
	"portal/backend/pkg/response"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Skip marks the current entry as skipped and immediately assigns the next
// person in round-robin to the same date so the meeting can proceed.
func (h *handler) Skip(c *gin.Context) {
	ctx := c.Request.Context()

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, err)
		return
	}

	// 1. Mark this entry as skipped.
	skipped, err := h.queueStorage.UpdateStatus(ctx, id, queueaccess.StatusSkipped)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.NotFound(c)
			return
		}
		response.InternalError(c, err)
		return
	}

	// 2. Find the next active member after the skipped one.
	members, err := h.memberStorage.ListActive(ctx)
	if err != nil || len(members) == 0 {
		response.InternalError(c, err)
		return
	}
	next := roundRobinNext(members, skipped.MemberID.String())

	// 3. Create a new entry for the same date with the next person.
	newEntry, err := h.queueStorage.Create(ctx, queueaccess.DailyQueue{
		QueueDate: skipped.QueueDate,
		MemberID:  next.ID,
		Status:    queueaccess.StatusPending,
	})
	if err != nil {
		slog.Error("skip: failed to create replacement entry", "error", err)
		response.InternalError(c, err)
		return
	}

	response.OK(c, toResponse(newEntry, next.ID.String(), next.Name, next.AvatarColor, ""))
}
