package queue

import (
	"time"

	"portal/backend/app/queue/access"
)

type HandlerConfig struct {
	MemberStorage  access.MemberStorage
	QueueStorage   access.QueueStorage
	HolidayStorage access.HolidayStorage
}

type handler struct {
	memberStorage  access.MemberStorage
	queueStorage   access.QueueStorage
	holidayStorage access.HolidayStorage
}

func NewHandler(cfg HandlerConfig) *handler {
	return &handler{
		memberStorage:  cfg.MemberStorage,
		queueStorage:   cfg.QueueStorage,
		holidayStorage: cfg.HolidayStorage,
	}
}

// roundRobinNext returns the next member after lastMemberID.
// Falls back to first member if lastMemberID is not found.
func roundRobinNext(members []access.Member, lastMemberID string) access.Member {
	if len(members) == 0 {
		return access.Member{}
	}
	for i, m := range members {
		if m.ID.String() == lastMemberID {
			return members[(i+1)%len(members)]
		}
	}
	return members[0]
}

// isWorkingDay returns true if the date is Mon–Fri and not in the holiday set.
func isWorkingDay(date time.Time, holidaySet map[string]bool) bool {
	wd := date.Weekday()
	if wd == time.Saturday || wd == time.Sunday {
		return false
	}
	return !holidaySet[date.Format("2006-01-02")]
}

// holidaySetFrom converts a slice of holidays to a lookup map.
func holidaySetFrom(holidays []access.Holiday) map[string]bool {
	set := make(map[string]bool, len(holidays))
	for _, h := range holidays {
		set[h.HolidayDate.Format("2006-01-02")] = true
	}
	return set
}
