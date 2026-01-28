package setup

import (
	"bytes"
	"os"
	"runtime/debug"
	"strings"

	"github.com/mattn/go-isatty"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type fatalStackHook struct{}

func trimFatalStack(b []byte) []byte {
	// debug.Stack() format is:
	// header line, then repeated pairs of (func line, file line).
	lines := bytes.Split(b, []byte("\n"))
	if len(lines) == 0 {
		return b
	}

	// Keep the goroutine header line.
	out := make([][]byte, 0, len(lines))
	out = append(out, lines[0])

	// Frames are pairs of lines after the header.
	i := 1
	started := false
	for i+1 < len(lines) {
		fn := string(lines[i])

		// Skip empty lines.
		if strings.TrimSpace(fn) == "" {
			i++
			continue
		}

		// Drop frames until we reach the user/app callsite.
		if !started {
			if strings.Contains(fn, "runtime/debug.Stack") ||
				strings.Contains(fn, "fatalStackHook.Run") ||
				strings.Contains(fn, "github.com/rs/zerolog") {
				i += 2
				continue
			}
			started = true
		}

		out = append(out, lines[i], lines[i+1])
		i += 2
	}

	return bytes.Join(out, []byte("\n"))
}

func (h fatalStackHook) Run(e *zerolog.Event, level zerolog.Level, msg string) {
	if level == zerolog.FatalLevel {
		st := trimFatalStack(debug.Stack())
		os.Stderr.WriteString("FATAL: " + msg)
		os.Stderr.WriteString("\n--- STACK TRACE ---\n")
		os.Stderr.Write(st)
		os.Stderr.WriteString("\n")
		e.Str("stack", string(st))
	}
}

type LoggingOptions struct {
	NoColor *bool
}

func SetupLogging() {
	SetupLoggingWithOptions(LoggingOptions{})
}

func SetupLoggingWithOptions(opts LoggingOptions) {
	noColor := !isatty.IsTerminal(os.Stdout.Fd()) && !isatty.IsCygwinTerminal(os.Stdout.Fd())
	if opts.NoColor != nil {
		noColor = *opts.NoColor
	}

	log.Logger = zerolog.New(
		zerolog.ConsoleWriter{
			Out:        os.Stdout,
			NoColor:    noColor,
			TimeFormat: "15:04:05.000",
		},
	).Hook(fatalStackHook{}).With().Timestamp().Logger()
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnixMicro
}
