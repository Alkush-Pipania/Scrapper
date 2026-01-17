package logger

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/Alkush-Pipania/Scrapper/config"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func Init(cfg *config.Config) *zerolog.Logger {

	const prodStr string = "production"

	// Set global level based on environment
	switch cfg.Env {
	case prodStr:
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	default:
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	}

	var baseLogger zerolog.Logger

	if cfg.Env == prodStr {
		baseLogger = zerolog.New(os.Stdout)
	} else {
		baseLogger = zerolog.New(zerolog.ConsoleWriter{
			Out:        os.Stdout,
			TimeFormat: time.RFC3339,
			NoColor:    false, // Enable colors
			PartsOrder: []string{
				"time", "level", "caller", "service", "env", "message", "err",
			},
			FormatLevel: func(i any) string {
				return strings.ToUpper(fmt.Sprintf("[%s]", i))
			},
			FormatCaller: func(caller any) string {
				return fmt.Sprintf("(%s)", caller)
			},
		})
	}

	baseLogger = baseLogger.With().
		Timestamp().
		Str("env", cfg.Env).
		Logger() // finalize

	// Add caller info for dev
	if cfg.Env != prodStr {
		baseLogger = baseLogger.With().Caller().Logger()
	}

	log.Logger = baseLogger

	return &baseLogger
}
