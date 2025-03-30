package storage

import (
	"context"

	"github.com/EbumbaE/bandit/pkg/redis"

	model "github.com/EbumbaE/bandit/services/rule-diller/internal"
)

type Storage struct {
	conn redis.Client
}

func NewStorage(conn redis.Client) *Storage {
	return &Storage{
		conn: conn,
	}
}

func (s *Storage) SaveRuleVariants(ctx context.Context, key string, variants []model.Variant) error {
	return nil
}

func (s *Storage) SaveRuleVersion(ctx context.Context, key string, version uint64) error {
	return nil
}

func (s *Storage) SaveVariantData(ctx context.Context, key string, data []byte) error {
	return nil
}

func (s *Storage) SetVariantCount(ctx context.Context, key string, count uint64) error {
	return nil
}

func (s *Storage) GetRuleVariants(ctx context.Context, key string) ([]model.Variant, error) {
	return nil, nil
}

func (s *Storage) GetRuleVersion(ctx context.Context, key string) (uint64, error) {
	return 0, nil
}

func (s *Storage) GetVariantData(ctx context.Context, key string) ([]byte, error) {
	return nil, nil
}

func (s *Storage) GetVariantCount(ctx context.Context, key string) (uint64, error) {
	return 0, nil
}

func (s *Storage) IncVariantCount(ctx context.Context, key string) error {
	return nil
}
