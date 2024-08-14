package godot

import (
	"fmt"
	"regexp"
	"strconv"
)

var re *regexp.Regexp

func Parse(str string) (*Official, error) {
	// TODO: other Godot types.

	if re == nil {
		re = regexp.MustCompile(
			//    1    2   3      4 5      6
			`^4[.](\d+)([.](\d+))?(-(.+?))?(_mono)?$`,
		)
	}

	match := re.FindStringSubmatch(str)
	if match == nil {
		return nil, fmt.Errorf("not a Godot version: '%s'", str)
	}

	minor, _ := strconv.Atoi(match[1])
	patch, _ := strconv.Atoi(match[3]) // NOTE: will be 0 if group excluded

	godot := Official{
		Minor:  uint(minor),
		Patch:  uint(patch),
		Suffix: match[5],
		Mono:   match[6] != "",
	}

	return &godot, nil
}
