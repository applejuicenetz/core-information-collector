// Package buildinfo carries the release version, overridable with -ldflags.
package buildinfo

import "strings"

// Version is the release string. Release builds set it with
// -ldflags "-X github.com/applejuicenetz/core-information-collector/internal/buildinfo.Version=<tag>".
var Version = "dev"

// IsDevelopment reports whether Version is not a release (no update check is made then).
func IsDevelopment() bool {
	return Version == "" || Version == "dev" || strings.Contains(Version, "SNAPSHOT")
}
