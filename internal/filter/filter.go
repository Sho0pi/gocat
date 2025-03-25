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
	minLevel     types.LogLevel
	processNames []string
}

// NewLogFilter creates a new LogFilter with the specified filters
func NewLogFilter(
	inputChan <-chan *logreader.LogEntry,
	outputChan chan<- *logreader.LogEntry,
	tags []string,
	minLevel types.LogLevel,
	processNames []string,
) *LogFilter {
	return &LogFilter{
		inputChan:    inputChan,
		outputChan:   outputChan,
		tags:         normalizeTags(tags),
		minLevel:     minLevel,
		processNames: normalizeProcessNames(processNames),
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

	return true
}

// shouldInclude determines if a log entry should be included based on the filters
func (lf *LogFilter) shouldInclude(entry *logreader.LogEntry) bool {
	// Stack trace lines inherit their parent's filtering decision
	if entry.IsStackLine {
		return true
	}

	// If no filters are set, include all entries
	if len(lf.tags) == 0 && len(lf.processNames) == 0 {
		return true
	}

	// Check tag filter
	if len(lf.tags) > 0 {
		tagMatched := false
		for _, tag := range lf.tags {
			if strings.EqualFold(entry.Tag, tag) {
				tagMatched = true
				break
			}
		}
		if !tagMatched {
			return false
		}
	}

	// Check log level filter
	if lf.minLevel.Repr != "he" {

		//levelMatched := false
		//	if (entry.LogLevel) == level {
		//		levelMatched = true
		//		break
		//	}
		//}
		//if !levelMatched {
		//	return false
		//}
	}

	// Check process name filter
	// Note: In the current implementation, we don't have process names, only PIDs
	// This could be enhanced if process name information becomes available
	if len(lf.processNames) > 0 && entry.ProcessName != "" {
		processMatched := false
		for _, process := range lf.processNames {
			if strings.Contains(strings.ToLower(entry.ProcessName), strings.ToLower(process)) {
				processMatched = true
				break
			}
		}
		if !processMatched {
			return false
		}
	}

	return true
}

// Helper functions to normalize filter values

func normalizeTags(tags []string) []string {
	normalized := make([]string, len(tags))
	for i, tag := range tags {
		normalized[i] = strings.TrimSpace(tag)
	}
	return normalized
}

func normalizeLogLevels(levels []string) []string {
	normalized := make([]string, len(levels))
	for i, level := range levels {
		normalized[i] = strings.ToUpper(strings.TrimSpace(level))
	}
	return normalized
}

func normalizeProcessNames(names []string) []string {
	normalized := make([]string, len(names))
	for i, name := range names {
		normalized[i] = strings.TrimSpace(name)
	}
	return normalized
}
