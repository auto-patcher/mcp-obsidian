package obsidian

import (
	"regexp"
	"strings"
)

var (
	headingRE  = regexp.MustCompile(`^\s{0,3}(#{1,6})\s+(.+?)\s*$`)
	fenceRE    = regexp.MustCompile("^\\s*(```|~~~)")
	trailingRE = regexp.MustCompile(`\s+#+\s*$`)
)

// FindHeadingPaths returns fully-qualified heading paths whose last segment
// matches target (case-insensitive), skipping headings inside fenced code blocks.
// The path joins enclosing heading texts with '::'.
func FindHeadingPaths(content, target string) []string {
	type frame struct {
		level int
		text  string
	}

	targetLower := strings.ToLower(target)
	inFence := false
	var stack []frame
	var matches []string

	for _, line := range strings.Split(content, "\n") {
		if fenceRE.MatchString(line) {
			inFence = !inFence
			continue
		}
		if inFence {
			continue
		}
		m := headingRE.FindStringSubmatch(line)
		if m == nil {
			continue
		}
		level := len(m[1])
		text := strings.TrimSpace(trailingRE.ReplaceAllString(m[2], ""))

		for len(stack) > 0 && stack[len(stack)-1].level >= level {
			stack = stack[:len(stack)-1]
		}
		stack = append(stack, frame{level, text})

		if strings.ToLower(text) == targetLower {
			parts := make([]string, len(stack))
			for i, f := range stack {
				parts[i] = f.text
			}
			matches = append(matches, strings.Join(parts, "::"))
		}
	}
	return matches
}
