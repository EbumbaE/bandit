package storage

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/pkg/errors"

	"github.com/EbumbaE/bandit/pkg/clickhouse"
	"github.com/EbumbaE/bandit/pkg/psql"

	model "github.com/EbumbaE/bandit/services/rule-analytic/internal"
)

var ErrNotFound = pgx.ErrNoRows

type Storage struct {
	psqlDB  psql.Database
	clickDB clickhouse.Database
}

func New(ctx context.Context, psqlDB psql.Database, clickDB clickhouse.Database) (*Storage, error) {
	if err := initPsqlSchema(ctx, psqlDB); err != nil {
		return nil, errors.Wrap(err, "init psql schema")
	}
	if err := initClickSchema(ctx, clickDB); err != nil {
		return nil, errors.Wrap(err, "init click schema")
	}
	return &Storage{
		psqlDB:  psqlDB,
		clickDB: clickDB,
	}, nil
}

func initPsqlSchema(ctx context.Context, db psql.Database) error {
	query := `
		CREATE TABLE IF NOT EXISTS analytic_info (
			rule_id UUID,
			variant_id UUID,

			reward DOUBLE PRECISION,
			rule_version BIGINT,
			event_amount BIGINT,
			
			created_at TIMESTAMP NOT NULL DEFAULT now(),
			updated_at TIMESTAMP NOT NULL DEFAULT now()
		);
		
		CREATE UNIQUE INDEX IF NOT EXISTS analytic_info_rule_id_variant_id ON analytic_info(rule_id, variant_id);
`

	_, err := db.Exec(ctx, query)
	return err
}

func initClickSchema(ctx context.Context, db clickhouse.Database) error {
	query := `
		CREATE TABLE IF NOT EXISTS full_analytic_info (
			service       String,
			context       String,
			rule_id       UUID,
			variant_id    UUID,
			rule_version  UInt64,
			action        String,
			amount        String,
			created_at    DateTime DEFAULT now(),
			updated_at    DateTime DEFAULT now()
		) ENGINE = MergeTree()
		ORDER BY (created_at, rule_id, variant_id);
	`

	_, err := db.Exec(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to create table: %w", err)
	}
	return nil
}

func (s *Storage) CreateAnalyticEvent(ctx context.Context, event model.BanditEvent) error {
	query := `
		INSERT INTO analytic_info
		(
			created_at,
			rule_id, variant_id, reward, rule_version, event_amount
		)
		VALUES
		(
			NOW() at time zone 'utc',
			$1, $2, $3, $4, $5
		);
`

	_, err := s.psqlDB.Exec(ctx, query, event.RuleID, event.VariantID, event.Reward, event.RuleVersion, 1)

	return err
}

func (s *Storage) RemoveAnalyticEvent(ctx context.Context, ruleID, variantID string) error {
	query := `
		DELETE FROM analytic_info
		WHERE rule_id = $1 AND variant_id = $2;
`

	_, err := s.psqlDB.Exec(ctx, query, ruleID, variantID)

	return err
}

func (s *Storage) InsertHistoryBatch(ctx context.Context, batch []model.HistoryEvent) error {
	return s.clickDB.WrapBatchWithTx(
		"INSERT INTO full_analytic_info",
		func(tx *sql.Stmt) error {
			for _, event := range batch {
				_, err := tx.Exec(
					event.Payload.Service,
					event.Payload.Context,
					event.Payload.RuleID,
					event.Payload.VariantID,
					event.Payload.RuleVersion,
					event.Action.String(),
					event.Amount,
					nil, nil,
				)
				if err != nil {
					return err
				}
			}

			return nil
		},
	)
}
