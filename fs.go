package lariv

import "io/fs"

// UsefulFilesystem represents a comprehensive filesystem interface combining basic reads, directory lists, and file retrieval capabilities.
// It embeds standard Go [fs.FS], [fs.ReadDirFS], and [fs.ReadFileFS] interfaces.
//
// Use Cases:
//   - Bundling embedded resource directories (e.g. database migration files, template directory pools) in plugin packages.
//
// Example:
//
//	func LoadSqlMigrations(filesystem lariv.UsefulFilesystem) {
//		entries, err := filesystem.ReadDir("migrations")
//		// ... parse and run migrations ...
//	}
type UsefulFilesystem interface {
	fs.FS
	fs.ReadDirFS
	fs.ReadFileFS
}
