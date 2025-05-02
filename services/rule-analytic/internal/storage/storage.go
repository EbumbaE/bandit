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
			count BIGINT,

			created_at TIMESTAMP NOT NULL DEFAULT now(),
			updated_at TIMESTAMP NOT NULL DEFAULT now()
		);
		
		CREATE UNIQUE INDEX IF NOT EXISTS analytic_info_rule_id_variant_id ON analytic_info(rule_id, variant_id, rule_version);
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
			amount        Float64,
			created_at    DateTime DEFAULT now()
		) ENGINE = MergeTree()
		ORDER BY (created_at, rule_id, variant_id);
	`

	_, err := db.Exec(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to create table: %w", err)
	}
	return nil
}

func (s *Storage) ApplyAnalyticEvent(ctx context.Context, events []model.BanditEvent) error {
	query := `
		INSERT INTO analytic_info 
			(created_at, updated_at, rule_id, variant_id, rule_version, reward, count)
		VALUES (
			NOW() at time zone 'utc', NOW() at time zone 'utc',
			$1, $2, $3, $4, $5
		)
		ON CONFLICT (rule_id, variant_id, rule_version) DO UPDATE SET
			reward = analytic_info.reward + EXCLUDED.reward,
			count = analytic_info.count + EXCLUDED.count,
			updated_at = NOW()
`

	err := s.psqlDB.WrapWithTx(ctx, func(tx pgx.Tx) error {
		stmt, err := tx.Prepare(ctx, "apply-events", query)
		if err != nil {
			return errors.Wrap(err, "prepare statement")
		}

		for _, event := range events {
			_, err := tx.Exec(ctx, stmt.SQL,
				event.RuleID,
				event.VariantID,
				event.RuleVersion,
				event.Reward,
				event.Count,
			)
			if err != nil {
				return errors.Wrapf(err, "exec event: %v", event)
			}
		}

		return nil
	})

	return err
}

func (s *Storage) GetAnalyticEvents(ctx context.Context) ([]model.BanditEvent, error) {
	query := `
        SELECT 
            rule_id, variant_id, rule_version, 
            reward, count 
        FROM analytic_info;
    `

	var events []model.BanditEvent
	if err := s.psqlDB.GetSlice(ctx, &events, query); err != nil {
		return nil, errors.Wrap(err, "query events")
	}

	return events, nil
}

func (s *Storage) DeleteAnalyticEvents(ctx context.Context, events []model.BanditEvent) error {
	if len(events) == 0 {
		return nil
	}

	query := `
        DELETE FROM analytic_info
        WHERE rule_id = $1 AND variant_id = $2 AND rule_version = $3
`

	for _, event := range events {
		_, err := s.psqlDB.Exec(ctx, query, event.RuleID, event.VariantID, event.RuleVersion)
		if err != nil {
			return errors.Wrapf(err, "delete event: %v", event)
		}
	}

	return nil
}

func (s *Storage) InsertHistoryBatch(ctx context.Context, batch []model.HistoryEvent) error {
	return s.clickDB.WrapBatchWithTx(
		"INSERT INTO full_analytic_info (service, context, rule_id, variant_id, rule_version, action, amount)",
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
				)
				if err != nil {
					return err
				}
			}

			return nil
		},
	)
}
