package access

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"
)

type Holiday struct {
	HolidayDate time.Time `gorm:"column:holiday_date;type:date"`
}

func (Holiday) TableName() string { return "holidays" }

type HolidayStorage interface {
	ListInRange(ctx context.Context, from, to time.Time) ([]Holiday, error)
}

type holidayStorage struct {
	db *gorm.DB
}

var _ HolidayStorage = (*holidayStorage)(nil)

func NewHolidayStorage(db *gorm.DB) HolidayStorage {
	return &holidayStorage{db: db}
}

func (s *holidayStorage) ListInRange(ctx context.Context, from, to time.Time) ([]Holiday, error) {
	var list []Holiday
	err := s.db.WithContext(ctx).
		Select("holiday_date").
		Where("holiday_date::date BETWEEN ? AND ?", from.Format("2006-01-02"), to.Format("2006-01-02")).
		Order("holiday_date ASC").
		Find(&list).Error
	if err != nil {
		return nil, fmt.Errorf("failed to list holidays in range: %w", err)
	}
	return list, nil
}
