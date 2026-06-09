package access

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Holiday struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	HolidayDate time.Time `gorm:"type:date;uniqueIndex;not null" json:"holidayDate"`
	Name        string    `gorm:"size:200;not null" json:"name"`
	CreatedAt   time.Time `json:"createdAt"`
}

func (Holiday) TableName() string {
	return "holidays"
}

type HolidayStorage interface {
	ListSnapshots(ctx context.Context) ([]Holiday, error)
	ListInRangeSnapshots(ctx context.Context, from, to time.Time) ([]Holiday, error)
}

type holidayStorage struct {
	db *gorm.DB
}

var _ HolidayStorage = (*holidayStorage)(nil)

func NewHolidayStorage(db *gorm.DB) HolidayStorage {
	return &holidayStorage{db: db}
}

func (s *holidayStorage) ListSnapshots(ctx context.Context) ([]Holiday, error) {
	var list []Holiday
	if err := s.db.WithContext(ctx).Order("holiday_date DESC").Find(&list).Error; err != nil {
		return nil, fmt.Errorf("failed to list holidays: %w", err)
	}
	return list, nil
}

func (s *holidayStorage) ListInRangeSnapshots(ctx context.Context, from, to time.Time) ([]Holiday, error) {
	var list []Holiday
	err := s.db.WithContext(ctx).
		Where("holiday_date::date BETWEEN ? AND ?", from.Format("2006-01-02"), to.Format("2006-01-02")).
		Order("holiday_date ASC").
		Find(&list).Error
	if err != nil {
		return nil, fmt.Errorf("failed to list holidays in range: %w", err)
	}
	return list, nil
}
