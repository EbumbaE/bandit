package storage

import (
	"context"

	"github.com/pkg/errors"

	"github.com/EbumbaE/bandit/pkg/psql"
	model "github.com/EbumbaE/bandit/services/rule-admin/internal"
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
		CREATE TABLE IF NOT EXISTS rule_info (
			id UUID PRIMARY KEY,

			name TEXT NOT NULL,
			description TEXT NOT NULL,
			state TEXT NOT NULL,
			
			created_at TIMESTAMP NOT NULL DEFAULT now(),
			updated_at TIMESTAMP NOT NULL DEFAULT now(),
			deleted_at TIMESTAMP
		);
			
		CREATE TABLE IF NOT EXISTS variant_info (
			id UUID PRIMARY KEY,
			rule_id UUID,
			
			name TEXT NOT NULL,
			data JSONB,
			state TEXT NOT NULL,
			
			created_at TIMESTAMP NOT NULL DEFAULT now(),
			updated_at TIMESTAMP NOT NULL DEFAULT now(),
			deleted_at TIMESTAMP
		);
		
		CREATE UNIQUE INDEX IF NOT EXISTS variant_info_rule_id ON variant_info(rule_id);
`

	_, err := db.Exec(ctx, query)
	return err
}

func (s *Storage) GetRule(ctx context.Context, id string) (model.Rule, error) {
	var r model.Rule

	query := `
		SELECT id, name, description, state
		FROM rule_info
		WHERE id = $1;
		`

	err := s.conn.GetSingle(ctx, &r, query, id)

	return r, err
}

func (s *Storage) CreateRule(ctx context.Context, rule model.Rule) (model.Rule, error) {
	query := `
		INSERT INTO rule_info
		(
			id, created_at, updated_at,
			name, description, state
		)
		VALUES
		(
			gen_random_uuid(), NOW() at time zone 'utc', NOW() at time zone 'utc',
			$1, $2, $3
		)
		RETURNING id;
`

	var id string
	err := s.conn.QueryRow(ctx, query, rule.Name, rule.Description, rule.State).Scan(&id)

	rule.Id = id

	return rule, err
}

func (s *Storage) UpdateRule(ctx context.Context, rule model.Rule) (model.Rule, error) {
	query := `
		UPDATE rule_info 
		SET 
			name = $2
			description = $3
			updated_at = NOW() at time zone 'utc' 
		WHERE id = $1;
`

	_, err := s.conn.Exec(ctx, query, rule.Id, rule.Name, rule.Description)

	return rule, err
}

func (s *Storage) SetRuleState(ctx context.Context, id string, state model.StateType) error {
	query := `
		UPDATE rule_info 
		SET 
			state = $2
			updated_at = NOW() at time zone 'utc' 
		WHERE id = $1;
`

	_, err := s.conn.Exec(ctx, query, id, state)

	return err
}

func (s *Storage) GetVariant(ctx context.Context, ruleID, variantID string) (model.Variant, error) {
	var v model.Variant

	query := `
		SELECT id, name, data, state
		FROM variant_info
		WHERE id = $1 AND rule_id = $2;
`

	err := s.conn.GetSingle(ctx, &v, query, variantID, ruleID)

	return v, err
}

func (s *Storage) GetVariants(ctx context.Context, ruleID string) ([]model.Variant, error) {
	var v []model.Variant

	query := `
		SELECT id, name, data, state
		FROM variant_info
		WHERE rule_id = $1;
`

	err := s.conn.GetSlice(ctx, &v, query, ruleID)

	return v, err
}

func (s *Storage) AddVariant(ctx context.Context, ruleID string, v model.Variant) (model.Variant, error) {
	query := `
		INSERT INTO variant_info
		(
			id, created_at, updated_at,
			rule_id, name, data, state
		)
		VALUES
		(
			gen_random_uuid(), NOW() at time zone 'utc', NOW() at time zone 'utc',
			$1, $2, $3, $4
		)
		RETURNING id;
`

	var id string
	err := s.conn.QueryRow(ctx, query, ruleID, v.Name, v.Data, v.State).Scan(&id)

	v.Id = id

	return v, err
}

func (s *Storage) SetVariantState(ctx context.Context, id string, state model.StateType) error {
	query := `
		UPDATE variant_info 
		SET 
			state = $2
			updated_at = NOW() at time zone 'utc' 
		WHERE id = $1;
`

	_, err := s.conn.Exec(ctx, query, id, state)

	return err
}
