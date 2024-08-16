package godot

import "testing"

func TestBadStrings(t *testing.T) {
	// NOTE: this includes some relevant bad examples for planned version
	// strings (stolen from the older Python implementation).
	bad := []string{
		"4",
		"4-beta",
		"src:",
		"url:",
		"local:",
		"poop",
		"src:poop",
		"src:remote@",
		"src:remote#",
		"src:remote@#",
	}

	for _, str := range bad {
		_, err := Parse(str)
		if err == nil {
			t.Error(str)
		}
	}
}

func TestOfficial(t *testing.T) {
	compare := func(str string, expected Official) {
		g, err := Parse(str)
		if err != nil {
			t.Errorf("Got error: \"%v\", expected %v", err, expected)
		}
		if g != expected {
			t.Errorf("Got %v, expected %v", g, expected)
		}
	}

	compare("4.0", Official{})
	compare("4.1.2", Official{Minor: 1, Patch: 2})
	compare("4.1.2-beta1", Official{Minor: 1, Patch: 2, Suffix: "beta1"})
	compare("4.1.2-beta1_mono", Official{Minor: 1, Patch: 2, Suffix: "beta1", Mono: true})
}
