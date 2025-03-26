package types

import (
	"fmt"
	"github.com/gookit/color"
	"strings"
)

// LogLevel represents the log level of an Android logcat entry
type LogLevel struct {
	ID         int
	Name       string
	LevelColor color.RGBColor
	LevelStyle *color.RGBStyle
	Repr       string
}

const (
	whiteColor   = "#FDF8DC"
	blackColor   = "#000000"
	blueColor    = "#4AA6EF"
	greenColor   = "#5CD0A7"
	yellowColor  = "#FFD866"
	redColor     = "#EC6665"
	darkRedColor = "#EF2D24"
)

var logLevels = map[string]LogLevel{
	"v": LogLevel{
		ID:         0,
		Name:       "verbose",
		LevelColor: color.Hex(whiteColor, false),
		LevelStyle: color.HEXStyle(whiteColor, blackColor),
		Repr:       "V",
	},
	"d": LogLevel{
		ID:         1,
		Name:       "debug",
		LevelColor: color.Hex(blueColor),
		LevelStyle: color.HEXStyle(whiteColor, blueColor),
		Repr:       "D",
	},
	"i": LogLevel{
		ID:         2,
		Name:       "info",
		LevelColor: color.Hex(greenColor),
		LevelStyle: color.HEXStyle(blackColor, greenColor),
		Repr:       "I",
	},
	"w": LogLevel{
		ID:         3,
		Name:       "warning",
		LevelColor: color.Hex(yellowColor),
		LevelStyle: color.HEXStyle(blackColor, yellowColor),
		Repr:       "W",
	},
	"e": LogLevel{
		ID:         4,
		Name:       "error",
		LevelColor: color.Hex(redColor),
		LevelStyle: color.HEXStyle(whiteColor, redColor),
		Repr:       "E",
	},
	"f": LogLevel{
		ID:         5,
		Name:       "fatal",
		LevelColor: color.Hex(darkRedColor, false),
		LevelStyle: color.HEXStyle(darkRedColor, blackColor),
		Repr:       "F",
	},
}

var VerboseLevel = logLevels["v"]

const (
	LevelVerbose = iota
	LevelDebug
	LevelInfo
	LevelWarning
	LevelError
	LevelFatal
)

func ToLogLevel(level string) LogLevel {
	return logLevels[strings.ToLower(level)]
}

func (logLevel *LogLevel) Color() color.RGBColor {
	return logLevel.LevelColor
}

func (logLevel *LogLevel) String() string {
	return logLevel.Name
}

func (logLevel *LogLevel) Pretty() string {
	return logLevel.LevelStyle.SetOpts(color.Opts{
		color.OpItalic, color.OpBold,
	}).Sprintf(" %s ", logLevel.Repr)
}

func (logLevel *LogLevel) Sprint(a ...any) string {
	return logLevel.LevelColor.Sprint(a...)
}

func (logLevel *LogLevel) Set(s string) error {
	l := string(s[0])
	level, ok := logLevels[strings.ToLower(l)]
	if !ok {
		return fmt.Errorf("invalid log level: %s", s)
	}
	*logLevel = level
	return nil
}

func (logLevel *LogLevel) Type() string {
	return "Level"
}
