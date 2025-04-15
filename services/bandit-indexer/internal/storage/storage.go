package storage

import (
	"context"

	"github.com/EbumbaE/bandit/pkg/psql"
	"github.com/pkg/errors"

	model "github.com/EbumbaE/bandit/services/bandit-indexer/internal"
)

type Storage struct {
	conn psql.Database
}

func New(ctx context.Context, conn psql.Database) (*Storage, error) {
	if err := initSchema(ctx, conn); err != nil {
		return nil, errors.Wrap(err, "init schema")
	}
	return &Storage{
		conn: conn,
	}, nil
}

func initSchema(ctx context.Context, db psql.Database) error {
	query := `
		CREATE TABLE IF NOT EXISTS wanted_registry (
			key TEXT NOT NULL PRIMARY KEY,
			name TEXT NOT NULL,
			
			created_at TIMESTAMP NOT NULL DEFAULT now(),
			deleted_at TIMESTAMP
		);

		CREATE TABLE IF NOT EXISTS bandit_info (
			rule_id UUID NOT NULL,
			version BIGINT NOT NULL,
			
			bandit_key TEXT NOT NULL,
			config JSONB NOT NULL,
			state TEXT NOT NULL,
			
			created_at TIMESTAMP NOT NULL DEFAULT now(),
			updated_at TIMESTAMP NOT NULL DEFAULT now(),
			deleted_at TIMESTAMP
		);
		CREATE UNIQUE INDEX IF NOT EXISTS bandit_info_rule_id ON bandit_info(rule_id);
			
		CREATE TABLE IF NOT EXISTS arm_info (
			variant_id UUID  NOT NULL,
			rule_id TEXT NOT NULL,

			data JSONB NOT NULL,
			count BIGINT NOT NULL,
			
			config JSONB,
			state TEXT NOT NULL,
			
			created_at TIMESTAMP NOT NULL DEFAULT now(),
			deleted_at TIMESTAMP
		);
		
		CREATE INDEX IF NOT EXISTS arm_info_bandit_id ON arm_info(bandit_id);
`

	_, err := db.Exec(ctx, query)
	return err
}

func (s *Storage) GetBanditByRuleID(ctx context.Context, ruleID string) (model.Bandit, error) {
	var r model.Bandit

	query := `
		SELECT id, rule_id, service, context, config, bandit_key, state
		FROM bandit_info
		WHERE rule_id = $1;
		`

	err := s.conn.GetSingle(ctx, &r, query, ruleID)

	return r, err
}

func (s *Storage) CreateBandit(ctx context.Context, bandit model.Bandit) (model.Bandit, error) {
	query := `
		INSERT INTO bandit_info
		(
			id, created_at, updated_at,
			rule_id, service, context, config, bandit_key, state
		)
		VALUES
		(
			gen_random_uuid(), NOW() at time zone 'utc', NOW() at time zone 'utc',
			$1, $2, $3, $4, $5, $6
		)
		RETURNING id;
`

	var id string
	err := s.conn.QueryRow(ctx, query, bandit.RuleId, bandit.Service, bandit.Context, bandit.Config, bandit.BanditKey, bandit.State).Scan(&id)

	bandit.Id = id

	return bandit, err
}

func (s *Storage) SetBanditStateByRuleID(ctx context.Context, ruleID string, state model.StateType) error {
	query := `
		UPDATE bandit_info 
		SET 
			state = $2
			updated_at = NOW() at time zone 'utc' 
		WHERE rule_id = $1;
`

	_, err := s.conn.Exec(ctx, query, ruleID, state)

	return err
}

func (s *Storage) GetArms(ctx context.Context, banditID string) ([]model.Arm, error) {
	var v []model.Arm

	query := `
		SELECT id, data, count, variant_id, config, state
		FROM arm_info
		WHERE bandit_id = $1;
`

	err := s.conn.GetSlice(ctx, &v, query, banditID)

	return v, err
}

func (s *Storage) AddArm(ctx context.Context, banditID string, v model.Arm) (model.Arm, error) {
	query := `
		INSERT INTO arm_info
		(
			id, created_at,
			bandit_id, data, count, variant_id, config, state
		)
		VALUES
		(
			gen_random_uuid(), NOW() at time zone 'utc', 
			$1, $2, $3, $4, $5, $6
		)
		RETURNING id;
`

	var id string
	err := s.conn.QueryRow(ctx, query, banditID, v.Data, v.Count, v.VariantId, v.Config, v.State).Scan(&id)

	v.Id = id

	return v, err
}

func (s *Storage) SetArmStateByVariantID(ctx context.Context, variantID string, state model.StateType) error {
	query := `
		UPDATE arm_info 
		SET 
			state = $2
			updated_at = NOW() at time zone 'utc' 
		WHERE variant_id = $1;
`

	_, err := s.conn.Exec(ctx, query, variantID, state)

	return err
}
