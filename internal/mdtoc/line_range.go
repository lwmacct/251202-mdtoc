package mdtoc

import "fmt"

// lineRange describes a heading block as [start, endExclusive).
type lineRange struct {
	Start int
	Count int
}

func newLineRange(start, endInclusive int) lineRange {
	return lineRange{
		Start: start,
		Count: endInclusive - start + 1,
	}
}

func (r lineRange) EndExclusive() int {
	return r.Start + r.Count
}

func formatLineRange(opts Options, r lineRange) string {
	if opts.ShowPath && opts.FilePath != "" {
		return fmt.Sprintf("%s:%d+%d=%d", opts.FilePath, r.Start, r.Count, r.EndExclusive())
	}
	return fmt.Sprintf(":%d+%d=%d", r.Start, r.Count, r.EndExclusive())
}

func formatHeaderLineRange(opts Options, h *Header) string {
	if h.Line <= 0 {
		return ""
	}
	return formatLineRange(opts, newLineRange(h.Line, h.EndLine))
}
