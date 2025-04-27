package clickhouse

import (
	"context"
	"database/sql"

	_ "github.com/ClickHouse/clickhouse-go/v2"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
)

type Database interface {
	GetSlice(ctx context.Context, pointerOnSliceDst any, query string, args ...any) error

	Exec(ctx context.Context, sql string, arguments ...any) (sql.Result, error)
	WrapBatchWithTx(query string, exec func(*sql.Stmt) error) error

	Close() error
}

type database struct {
	db *sqlx.DB
}

func NewDatabase(ctx context.Context, dsn string) (Database, error) {
	conn, err := sqlx.ConnectContext(ctx, "clickhouse", dsn)
	if err != nil {
		return nil, errors.Wrap(err, "failed to connect to database")
	}

	if err := conn.PingContext(ctx); err != nil {
		return nil, errors.Wrap(err, "database ping failed")
	}

	return &database{db: conn}, nil
}

func (d *database) GetSlice(ctx context.Context, pointerOnSliceDst any, query string, args ...any) error {
	return d.db.SelectContext(ctx, pointerOnSliceDst, query, args...)
}

func (d *database) Exec(ctx context.Context, query string, args ...any) (sql.Result, error) {
	return d.db.ExecContext(ctx, query, args...)
}

func (d *database) WrapBatchWithTx(query string, exec func(*sql.Stmt) error) error {
	scope, err := d.db.Begin()
	if err != nil {
		return errors.Wrapf(err, "begin")
	}

	batch, err := scope.Prepare(query)
	if err != nil {
		return errors.Wrapf(err, "begin prepare")
	}

	if execErr := exec(batch); execErr != nil {
		if err := scope.Rollback(); err != nil {
			return errors.Wrapf(err, "rollback err with exec func %s", execErr.Error())
		}
		return errors.Wrapf(execErr, "exec func")
	}

	if err = scope.Commit(); err != nil {
		return errors.Wrapf(err, "scope commit")
	}

	return nil
}

func (d *database) Close() error {
	return d.db.Close()
}
