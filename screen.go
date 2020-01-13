package terminal

import (
	"math"
	"strconv"
	"strings"
	"regexp"
)

// A terminal 'screen'. Current cursor position, cursor style, and characters
type screen struct {
	x      int
	y      int
	screen [][]node
	style  *style
}

const screenEndOfLine = -1
const screenStartOfLine = 0

// Clear part (or all) of a line on the screen
func (s *screen) clear(y int, xStart int, xEnd int) {
	if len(s.screen) <= y {
		return
	}

	if xStart == screenStartOfLine && xEnd == screenEndOfLine {
		s.screen[y] = make([]node, 0, 80)
	} else {
		line := s.screen[y]

		if xEnd == screenEndOfLine {
			xEnd = len(line) - 1
		}
		for i := xStart; i <= xEnd && i < len(line); i++ {
			line[i] = emptyNode
		}
	}
}

// "Safe" parseint for parsing ANSI instructions
func pi(s string) int {
	if s == "" {
		return 1
	}
	i, _ := strconv.ParseInt(s, 10, 8)
	return int(i)
}

// Move the cursor up, if we can
func (s *screen) up(i string) {
	s.y -= pi(i)
	s.y = int(math.Max(0, float64(s.y)))
}

// Move the cursor down
func (s *screen) down(i string) {
	s.y += pi(i)
}

// Move the cursor forward on the line
func (s *screen) forward(i string) {
	s.x += pi(i)
}

// Move the cursor backward, if we can
func (s *screen) backward(i string) {
	s.x -= pi(i)
	s.x = int(math.Max(0, float64(s.x)))
}

// Add rows to our screen if necessary
func (s *screen) growScreenHeight() {
	for i := len(s.screen); i <= s.y; i++ {
		s.screen = append(s.screen, make([]node, 0, 80))
	}
}

// Add columns to our current line if necessary
func (s *screen) growLineWidth(line []node) []node {
	for i := len(line); i <= s.x; i++ {
		line = append(line, emptyNode)
	}
	return line
}

// Write a character to the screen's current X&Y, along with the current screen style
func (s *screen) write(data rune) {
	s.growScreenHeight()

	line := s.screen[s.y]
	line = s.growLineWidth(line)

	line[s.x] = node{blob: data, style: s.style}
	s.screen[s.y] = line
}

// Append a character to the screen
func (s *screen) append(data rune) {
	s.write(data)
	s.x++
}

// Append multiple characters to the screen
func (s *screen) appendMany(data []rune) {
	for _, char := range data {
		s.append(char)
	}
}

func (s *screen) appendElement(i *element) {
	s.growScreenHeight()
	line := s.growLineWidth(s.screen[s.y])

	line[s.x] = node{style: s.style, elem: i}
	s.screen[s.y] = line
	s.x++
}

// Apply color instruction codes to the screen's current style
func (s *screen) color(i []string) {
	s.style = s.style.color(i)
}

// Apply an escape sequence to the screen
func (s *screen) applyEscape(code rune, instructions []string) {
	if len(instructions) == 0 {
		// Ensure we always have a first instruction
		instructions = []string{""}
	}

	switch code {
	case 'M':
		s.color(instructions)
	case 'G':
		s.x = 0
	case 'K':
		switch instructions[0] {
		case "0", "":
			s.clear(s.y, s.x, screenEndOfLine)
		case "1":
			s.clear(s.y, screenStartOfLine, s.x)
		case "2":
			s.clear(s.y, screenStartOfLine, screenEndOfLine)
		}
	case 'A':
		s.up(instructions[0])
	case 'B':
		s.down(instructions[0])
	case 'C':
		s.forward(instructions[0])
	case 'D':
		s.backward(instructions[0])
	}
}

// Parse ANSI input, populate our screen buffer with nodes
func (s *screen) parse(ansi []byte) {
	s.style = &emptyStyle
	parseANSIToScreen(s, ansi)
}

func (s *screen) asHTML(urls [][]string) []byte {
	var lines []string
	pattern, _ := regexp.Compile(`(https|http):&#47;&#47;(([#-~!])+)`)

	for line_num, line := range s.screen {
		html_line := outputLineAsHTML(line)
		new_lines := ""
		if len(urls[line_num]) != 0 {
			loc := pattern.FindAllStringIndex(html_line, -1)
			var new_line_list []string
			new_line_list = append(new_line_list, html_line[:loc[0][0]])
			new_line_list = append(new_line_list, urls[line_num][0])
			for i := 1; i<len(loc); i++ {
				new_line_list = append(new_line_list, html_line[loc[i-1][1]:loc[i][0]])
				new_line_list = append(new_line_list, urls[line_num][i])
			}
			new_line_list = append(new_line_list, html_line[loc[len(loc)-1][1]:])
			new_lines = strings.Join(new_line_list, "")
		}else {
			new_lines = html_line
		}
		lines = append(lines, new_lines)
	}

	return []byte(strings.Join(lines, "\n"))
}

func (s *screen) newLine() {
	s.x = 0
	s.y++
}

func (s *screen) carriageReturn() {
	s.x = 0
}

func (s *screen) backspace() {
	if s.x > 0 {
		s.x--
	}
}
