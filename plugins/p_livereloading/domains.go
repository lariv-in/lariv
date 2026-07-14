package p_livereloading

import "strings"

// allowedHosts lists hostnames where the live-reload client may connect.
// Add a domain here when live reload should work on a new environment.
var allowedHosts = []string{
	"localhost",
}

func allowedHostsJS() string {
	quoted := make([]string, len(allowedHosts))
	for i, host := range allowedHosts {
		quoted[i] = "'" + strings.ReplaceAll(host, "'", "\\'") + "'"
	}
	return "[" + strings.Join(quoted, ",") + "]"
}
