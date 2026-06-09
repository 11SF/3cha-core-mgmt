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
