package fscmp

import (
	"bufio"
	"io"
	"strings"
)

type Scanner struct {
	*bufio.Scanner
	LineNum int
}

func NewScanner(r io.Reader) *Scanner {
	return &Scanner{
		Scanner: bufio.NewScanner(r),
		LineNum: 0,
	}
}

// NextNonEmptyLine scans to the next non-empty line. Non-empty is a line that contains text
// that is not trimmed from strings.TrimSpace
func (s *Scanner) NextNonEmptyLine() string {
	text := s.Text()
	for text == "" {
		if s.Scan() {
			text = strings.TrimSpace(s.Text())
		} else {
			break
		}
	}
	return text
}

func (s *Scanner) Scan() bool {
	s.LineNum++
	return s.Scanner.Scan()
}
