// Package migrations embeds SQL migration files shipped with the backend so
// they can be applied at runtime without external file lookups.
package migrations

import "embed"

// MigrationsFS holds the ordered set of migration SQL files (migrations/*.sql).
//
//go:embed *.sql
var MigrationsFS embed.FS
