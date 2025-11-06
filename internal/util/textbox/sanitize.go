package textbox

import "strings"

var messageSanitizer = strings.NewReplacer(
	"…", "...",
	"’", "'",
	"“", "\"",
	"”", "\"",
	"—", "-",
)
