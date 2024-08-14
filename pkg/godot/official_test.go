package godot

import "testing"

func TestStringAndEx(t *testing.T) {
	compare := func(got, expected string) {
		if got != expected {
			t.Errorf("Got %s, expected %s", got, expected)
		}
	}

	g400 := Official{}
	g412Mono := Official{Minor: 1, Patch: 2, Mono: true}
	g412BetaMono := Official{Minor: 1, Patch: 2, Suffix: "beta1", Mono: true}

	compare(g400.String(), "4.0")
	compare(g400.StringEx(false, true, false), "4.0-stable")
	compare(g412Mono.String(), "4.1.2_mono")
	compare(g412Mono.StringEx(false, true, true), "4.1.2-stable_mono")
	compare(g412Mono.StringEx(false, true, false), "4.1.2-stable")
	compare(g412BetaMono.String(), "4.1.2-beta1_mono")
	compare(g412BetaMono.StringEx(true, false, false), "4.1.2.beta1")
	compare(g412BetaMono.StringEx(true, true, false), "4.1.2.beta1")
}

func TestBinaryPath(t *testing.T) {
	compare := func(got, expected string) {
		if got != expected {
			t.Errorf("Got %s, expected %s", got, expected)
		}
	}

	g400 := Official{}
	g412 := Official{Minor: 1, Patch: 2}
	g412Beta := Official{Minor: 1, Patch: 2, Suffix: "beta1"}
	g412BetaMono := Official{Minor: 1, Patch: 2, Suffix: "beta1", Mono: true}

	compare(g400.BinaryPath(), "official/4.0/godot-4.0")
	compare(g412.BinaryPath(), "official/4.1.2/godot-4.1.2")
	compare(g412Beta.BinaryPath(), "official/4.1.2-beta1/godot-4.1.2-beta1")
	compare(g412BetaMono.BinaryPath(), "official/4.1.2-beta1_mono/godot-4.1.2-beta1_mono")
}

func TestDownloadURL(t *testing.T) {
	compare := func(got, expected string) {
		if got != expected {
			t.Errorf("Got %s, expected %s", got, expected)
		}
	}

	g412 := Official{Minor: 1, Patch: 2}
	g412Mono := Official{Minor: 1, Patch: 2, Mono: true}
	g412RC := Official{Minor: 1, Patch: 2, Suffix: "rc1"}
	g412RCMono := Official{Minor: 1, Patch: 2, Suffix: "rc1", Mono: true}

	compare(g412.DownloadURL(Windows), "https://github.com/godotengine/godot/releases/download/4.1.2-stable/Godot_v4.1.2-stable_win64.exe.zip")
	compare(g412.DownloadURL(Linux), "https://github.com/godotengine/godot/releases/download/4.1.2-stable/Godot_v4.1.2-stable_linux.x86_64.zip")
	compare(g412.DownloadURL(MacOS), "https://github.com/godotengine/godot/releases/download/4.1.2-stable/Godot_v4.1.2-stable_macos.universal.zip")
	compare(g412.DownloadURL(ExportTemplates), "https://github.com/godotengine/godot/releases/download/4.1.2-stable/Godot_v4.1.2-stable_export_templates.tpz")
	compare(g412Mono.DownloadURL(Windows), "https://github.com/godotengine/godot/releases/download/4.1.2-stable/Godot_v4.1.2-stable_mono_win64.zip")
	compare(g412Mono.DownloadURL(Linux), "https://github.com/godotengine/godot/releases/download/4.1.2-stable/Godot_v4.1.2-stable_mono_linux_x86_64.zip")
	compare(g412Mono.DownloadURL(MacOS), "https://github.com/godotengine/godot/releases/download/4.1.2-stable/Godot_v4.1.2-stable_mono_macos.universal.zip")
	compare(g412Mono.DownloadURL(ExportTemplates), "https://github.com/godotengine/godot/releases/download/4.1.2-stable/Godot_v4.1.2-stable_mono_export_templates.tpz")
	compare(g412RC.DownloadURL(Windows), "https://github.com/godotengine/godot-builds/releases/download/4.1.2-rc1/Godot_v4.1.2-rc1_win64.exe.zip")
	compare(g412RC.DownloadURL(Linux), "https://github.com/godotengine/godot-builds/releases/download/4.1.2-rc1/Godot_v4.1.2-rc1_linux.x86_64.zip")
	compare(g412RC.DownloadURL(MacOS), "https://github.com/godotengine/godot-builds/releases/download/4.1.2-rc1/Godot_v4.1.2-rc1_macos.universal.zip")
	compare(g412RC.DownloadURL(ExportTemplates), "https://github.com/godotengine/godot-builds/releases/download/4.1.2-rc1/Godot_v4.1.2-rc1_export_templates.tpz")
	compare(g412RCMono.DownloadURL(Windows), "https://github.com/godotengine/godot-builds/releases/download/4.1.2-rc1/Godot_v4.1.2-rc1_mono_win64.zip")
	compare(g412RCMono.DownloadURL(Linux), "https://github.com/godotengine/godot-builds/releases/download/4.1.2-rc1/Godot_v4.1.2-rc1_mono_linux_x86_64.zip")
	compare(g412RCMono.DownloadURL(MacOS), "https://github.com/godotengine/godot-builds/releases/download/4.1.2-rc1/Godot_v4.1.2-rc1_mono_macos.universal.zip")
	compare(g412RCMono.DownloadURL(ExportTemplates), "https://github.com/godotengine/godot-builds/releases/download/4.1.2-rc1/Godot_v4.1.2-rc1_mono_export_templates.tpz")
}
