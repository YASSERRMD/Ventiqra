// Package migrations embeds SQL migration and seed files shipped with the
// backend so they can be applied at runtime without external file lookups.
package migrations

import "embed"

// MigrationsFS holds the ordered set of migration SQL files (migrations/*.sql).
//
//go:embed *.sql
var MigrationsFS embed.FS

// SeedsFS holds the ordered set of seed SQL files (migrations/seeds/*.sql).
//
//go:embed seeds/*.sql
var SeedsFS embed.FS
