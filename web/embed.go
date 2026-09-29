// Package web expone el viewer del personal (fuente única: viewer.html)
// para embeberlo en el binario sin duplicar archivos ni builds.
package web

import "embed"

//go:embed viewer.html
var FS embed.FS
