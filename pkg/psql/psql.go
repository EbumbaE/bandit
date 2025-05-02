package psql

import (
	"context"

	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pkg/errors"
)

type Database interface {
	GetSingle(ctx context.Context, pointerOnDst any, query string, args ...any) error
	GetSlice(ctx context.Context, pointerOnSliceDst any, query string, args ...any) error

	Exec(ctx context.Context, sql string, arguments ...any) (commandTag pgconn.CommandTag, err error)
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row

	WrapWithTx(ctx context.Context, exec func(tx pgx.Tx) error) error

	Close()
}

type database struct {
	conn *pgxpool.Pool
}

func NewDatabase(ctx context.Context, dsn string) (Database, error) {
	conn, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, errors.Wrap(err, "database connection")
	}

	db := &database{conn: conn}

	return db, nil
}

func (d *database) GetSingle(ctx context.Context, pointerOnDst any, query string, args ...any) error {
	err := pgxscan.Get(
		ctx,
		d.conn,
		pointerOnDst,
		query,
		args...,
	)
	return err
}

func (d *database) GetSlice(ctx context.Context, pointerOnSliceDst any, query string, args ...any) error {
	err := pgxscan.Select(
		ctx,
		d.conn,
		pointerOnSliceDst,
		query,
		args...,
	)
	return err
}

func (d *database) Exec(ctx context.Context, query string, args ...any) (commandTag pgconn.CommandTag, err error) {
	return d.conn.Exec(ctx, query, args...)
}

func (d *database) Begin(ctx context.Context) (pgx.Tx, error) {
	return d.conn.Begin(ctx)
}

func (d *database) WrapWithTx(ctx context.Context, exec func(tx pgx.Tx) error) error {
	tx, err := d.conn.Begin(ctx)
	if err != nil {
		return errors.Wrapf(err, "begin")
	}

	if execErr := exec(tx); execErr != nil {
		if err := tx.Rollback(ctx); err != nil {
			return errors.Wrapf(err, "rollback err with exec func %s", execErr.Error())
		}
		return errors.Wrapf(execErr, "exec func")
	}

	if err = tx.Commit(ctx); err != nil {
		return errors.Wrapf(err, "scope commit")
	}

	return nil
}

func (d *database) Query(ctx context.Context, query string, args ...any) (pgx.Rows, error) {
	return d.conn.Query(ctx, query, args...)
}

func (d *database) QueryRow(ctx context.Context, query string, args ...any) pgx.Row {
	return d.conn.QueryRow(ctx, query, args...)
}

func (d *database) Close() {
	d.conn.Close()
}
