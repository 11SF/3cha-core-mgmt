package member

import (
	"portal/backend/app/member/access"
)

type HandlerConfig struct {
	MemberStorage access.MemberStorage
}

type handler struct {
	memberStorage access.MemberStorage
}

func NewHandler(cfg HandlerConfig) *handler {
	return &handler{
		memberStorage: cfg.MemberStorage,
	}
}
