package database

import (
	"context"

	"logistics-platform/lib/config"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4"
)

type PostgresDB struct {
	conn *pgx.Conn
}

func NewPostgresDB() (*PostgresDB, error) {
	dsn := config.GetDBConnectionString()
	conn, err := pgx.Connect(context.Background(), dsn)
	if err != nil {
		return nil, err
	}
	return &PostgresDB{conn: conn}, nil
}

func (db *PostgresDB) Exec(ctx context.Context, sql string, arguments ...interface{}) (pgconn.CommandTag, error) {
	return db.conn.Exec(ctx, sql, arguments...)
}

func (db *PostgresDB) Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error) {
	return db.conn.Query(ctx, sql, args...)
}

func (db *PostgresDB) QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row {
	return db.conn.QueryRow(ctx, sql, args...)
}

func (db *PostgresDB) Close(ctx context.Context) error {
	return db.conn.Close(ctx)
}
