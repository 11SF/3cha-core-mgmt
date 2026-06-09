package holiday

import "portal/backend/app/holiday/access"

type HandlerConfig struct {
	HolidayStorage access.HolidayStorage
}

type handler struct {
	holidayStorage access.HolidayStorage
}

func NewHandler(cfg HandlerConfig) *handler {
	return &handler{holidayStorage: cfg.HolidayStorage}
}
