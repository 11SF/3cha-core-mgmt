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

type NextEntry struct {
	MemberName  string `json:"memberName"`
	AvatarColor string `json:"avatarColor"`
	QueueDate   string `json:"queueDate"`
}

type QueueWithMember struct {
	ID            string                  `json:"id"`
	QueueDate     string                  `json:"queueDate"`
	Status        queueaccess.QueueStatus `json:"status"`
	MemberID      string                  `json:"memberId"`
	MemberName    string                  `json:"memberName"`
	AvatarColor   string                  `json:"avatarColor"`
	ConfluenceUrl string                  `json:"confluenceUrl,omitempty"`
	Position      int                     `json:"position"`
	TotalMembers  int                     `json:"totalMembers"`
	Next          *NextEntry              `json:"next,omitempty"`
}

func (h *handler) GetToday(c *gin.Context) {
	ctx := c.Request.Context()
	loc, _ := time.LoadLocation(h.cfg.Database.TimeZone)
	now := time.Now().In(loc)
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)

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

	activeMembers, err := h.memberStorage.ListActive(ctx)
	if err != nil {
		response.InternalError(c, err)
		return
	}

	holidays, err := h.holidayStorage.ListInRange(ctx, today.AddDate(0, 0, 1), today.AddDate(0, 3, 0))
	if err != nil {
		response.InternalError(c, err)
		return
	}

	position := 0
	for i, m := range activeMembers {
		if m.ID == q.MemberID {
			position = i + 1
			break
		}
	}

	resp := toResponse(q, member.ID.String(), member.Name, member.AvatarColor, h.cfg.ConfluenceUrl)
	resp.Position = position
	resp.TotalMembers = len(activeMembers)

	if nextDay := nextWorkingDay(today, holidays); nextDay != nil && len(activeMembers) > 0 {
		nextMem := roundRobinNext(activeMembers, q.MemberID.String())
		resp.Next = &NextEntry{
			MemberName:  nextMem.Name,
			AvatarColor: nextMem.AvatarColor,
			QueueDate:   nextDay.Format("2006-01-02"),
		}
	}

	response.OK(c, resp)
}

func toResponse(q queueaccess.DailyQueue, memberID, name, color, confluenceUrl string) QueueWithMember {
	return QueueWithMember{
		ID:            q.ID.String(),
		QueueDate:     q.QueueDate.Format("2006-01-02"),
		Status:        q.Status,
		MemberID:      memberID,
		MemberName:    name,
		AvatarColor:   color,
		ConfluenceUrl: confluenceUrl,
	}
}

func nextWorkingDay(from time.Time, holidays []queueaccess.Holiday) *time.Time {
	set := make(map[string]bool, len(holidays))
	for _, h := range holidays {
		set[h.HolidayDate.Format("2006-01-02")] = true
	}
	d := from
	for i := 0; i < 60; i++ {
		d = d.AddDate(0, 0, 1)
		if d.Weekday() != time.Saturday && d.Weekday() != time.Sunday && !set[d.Format("2006-01-02")] {
			return &d
		}
	}
	return nil
}
