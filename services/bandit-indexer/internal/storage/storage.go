package storage

import (
	"context"

	"github.com/EbumbaE/bandit/pkg/psql"
	"github.com/jackc/pgx/v5"
	"github.com/pkg/errors"

	model "github.com/EbumbaE/bandit/services/bandit-indexer/internal"
)

var ErrNotFound = pgx.ErrNoRows

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
		CREATE TABLE IF NOT EXISTS bandit_info (
			rule_id UUID NOT NULL PRIMARY KEY,
			version BIGINT NOT NULL DEFAULT 0,
			
			bandit_key TEXT NOT NULL,
			config JSONB NOT NULL DEFAULT '{}',
			state TEXT NOT NULL,
			
			created_at TIMESTAMP NOT NULL DEFAULT now(),
			updated_at TIMESTAMP NOT NULL DEFAULT now(),
			deleted_at TIMESTAMP
		);
			
		CREATE TABLE IF NOT EXISTS arm_info (
			variant_id UUID NOT NULL PRIMARY KEY,
			rule_id TEXT NOT NULL,

			count BIGINT NOT NULL DEFAULT 0,
			
			config JSONB NOT NULL DEFAULT '{}',
			state TEXT NOT NULL,
			
			created_at TIMESTAMP NOT NULL DEFAULT now(),
			updated_at TIMESTAMP NOT NULL DEFAULT now(),
			deleted_at TIMESTAMP
		);
		
		CREATE INDEX IF NOT EXISTS arm_info_rule_id ON arm_info(rule_id);
`

	_, err := db.Exec(ctx, query)
	return err
}

func (s *Storage) GetBanditByRuleID(ctx context.Context, ruleID string) (model.Bandit, error) {
	var r model.Bandit

	query := `
		SELECT rule_id, version, bandit_key, config, state
		FROM bandit_info
		WHERE rule_id = $1 AND deleted_at is NULL AND state = $2;
		`

	err := s.conn.GetSingle(ctx, &r, query, ruleID, model.StateTypeEnable)
	if len(r.RuleId) == 0 || errors.Is(err, pgx.ErrNoRows) {
		return model.Bandit{}, ErrNotFound
	}

	return r, err
}

func (s *Storage) CreateBandit(ctx context.Context, bandit model.Bandit) (model.Bandit, error) {
	query := `
		INSERT INTO bandit_info
		(
			created_at, updated_at,
			rule_id, bandit_key, config, state
		)
		VALUES
		(
			NOW() at time zone 'utc', NOW() at time zone 'utc',
			$1, $2, $3, $4
		)
		RETURNING rule_id;
`

	var ruleID string
	err := s.conn.QueryRow(ctx, query, bandit.RuleId, bandit.BanditKey, bandit.Config, bandit.State).Scan(&ruleID)
	if err != nil {
		return model.Bandit{}, err
	}

	if ruleID != bandit.RuleId {
		return model.Bandit{}, errors.New("invalid create")
	}

	return bandit, nil
}

func (s *Storage) SetBanditState(ctx context.Context, ruleID string, state model.StateType) error {
	query := `
		UPDATE bandit_info 
		SET 
			state = $2,
			updated_at = NOW() at time zone 'utc' 
		WHERE rule_id = $1;
`

	_, err := s.conn.Exec(ctx, query, ruleID, state)

	return err
}

func (s *Storage) UpBanditVersion(ctx context.Context, ruleID string) error {
	query := `
		UPDATE bandit_info 
		SET 
			version = version + 1,
			updated_at = NOW() at time zone 'utc' 
		WHERE rule_id = $1;
`

	_, err := s.conn.Exec(ctx, query, ruleID)

	return err
}

func (s *Storage) DeleteBandit(ctx context.Context, ruleID string) error {
	query := `
		UPDATE bandit_info 
		SET 
			deleted_at = NOW() at time zone 'utc',
			updated_at = NOW() at time zone 'utc' 
		WHERE rule_id = $1;
`

	_, err := s.conn.Exec(ctx, query, ruleID)

	return err
}

func (s *Storage) GetArms(ctx context.Context, ruleID string) ([]model.Arm, error) {
	var v []model.Arm

	query := `
		SELECT variant_id, count, config, state
		FROM arm_info
		WHERE rule_id = $1 AND deleted_at is NULL AND state = 'enabled';
`

	err := s.conn.GetSlice(ctx, &v, query, ruleID)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}

	return v, err
}

func (s *Storage) GetArm(ctx context.Context, variantID string) (model.Arm, error) {
	var v model.Arm

	query := `
		SELECT variant_id, count, config, state
		FROM arm_info
		WHERE variant_id = $1 AND deleted_at is NULL AND state = 'enabled';
`

	err := s.conn.GetSingle(ctx, &v, query, variantID)
	if len(v.VariantId) == 0 || errors.Is(err, pgx.ErrNoRows) {
		return model.Arm{}, ErrNotFound
	}

	return v, err
}

func (s *Storage) AddArm(ctx context.Context, ruleID string, v model.Arm) (model.Arm, error) {
	query := `
		INSERT INTO arm_info
		(
			created_at, updated_at,
			rule_id, variant_id, config, state
		)
		VALUES
		(
			NOW() at time zone 'utc', NOW() at time zone 'utc',
			$1, $2, $3, $4
		)
		RETURNING variant_id;
`

	var variantID string
	err := s.conn.QueryRow(ctx, query, ruleID, v.VariantId, v.Config, v.State).Scan(&variantID)
	if err != nil {
		return model.Arm{}, err
	}

	if v.VariantId != variantID {
		return model.Arm{}, errors.New("invalid create")
	}

	return v, nil
}

func (s *Storage) UpdateArm(ctx context.Context, variantID string, config []byte, count uint64) error {
	query := `
		UPDATE arm_info 
		SET 
			config = $2, count = $3
			updated_at = NOW() at time zone 'utc' 
		WHERE variant_id = $1;
`

	_, err := s.conn.Exec(ctx, query, variantID, config, count)

	return err
}

func (s *Storage) SetArmState(ctx context.Context, variantID string, state model.StateType) error {
	query := `
		UPDATE arm_info 
		SET 
			state = $2,
			updated_at = NOW() at time zone 'utc' 
		WHERE variant_id = $1;
`

	_, err := s.conn.Exec(ctx, query, variantID, state)

	return err
}

func (s *Storage) DeleteArm(ctx context.Context, variantID string) error {
	query := `
		UPDATE arm_info 
		SET 
			deleted_at = NOW() at time zone 'utc',
			updated_at = NOW() at time zone 'utc' 
		WHERE variant_id = $1;
`

	_, err := s.conn.Exec(ctx, query, variantID)

	return err
}
