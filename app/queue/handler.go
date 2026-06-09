package queue

import (
	"portal/backend/app/queue/access"
)

type HandlerConfig struct {
	MemberStorage access.MemberStorage
	QueueStorage  access.QueueStorage
}

type handler struct {
	memberStorage access.MemberStorage
	queueStorage  access.QueueStorage
}

func NewHandler(cfg HandlerConfig) *handler {
	return &handler{
		memberStorage: cfg.MemberStorage,
		queueStorage:  cfg.QueueStorage,
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
