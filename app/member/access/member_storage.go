package access

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type MemberStorage interface {
	List(ctx context.Context) ([]Member, error)
	ListActive(ctx context.Context) ([]Member, error)
	GetByID(ctx context.Context, id uuid.UUID) (Member, error)
	Create(ctx context.Context, m Member) (Member, error)
	SoftDelete(ctx context.Context, id uuid.UUID) error
	BulkUpdateSortOrder(ctx context.Context, orders []SortOrderUpdate) error
}

type SortOrderUpdate struct {
	ID        uuid.UUID
	SortOrder int
}

type memberStorage struct {
	db *gorm.DB
}

var _ MemberStorage = (*memberStorage)(nil)

func NewMemberStorage(db *gorm.DB) MemberStorage {
	return &memberStorage{db: db}
}

type Member struct {
	ID          uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	Name        string     `gorm:"size:100;not null" json:"name"`
	AvatarColor string     `gorm:"size:7;default:'#6366f1'" json:"avatarColor"`
	IsActive    bool       `gorm:"default:true" json:"isActive"`
	SortOrder   int        `gorm:"default:0" json:"sortOrder"`
	CreatedAt   time.Time  `json:"createdAt"`
	UpdatedAt   time.Time  `json:"updatedAt"`
	DeletedAt   *time.Time `gorm:"index" json:"-"`
}

func (Member) TableName() string {
	return "members"
}

func (s *memberStorage) List(ctx context.Context) ([]Member, error) {
	var members []Member
	result := s.db.WithContext(ctx).
		Where("deleted_at IS NULL").
		Order("sort_order ASC, created_at ASC").
		Find(&members)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to list members: %w", result.Error)
	}
	return members, nil
}

func (s *memberStorage) ListActive(ctx context.Context) ([]Member, error) {
	var members []Member
	result := s.db.WithContext(ctx).
		Where("deleted_at IS NULL AND is_active = true").
		Order("sort_order ASC, created_at ASC").
		Find(&members)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to list active members: %w", result.Error)
	}
	return members, nil
}

func (s *memberStorage) GetByID(ctx context.Context, id uuid.UUID) (Member, error) {
	var m Member
	result := s.db.WithContext(ctx).
		Where("id = ? AND deleted_at IS NULL", id).
		First(&m)
	if result.Error != nil {
		return Member{}, fmt.Errorf("failed to get member: %w", result.Error)
	}
	return m, nil
}

func (s *memberStorage) Create(ctx context.Context, m Member) (Member, error) {
	if m.ID == uuid.Nil {
		m.ID = uuid.New()
	}
	result := s.db.WithContext(ctx).Create(&m)
	if result.Error != nil {
		return Member{}, fmt.Errorf("failed to create member: %w", result.Error)
	}
	return m, nil
}

func (s *memberStorage) BulkUpdateSortOrder(ctx context.Context, orders []SortOrderUpdate) error {
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, o := range orders {
			if err := tx.Model(&Member{}).
				Where("id = ? AND deleted_at IS NULL", o.ID).
				Update("sort_order", o.SortOrder).Error; err != nil {
				return fmt.Errorf("failed to update sort_order for %s: %w", o.ID, err)
			}
		}
		return nil
	})
}

func (s *memberStorage) SoftDelete(ctx context.Context, id uuid.UUID) error {
	now := time.Now()
	result := s.db.WithContext(ctx).
		Model(&Member{}).
		Where("id = ? AND deleted_at IS NULL", id).
		Update("deleted_at", now)
	if result.Error != nil {
		return fmt.Errorf("failed to delete member: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}
