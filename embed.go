// Package sheet bundles the application's static game-data files so they can be
// embedded into the binary at build time. Embedding makes the compiled
// executable self-contained, allowing it to run from any working directory
// without shipping a separate data/ folder alongside the binary.
package sheet

import (
	"embed"
	"io/fs"
)

//go:embed data
var dataFiles embed.FS

// DataFS returns a read-only filesystem rooted at the bundled data directory.
// Files such as "races.json" and "classes.json" appear at the root of the
// returned filesystem (the "data/" prefix is stripped).
func DataFS() fs.FS {
	sub, err := fs.Sub(dataFiles, "data")
	if err != nil {
		// fs.Sub can only fail here if the embedded "data" directory is missing,
		// which is guaranteed to exist at build time by the //go:embed directive.
		panic("sheet: failed to open embedded data filesystem: " + err.Error())
	}
	return sub
}
