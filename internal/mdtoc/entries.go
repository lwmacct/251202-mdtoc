package mdtoc

import (
	"strconv"
	"strings"
)

type lineReferenceFunc func(*Header) string

func renderTOCEntries(headers []*Header, opts Options, baseLevel int, lineReference lineReferenceFunc) []string {
	if len(headers) == 0 {
		return nil
	}

	entries := make([]string, 0, len(headers))
	orderedCounters := make(map[int]int)

	for _, h := range headers {
		indentStr := strings.Repeat(" ", (h.Level-baseLevel)*2)
		marker := renderListMarker(h.Level, opts.Ordered, orderedCounters)
		link := renderLink(h, opts.ShowAnchor)

		if opts.LineNumber {
			if ref := lineReference(h); ref != "" {
				link += " `" + ref + "`"
			}
		}

		entries = append(entries, indentStr+marker+" "+link)
	}

	return entries
}

func renderTOCString(headers []*Header, opts Options, baseLevel int, lineReference lineReferenceFunc) string {
	return strings.Join(renderTOCEntries(headers, opts, baseLevel, lineReference), "\n")
}

func minHeaderLevel(headers []*Header) int {
	minLevel := 6
	for _, h := range headers {
		if h.Level < minLevel {
			minLevel = h.Level
		}
	}
	return minLevel
}

func renderListMarker(level int, ordered bool, counters map[int]int) string {
	if !ordered {
		return "-"
	}

	counters[level]++
	for childLevel := level + 1; childLevel <= 6; childLevel++ {
		counters[childLevel] = 0
	}
	return strconv.Itoa(counters[level]) + "."
}

func renderLink(h *Header, showAnchor bool) string {
	if showAnchor {
		return "[" + h.Text + "](#" + h.AnchorLink + ")"
	}
	return "[" + h.Text + "]"
}
