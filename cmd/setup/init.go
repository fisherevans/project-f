package setup

import (
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/rs/zerolog/pkgerrors"
	"os"
)

func init() {
	log.Logger = zerolog.New(
		zerolog.ConsoleWriter{
			Out:     os.Stdout,
			NoColor: false,
		},
	).With().Timestamp().Logger()
	zerolog.SetGlobalLevel(zerolog.TraceLevel)
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnixMicro
	zerolog.ErrorStackMarshaler = pkgerrors.MarshalStack
}
