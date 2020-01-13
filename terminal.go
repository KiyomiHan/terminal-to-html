/*
Package terminal converts ANSI input to HTML output.

The generated HTML needs to be used with the stylesheet at
https://raw.githubusercontent.com/buildkite/terminal-to-html/master/assets/terminal.css
and wrapped in a term-container div.

You can call this library from the command line with terminal-to-html:
go install github.com/buildkite/terminal-to-html/cmd/terminal-to-html
*/
package terminal

import (
	"bytes"
	"regexp"
	"strings"
	"fmt"
)

// Render converts ANSI to HTML and returns the result.
func Render(input string) []byte {
	
	pattern, _ := regexp.Compile(`(https|http)://(([#-~!])+)`)
	link := "<a onclick=\"window.open('%s', '_blank')\">%s</a>"

	lines := strings.Split(input, "\n")
	urls := make([][]string,len(lines))

	for line_num, line := range lines {
		line_loc := pattern.FindAllStringIndex(line, -1)
		for _, index := range line_loc{
			newStr := fmt.Sprintf(link, line[index[0]: index[1]], line[index[0]: index[1]])
			urls[line_num] = append(urls[line_num], newStr)
		}
	}
	screen := screen{}
	screen.parse([]byte(input))
	output := bytes.Replace(screen.asHTML(urls), []byte("\n\n"), []byte("\n&nbsp;\n"), -1)
	return output
}
