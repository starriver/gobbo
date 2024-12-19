package godot

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strings"
)

type Q map[string][]string

// Simple Godot config/resource file query function.
// path is the file to read. q is a map of sections (eg. "[application]", minus
// the brackets) to subkeys (eg. "config/name").
// Use * in a section name to read in arrayed sections (eg. "[preset.0]"...).
// Returns a map of section -> subkey -> value.
// Currently doesn't support in-config maps - these are skipped.
func Query(resourcePath string, q Q) (map[string]map[string]string, error) {
	// Transform q into a 2D map.
	qMap := make(map[string]map[string]struct{}, len(q))
	for k, v := range q {
		inner := make(map[string]struct{}, len(v))
		for _, vv := range v {
			inner[vv] = struct{}{}
		}
		qMap[k] = inner
	}

	r := make(map[string]map[string]string, len(qMap))
	// We track the total keys read so that we can short-circuit if we've
	// already got everything we need.
	totalKeys := 0

	type arrayRegex struct {
		originalKey string
		re          *regexp.Regexp
	}
	ar := []arrayRegex{}

	for k, v := range qMap {
		arrayIndex := strings.IndexRune(k, '*')
		if arrayIndex != -1 {
			// Using * is much more complex. We use an array of regex to keep things
			// relatively sane.
			pattern := fmt.Sprintf(
				"%s[0-9]+%s",
				regexp.QuoteMeta(k[:arrayIndex]),
				regexp.QuoteMeta(k[arrayIndex+1:]),
			)
			ar = append(ar, arrayRegex{k, regexp.MustCompile(pattern)})
		} else {
			// If this doesn't use *, we can be much more efficient.
			c := len(v)
			r[k] = make(map[string]string, c)
			totalKeys += c
		}
	}

	f, err := os.Open(resourcePath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	// Start outside any section.
	// Note these may be nil.
	section, _ := qMap[""]
	rSection, _ := r[""]

	mapDepth := 0
	keysRead := 0
	scanner := bufio.NewScanner(f)

	for scanner.Scan() {
		t := scanner.Text()

		if (len(t) == 0) || (t[0] == ';') {
			continue
		}

		if mapDepth != 0 {
			if t == "}" {
				mapDepth--
			} else if t[len(t)-1] == '{' {
				mapDepth++
			}
			continue
		}

		if t[0] == '[' {
			// Entering a new section.
			sectionName := t[1 : len(t)-1]
			// Again, these may be nil.
			section, _ = qMap[sectionName]
			rSection, _ = r[sectionName]

			// If we're looking for arrays, this'll take longer, and we need to
			// create entries in r as we go:
			if len(ar) != 0 && section == nil {
				for _, a := range ar {
					if !a.re.MatchString(sectionName) {
						continue
					}
					section = qMap[a.originalKey]
					rSection = make(map[string]string, len(section))
				}
			}
			continue
		}

		if t[len(t)-1] == '{' {
			// This is a k/v pair, but its value is a map.
			mapDepth = 1
			continue
		}

		if section == nil {
			// This is after the above conditional in case of any funky map syntax.
			continue
		}

		// This is a k/v pair we can examine.
		// Godot adds spaces around the = in some files, so:
		i := strings.Index(t, "=")
		k := strings.Trim(t[:i], " ")

		if _, ok := section[k]; ok {
			v := strings.Trim(t[i+1:], " ")
			rSection[k] = v

			keysRead += 1
			if keysRead == totalKeys {
				break
			}
		}
	}

	return r, nil
}
