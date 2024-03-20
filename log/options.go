package log

import (
	"time"

	"cosmossdk.io/log"
	"github.com/rs/zerolog"
)

var defaultConfig = log.Config{
	Level:      zerolog.NoLevel,
	Filter:     nil,
	OutputJSON: false,
	Color:      true,
	StackTrace: false,
	TimeFormat: time.Kitchen,
}
