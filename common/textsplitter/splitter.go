package textsplitter

import "strings"

type Options struct {
	// MinSize and MaxSize are measured in runes.
	MinSize int
	MaxSize int
}

// Chunk is a text slice with rune offsets in the original text.
type Chunk struct {
	Start int
	End   int
	Text  string
}

const (
	defaultMinSize = 200
	defaultMaxSize = 400
)

type Splitter struct {
	runes  []rune
	cursor int
	opts   Options
}

type boundaryLevel int

const (
	levelParagraph boundaryLevel = iota
	levelSentence
	levelPunctuation
	levelCount
)

// SplitAll splits the full text with Next until EOF.
func SplitAll(text string, opts Options) []Chunk {
	s := New(text, opts)
	chunks := make([]Chunk, 0)
	for {
		chunk, ok := s.Next()
		if !ok {
			return chunks
		}
		chunks = append(chunks, chunk)
	}
}

func New(text string, opts Options) *Splitter {
	return &Splitter{
		runes: []rune(text),
		opts:  normalizeOptions(opts),
	}
}

// Cursor returns the next rune offset to read from.
func (s *Splitter) Cursor() int {
	return s.cursor
}

func (s *Splitter) Next() (Chunk, bool) {
	if s.cursor < 0 {
		s.cursor = 0
	}
	if s.cursor >= len(s.runes) {
		return Chunk{}, false
	}

	start := s.cursor
	end := nextEnd(s.runes, start, len(s.runes), s.opts.MinSize, s.opts.MaxSize, levelParagraph)
	if end <= start {
		end = min(start+s.opts.MaxSize, len(s.runes))
	}
	s.cursor = end
	return Chunk{
		Start: start,
		End:   end,
		Text:  string(s.runes[start:end]),
	}, true
}

func normalizeOptions(opts Options) Options {
	if opts.MaxSize <= 0 {
		opts.MaxSize = defaultMaxSize
	}
	if opts.MinSize <= 0 {
		opts.MinSize = defaultMinSize
	}
	if opts.MinSize > opts.MaxSize {
		opts.MinSize = opts.MaxSize
	}
	return opts
}

func nextEnd(runes []rune, start int, limitEnd int, minSize int, maxSize int, level boundaryLevel) int {
	if limitEnd-start <= maxSize {
		return limitEnd
	}
	if level >= levelCount {
		return min(start+maxSize, limitEnd)
	}

	currentEnd := start
	unitStart := start
	for unitStart < limitEnd {
		scanEnd := min(limitEnd, start+maxSize+1)
		unitEnd := scanUnitEnd(runes, unitStart, scanEnd, level)
		if unitEnd <= unitStart {
			unitEnd = unitStart + 1
		}

		if unitEnd-start > maxSize {
			if currentEnd-start >= minSize {
				return currentEnd
			}
			return nextEnd(runes, start, unitEnd, minSize, maxSize, level+1)
		}

		currentEnd = unitEnd
		unitStart = unitEnd
	}

	if currentEnd > start {
		return currentEnd
	}
	return min(start+maxSize, limitEnd)
}

func scanUnitEnd(runes []rune, start int, limitEnd int, level boundaryLevel) int {
	for i := start; i < limitEnd; i++ {
		if isBoundary(runes[i], level) {
			end := i + 1
			if level == levelParagraph {
				for end < limitEnd && runes[end] == '\n' {
					end++
				}
			}
			return end
		}
	}
	return limitEnd
}

func isBoundary(r rune, level boundaryLevel) bool {
	switch level {
	case levelParagraph:
		return r == '\n'
	case levelSentence:
		return strings.ContainsRune("。！？!?", r)
	case levelPunctuation:
		return strings.ContainsRune("，,；;、：:", r)
	default:
		return false
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
