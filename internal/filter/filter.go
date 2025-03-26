package filter

import (
	"context"
	"github.com/sho0pi/gocat/internal/logreader"
	"github.com/sho0pi/gocat/internal/types"
	"strings"
)

// LogFilter filters log entries based on configured criteria
type LogFilter struct {
	inputChan    <-chan *logreader.LogEntry
	outputChan   chan<- *logreader.LogEntry
	tags         []string
	ignoreTags   []string
	minLevel     types.LogLevel
	processNames []string
}

// NewLogFilter creates a new LogFilter with the specified filters
func NewLogFilter(
	inputChan <-chan *logreader.LogEntry,
	outputChan chan<- *logreader.LogEntry,
	tags []string,
	ignoreTags []string,
	minLevel types.LogLevel,
	processNames []string,
) *LogFilter {
	return &LogFilter{
		inputChan:    inputChan,
		outputChan:   outputChan,
		tags:         tags,
		ignoreTags:   ignoreTags,
		minLevel:     minLevel,
		processNames: processNames,
	}
}

// Start begins filtering log entries from the input channel and sending filtered entries to the output channel
func (lf *LogFilter) Start(ctx context.Context) {
	defer close(lf.outputChan)

	for {
		select {
		case <-ctx.Done():
			return
		case entry, ok := <-lf.inputChan:
			if !ok {
				return
			}

			if lf.filter(entry) {
				lf.outputChan <- entry
			}
		}
	}
}

func (lf *LogFilter) filter(entry *logreader.LogEntry) bool {
	if !lf.validateTag(entry) {
		return false
	}
	if !lf.validateLogLevel(entry) {
		return false
	}

	return true
}

func (lf *LogFilter) validateTag(entry *logreader.LogEntry) bool {
	if len(lf.ignoreTags) != 0 {
		for _, ignoreTag := range lf.ignoreTags {
			if strings.EqualFold(entry.Tag, ignoreTag) {
				return false
			}
		}
	}
	if len(lf.tags) != 0 {
		for _, tag := range lf.tags {
			if strings.EqualFold(entry.Tag, tag) {
				return true
			}
		}
		return false
	}
	return true
}

func (lf *LogFilter) validateLogLevel(entry *logreader.LogEntry) bool {
	return entry.LogLevel.ID >= lf.minLevel.ID
}
