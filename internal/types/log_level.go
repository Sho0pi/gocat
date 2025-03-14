package types

import (
	"github.com/gookit/color"
)

// LogLevel represents the log level of an Android logcat entry
type LogLevel string

const (
	LevelVerbose LogLevel = "V"
	LevelDebug   LogLevel = "D"
	LevelInfo    LogLevel = "I"
	LevelWarning LogLevel = "W"
	LevelError   LogLevel = "E"
	LevelFatal   LogLevel = "F"
)

var logLevelColors = map[LogLevel]color.RGBColor{
	LevelVerbose: color.Hex("#FDF8DC", false),
	LevelDebug:   color.Hex("#4AA6EF"),
	LevelInfo:    color.Hex("#5CD0A7"),
	LevelWarning: color.Hex("#FFD866"),
	LevelError:   color.Hex("#EC6665"),
	LevelFatal:   color.Hex("#EF2D24", false),
}

const whiteColor = "#FDF8DC"

var logLevelStyles = map[LogLevel]*color.RGBStyle{
	LevelVerbose: color.HEXStyle(whiteColor, "#000"),
	LevelDebug:   color.HEXStyle(whiteColor, "#4AA6EF"),
	LevelInfo:    color.HEXStyle("#000", "#5CD0A7"),
	LevelWarning: color.HEXStyle("#000", "#FFD866"),
	LevelError:   color.HEXStyle(whiteColor, "#EC6665"),
	LevelFatal:   color.HEXStyle("#EF2D24", "#000"),
}

func (logLevel LogLevel) Color() color.RGBColor {
	return logLevelColors[logLevel]
}

func (logLevel LogLevel) String() string {
	return logLevelStyles[logLevel].SetOpts(color.Opts{
		color.OpItalic, color.OpBold,
	}).Sprintf(" %s ", string(logLevel))
}

func (logLevel LogLevel) Sprint(a ...any) string {
	return logLevelColors[logLevel].Sprint(a...)
}
