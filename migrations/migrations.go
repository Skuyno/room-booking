// Package migrations встраивает SQL-миграции в бинарник через //go:embed.
// FS используется postgres-репозиторием для применения миграций при старте,
// чтобы не зависеть от расположения файлов на диске.
package migrations

import "embed"

// FS содержит все .sql-файлы этой папки, упакованные в бинарь.
//
//go:embed *.sql
var FS embed.FS
