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
	List(ctx context.Context) ([]Holiday, error)
	ListInRange(ctx context.Context, from, to time.Time) ([]Holiday, error)
	Create(ctx context.Context, h Holiday) (Holiday, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

type holidayStorage struct {
	db *gorm.DB
}

var _ HolidayStorage = (*holidayStorage)(nil)

func NewHolidayStorage(db *gorm.DB) HolidayStorage {
	return &holidayStorage{db: db}
}

func (s *holidayStorage) List(ctx context.Context) ([]Holiday, error) {
	var list []Holiday
	if err := s.db.WithContext(ctx).Order("holiday_date DESC").Find(&list).Error; err != nil {
		return nil, fmt.Errorf("failed to list holidays: %w", err)
	}
	return list, nil
}

func (s *holidayStorage) ListInRange(ctx context.Context, from, to time.Time) ([]Holiday, error) {
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

func (s *holidayStorage) Create(ctx context.Context, h Holiday) (Holiday, error) {
	if h.ID == uuid.Nil {
		h.ID = uuid.New()
	}
	if err := s.db.WithContext(ctx).Create(&h).Error; err != nil {
		return Holiday{}, fmt.Errorf("failed to create holiday: %w", err)
	}
	return h, nil
}

func (s *holidayStorage) Delete(ctx context.Context, id uuid.UUID) error {
	result := s.db.WithContext(ctx).Where("id = ?", id).Delete(&Holiday{})
	if result.Error != nil {
		return fmt.Errorf("failed to delete holiday: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}
