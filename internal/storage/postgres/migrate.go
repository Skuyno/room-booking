package postgres

import (
	"context"
	"fmt"
	"io/fs"
	"path/filepath"
	"slices"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Skuyno/room-booking/migrations"
)

// migrationsDir — имя подкаталога внутри migrations.FS, в котором лежат
// .sql-файлы. embed.FS сохраняет полный путь относительно файла с директивой
// //go:embed, поэтому корнем для нас является имя пакета "migrations".
const migrationsDir = "."

// RunMigrations применяет все .sql-файлы из встроенной FS, которых ещё нет
// в таблице schema_migrations. Все миграции выполняются в одной транзакции —
// либо применяются все pending, либо ничего.
//
// Файлы сортируются лексикографически (поэтому имена принято начинать с
// порядкового номера: 001_init.sql, 002_..., 003_...).
func RunMigrations(ctx context.Context, pool *pgxpool.Pool) error {
	files, err := listMigrationFiles(migrations.FS)
	if err != nil {
		return err
	}

	tx, err := pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin migrations tx: %w", err)
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	if err := ensureSchemaMigrationsTable(ctx, tx); err != nil {
		return fmt.Errorf("ensure schema_migrations table: %w", err)
	}

	applied, err := loadAppliedMigrations(ctx, tx)
	if err != nil {
		return err
	}

	if err := applyPendingMigrations(ctx, tx, migrations.FS, files, applied); err != nil {
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit migrations tx: %w", err)
	}

	return nil
}

// listMigrationFiles возвращает отсортированный список .sql-файлов из
// корня встроенной FS.
func listMigrationFiles(migrationsFS fs.FS) ([]string, error) {
	entries, err := fs.ReadDir(migrationsFS, migrationsDir)
	if err != nil {
		return nil, fmt.Errorf("read migrations dir: %w", err)
	}

	files := make([]string, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".sql" {
			continue
		}
		files = append(files, entry.Name())
	}
	slices.Sort(files)

	return files, nil
}

// ensureSchemaMigrationsTable создаёт таблицу учёта применённых миграций,
// если её ещё нет.
func ensureSchemaMigrationsTable(ctx context.Context, tx pgx.Tx) error {
	_, err := tx.Exec(ctx, `
		create table if not exists schema_migrations (
			version text primary key,
			applied_at timestamptz not null default now()
		)
	`)
	return err
}

// loadAppliedMigrations читает имена уже применённых миграций в множество.
func loadAppliedMigrations(ctx context.Context, tx pgx.Tx) (map[string]struct{}, error) {
	rows, err := tx.Query(ctx, `select version from schema_migrations`)
	if err != nil {
		return nil, fmt.Errorf("list applied migrations: %w", err)
	}
	defer rows.Close()

	applied := make(map[string]struct{})
	for rows.Next() {
		var version string
		if scanErr := rows.Scan(&version); scanErr != nil {
			return nil, fmt.Errorf("scan applied migration: %w", scanErr)
		}
		applied[version] = struct{}{}
	}
	if rowsErr := rows.Err(); rowsErr != nil {
		return nil, fmt.Errorf("iterate applied migrations: %w", rowsErr)
	}

	return applied, nil
}

// applyPendingMigrations выполняет pending-миграции по порядку. Каждая
// миграция и запись в schema_migrations идут в той же транзакции, что и
// внешний tx, — поэтому при ошибке пути всё откатывается атомарно.
func applyPendingMigrations(
	ctx context.Context,
	tx pgx.Tx,
	migrationsFS fs.FS,
	files []string,
	applied map[string]struct{},
) error {
	for _, fileName := range files {
		if _, ok := applied[fileName]; ok {
			continue
		}

		body, err := fs.ReadFile(migrationsFS, fileName)
		if err != nil {
			return fmt.Errorf("read migration %s: %w", fileName, err)
		}

		statement := strings.TrimSpace(string(body))
		if statement == "" {
			continue
		}

		if _, err := tx.Exec(ctx, statement); err != nil {
			return fmt.Errorf("apply migration %s: %w", fileName, err)
		}

		if _, err := tx.Exec(ctx, `
			insert into schema_migrations (version)
			values ($1)
		`, fileName); err != nil {
			return fmt.Errorf("record migration %s: %w", fileName, err)
		}
	}

	return nil
}
