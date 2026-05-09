// Package postgres содержит реализации репозиториев из internal/service на
// pgx/v5. Каждый репозиторий принимает DBTX, поэтому одинаковый код
// работает и поверх *pgxpool.Pool, и поверх pgx.Tx.
package postgres

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

// DBTX — общий интерфейс над pgx-соединением. Его реализуют и
// *pgxpool.Pool, и pgx.Tx, благодаря чему репозитории не зависят от того,
// в каком контексте их используют (обычный запрос или транзакция).
type DBTX interface {
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}
