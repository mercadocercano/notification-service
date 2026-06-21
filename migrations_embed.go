package notificationservice

import "embed"

// MigrationsFS embeds all migration files for notification-service.
// The "migrations" subdirectory name is required by the go-shared migrate helper
// (iofs.New expects the files under a named subdirectory of the provided FS).
//
// Vive en la raíz del módulo para que //go:embed pueda referenciar el directorio
// hermano migrations/ — src/main.go no puede embeberlo directo porque //go:embed
// no admite paths que escapen del package. El package se llama `notificationservice`
// (el module path notification-service tiene guion, inválido como identificador Go).
//
//go:embed migrations/*.up.sql migrations/*.down.sql
var MigrationsFS embed.FS
