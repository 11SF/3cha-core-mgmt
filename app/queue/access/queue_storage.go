package access

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type QueueStatus string

const (
	StatusPending QueueStatus = "pending"
	StatusDone    QueueStatus = "done"
	StatusSkipped QueueStatus = "skipped"
)

type DailyQueue struct {
	ID        uuid.UUID   `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	QueueDate time.Time   `gorm:"type:date;index;not null" json:"queueDate"`
	MemberID  uuid.UUID   `gorm:"type:uuid;not null" json:"memberId"`
	Status    QueueStatus `gorm:"type:varchar(20);default:'pending'" json:"status"`
	CreatedAt time.Time   `json:"createdAt"`
	UpdatedAt time.Time   `json:"updatedAt"`
}

func (DailyQueue) TableName() string {
	return "daily_queue"
}

// QueueConfig is a singleton row that stores the round-robin reset marker.
type QueueConfig struct {
	ID           string     `gorm:"primaryKey"` // always "singleton"
	LastMemberID *uuid.UUID `gorm:"type:uuid"`  // member round-robin should resume AFTER
	UpdatedAt    time.Time
}

func (QueueConfig) TableName() string { return "queue_config" }

type QueueStorage interface {
	GetByDate(ctx context.Context, date time.Time) (DailyQueue, error)
	GetLast(ctx context.Context) (DailyQueue, error)
	List(ctx context.Context, from, to time.Time) ([]DailyQueue, error)
	Create(ctx context.Context, q DailyQueue) (DailyQueue, error)
	UpdateStatus(ctx context.Context, id uuid.UUID, status QueueStatus) (DailyQueue, error)
	DeleteByDate(ctx context.Context, date time.Time) error
	AutoClosePending(ctx context.Context, beforeDate time.Time) error
	GetConfig(ctx context.Context) (QueueConfig, error)
	UpsertConfig(ctx context.Context, cfg QueueConfig) error
}

type queueStorage struct {
	db *gorm.DB
}

var _ QueueStorage = (*queueStorage)(nil)

func NewQueueStorage(db *gorm.DB) QueueStorage {
	return &queueStorage{db: db}
}

func (s *queueStorage) GetByDate(ctx context.Context, date time.Time) (DailyQueue, error) {
	var q DailyQueue
	// Return the latest entry for the date — prefers pending, then any status.
	result := s.db.WithContext(ctx).
		Where("queue_date::date = ?", date.Format("2006-01-02")).
		Order("created_at DESC").
		First(&q)
	if result.Error != nil {
		return DailyQueue{}, result.Error
	}
	return q, nil
}

func (s *queueStorage) GetLast(ctx context.Context) (DailyQueue, error) {
	var q DailyQueue
	// Latest date wins; break ties by latest created_at (handles multiple entries per day).
	result := s.db.WithContext(ctx).
		Order("queue_date DESC, created_at DESC").
		First(&q)
	if result.Error != nil {
		return DailyQueue{}, result.Error
	}
	return q, nil
}

func (s *queueStorage) List(ctx context.Context, from, to time.Time) ([]DailyQueue, error) {
	var queues []DailyQueue
	result := s.db.WithContext(ctx).
		Where("queue_date BETWEEN ? AND ?", from.Format("2006-01-02"), to.Format("2006-01-02")).
		Order("queue_date DESC").
		Find(&queues)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to list queue: %w", result.Error)
	}
	return queues, nil
}

func (s *queueStorage) Create(ctx context.Context, q DailyQueue) (DailyQueue, error) {
	if q.ID == uuid.Nil {
		q.ID = uuid.New()
	}
	result := s.db.WithContext(ctx).Create(&q)
	if result.Error != nil {
		return DailyQueue{}, fmt.Errorf("failed to create queue: %w", result.Error)
	}
	return q, nil
}

func (s *queueStorage) DeleteByDate(ctx context.Context, date time.Time) error {
	return s.db.WithContext(ctx).
		Where("queue_date::date = ?", date.Format("2006-01-02")).
		Delete(&DailyQueue{}).Error
}

func (s *queueStorage) AutoClosePending(ctx context.Context, beforeDate time.Time) error {
	result := s.db.WithContext(ctx).
		Model(&DailyQueue{}).
		Where("status = ? AND queue_date::date < ?", StatusPending, beforeDate.Format("2006-01-02")).
		Update("status", StatusDone)
	if result.Error != nil {
		return fmt.Errorf("failed to auto-close pending entries: %w", result.Error)
	}
	return nil
}

func (s *queueStorage) GetConfig(ctx context.Context) (QueueConfig, error) {
	var cfg QueueConfig
	err := s.db.WithContext(ctx).First(&cfg, "id = ?", "singleton").Error
	if err != nil {
		return QueueConfig{}, err
	}
	return cfg, nil
}

func (s *queueStorage) UpsertConfig(ctx context.Context, cfg QueueConfig) error {
	cfg.ID = "singleton"
	return s.db.WithContext(ctx).Save(&cfg).Error
}

func (s *queueStorage) UpdateStatus(ctx context.Context, id uuid.UUID, status QueueStatus) (DailyQueue, error) {
	var q DailyQueue
	result := s.db.WithContext(ctx).
		Model(&q).
		Where("id = ?", id).
		Update("status", status)
	if result.Error != nil {
		return DailyQueue{}, fmt.Errorf("failed to update queue status: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return DailyQueue{}, gorm.ErrRecordNotFound
	}
	if err := s.db.WithContext(ctx).Where("id = ?", id).First(&q).Error; err != nil {
		return DailyQueue{}, fmt.Errorf("failed to fetch updated queue: %w", err)
	}
	return q, nil
}
