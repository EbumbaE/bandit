package storage

import (
	"context"
	"fmt"
	"strconv"

	redis_client "github.com/EbumbaE/bandit/pkg/redis"
	"github.com/pkg/errors"
	"github.com/redis/go-redis/v9"

	model "github.com/EbumbaE/bandit/services/rule-diller/internal"
)

type Storage struct {
	conn *redis_client.Client
}

func NewStorage(conn *redis_client.Client) *Storage {
	return &Storage{
		conn: conn,
	}
}

func keyRuleVersion(service, context string) string {
	return fmt.Sprintf("rule:%s:%s:version", service, context)
}

func keyRuleVariants(service, context string) string {
	return fmt.Sprintf("rule:%s:%s:variants", service, context)
}

func keyVariantData(service, context, variantID string) string {
	return fmt.Sprintf("variant:%s:%s:%s", service, context, variantID)
}

func (s *Storage) SaveRuleVariants(ctx context.Context, service, context, ruleID string, variants []model.Variant) error {
	pipe := s.conn.TxPipeline()

	oldVariants, err := s.conn.ZRange(ctx, keyRuleVariants(service, context), 0, -1).Result()
	if err != nil && !errors.Is(err, redis.Nil) {
		return errors.Wrap(err, "get current variants")
	}

	newIDs := make(map[string]struct{}, len(variants))
	for _, v := range variants {
		newIDs[v.Key] = struct{}{}
	}

	toDelete := make([]string, 0)
	for _, oldID := range oldVariants {
		if _, exists := newIDs[oldID]; !exists {
			toDelete = append(toDelete, oldID)
		}
	}

	if len(toDelete) > 0 {
		pipe.ZRem(ctx, keyRuleVariants(service, context), toDelete)
		for _, id := range toDelete {
			pipe.Del(ctx, keyVariantData(service, context, id))
		}
	}

	for _, v := range variants {
		pipe.ZAdd(ctx, keyRuleVariants(service, context), redis.Z{
			Score:  v.Score,
			Member: v.Key,
		})

		pipe.HSet(ctx, keyVariantData(service, context, v.Key), "data", v.Data, "count", v.Count, "rule_id", ruleID)
	}

	_, err = pipe.Exec(ctx)
	return errors.Wrap(err, "save rule variants")
}

func (s *Storage) SaveRuleVersion(ctx context.Context, service, context string, version uint64) error {
	return s.conn.Set(ctx, keyRuleVersion(service, context), version, 0).Err()
}

func (s *Storage) GetRuleVariants(ctx context.Context, service, context string, withData bool) ([]model.Variant, error) {
	variantIDs, err := s.conn.ZRange(ctx, keyRuleVariants(service, context), 0, -1).Result()
	if err != nil {
		return nil, err
	}

	variants := make([]model.Variant, 0, len(variantIDs))
	pipe := s.conn.TxPipeline()

	cmds := make(map[string]any, len(variantIDs))
	for _, variantID := range variantIDs {
		dataKey := keyVariantData(service, context, variantID)
		if withData {
			cmds[variantID] = pipe.HGetAll(ctx, dataKey)
		} else {
			cmds[variantID] = pipe.HGet(ctx, dataKey, "count")
		}
	}

	if _, err := pipe.Exec(ctx); err != nil {
		return nil, errors.Wrap(err, "pipeline execution failed")
	}

	for _, variantID := range variantIDs {
		var (
			count  uint64
			data   string
			ruleID string
		)

		if withData {
			cmd, ok := cmds[variantID].(*redis.MapStringStringCmd)
			if !ok {
				continue
			}
			res, err := cmd.Result()
			if err != nil {
				return nil, errors.Wrapf(err, "get data for variant %s", variantID)
			}

			if cnt, ok := res["count"]; ok {
				count, err = strconv.ParseUint(cnt, 10, 64)
				if err != nil {
					return nil, err
				}
			}
			data = res["data"]
			ruleID = res["rule_id"]
		} else {
			cmd, ok := cmds[variantID].(*redis.StringCmd)
			if !ok {
				continue
			}
			count, err = cmd.Uint64()
			if err != nil {
				return nil, err
			}
		}

		score, err := s.conn.ZScore(ctx, keyRuleVariants(service, context), variantID).Result()
		if err != nil {
			return nil, err
		}

		variants = append(variants, model.Variant{
			Key:    variantID,
			Data:   data,
			Count:  count,
			Score:  score,
			RuleID: ruleID,
		})
	}

	return variants, nil
}

func (s *Storage) IncVariantCount(ctx context.Context, service, context, variantID string) error {
	err := s.conn.HIncrBy(ctx, keyVariantData(service, context, variantID), "count", 1).Err()
	return errors.Wrap(err, "increment variant count")
}

func (s *Storage) GetRuleVersion(ctx context.Context, service, context string) (uint64, error) {
	v, err := s.conn.Get(ctx, keyRuleVersion(service, context)).Uint64()
	return v, err
}

func (s *Storage) GetVariantData(ctx context.Context, service, context, variantID string) (string, error) {
	return s.conn.HGet(ctx, keyVariantData(service, context, variantID), "data").Result()
}

func (s *Storage) GetVariantCount(ctx context.Context, service, context, variantID string) (uint64, error) {
	count, err := s.conn.HGet(ctx, keyVariantData(service, context, variantID), "count").Uint64()
	if err != nil {
		return 0, err
	}
	return count, nil
}
