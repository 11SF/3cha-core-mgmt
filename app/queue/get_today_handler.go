package queue

import (
	"errors"
	"log/slog"
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

	// Fast path: today already has a valid entry.
	if q, err := h.queueStorage.GetByDate(ctx, today); err == nil {
		member, err := h.memberStorage.GetByID(ctx, q.MemberID)
		if err != nil {
			response.InternalError(c, err)
			return
		}
		response.OK(c, toResponse(q, member.ID.String(), member.Name, member.AvatarColor))
		return
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		response.InternalError(c, err)
		return
	}

	members, err := h.memberStorage.ListActive(ctx)
	if err != nil {
		response.InternalError(c, err)
		return
	}
	if len(members) == 0 {
		response.Respond(c, response.Option[any]{
			HTTPStatus: http.StatusNotFound,
			Code:       response.CodeNotFound,
			Message:    "no active members — please add team members first",
		})
		return
	}

	lastQ, lastErr := h.queueStorage.GetLast(ctx)
	cfg, _ := h.queueStorage.GetConfig(ctx)

	var startDate time.Time
	var lastMemberID string

	if errors.Is(lastErr, gorm.ErrRecordNotFound) {
		// No history — start from today with config override (if any).
		startDate = today
		if cfg.LastMemberID != nil {
			lastMemberID = cfg.LastMemberID.String()
		}
	} else if lastErr != nil {
		response.InternalError(c, lastErr)
		return
	} else {
		if err := h.queueStorage.AutoClosePending(ctx, today); err != nil {
			slog.Warn("auto-close pending failed", "error", err)
		}
		startDate = lastQ.QueueDate.UTC().Truncate(24*time.Hour).AddDate(0, 0, 1)
		// Use config override when it was written after the last entry (i.e. a reset happened).
		if cfg.LastMemberID != nil && lastQ.CreatedAt.Before(cfg.UpdatedAt) {
			lastMemberID = cfg.LastMemberID.String()
		} else {
			lastMemberID = lastQ.MemberID.String()
		}
	}

	holidayList, err := h.holidayStorage.ListInRangeSnapshots(ctx, startDate, today)
	if err != nil {
		response.InternalError(c, err)
		return
	}
	holidays := holidaySetFrom(holidayList)

	var todayEntry *queueaccess.DailyQueue
	for d := startDate; !d.After(today); d = d.AddDate(0, 0, 1) {
		if !isWorkingDay(d, holidays) {
			continue
		}
		next := roundRobinNext(members, lastMemberID)
		q, err := h.queueStorage.Create(ctx, queueaccess.DailyQueue{
			QueueDate: d,
			MemberID:  next.ID,
			Status:    queueaccess.StatusPending,
		})
		if err != nil {
			slog.Error("backfill: failed to create entry", "date", d.Format("2006-01-02"), "error", err)
			response.InternalError(c, err)
			return
		}
		lastMemberID = next.ID.String()
		if d.Equal(today) {
			todayEntry = &q
		}
	}

	if todayEntry == nil {
		response.Respond(c, response.Option[any]{
			HTTPStatus: http.StatusOK,
			Code:       "NO_MEETING_TODAY",
			Message:    "วันนี้ไม่มี daily meeting (วันหยุด)",
		})
		return
	}

	member, err := h.memberStorage.GetByID(ctx, todayEntry.MemberID)
	if err != nil {
		response.InternalError(c, err)
		return
	}
	response.OK(c, toResponse(*todayEntry, member.ID.String(), member.Name, member.AvatarColor))
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
