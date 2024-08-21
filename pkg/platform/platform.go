package platform

import (
	"runtime"
)

type Platform struct {
	OS   OS
	Arch Arch
}

type OS uint8

const (
	Linux OS = iota
	MacOS
	Windows
)

type Arch uint8

const (
	X86_32 Arch = iota
	X86_64
	ARM_32
	ARM_64
)

func FromRuntime() Platform {
	p := Platform{}

	switch runtime.GOOS {
	case "linux":
		p.OS = Linux
	case "darwin":
		p.OS = MacOS
	case "windows":
		p.OS = Windows
	default:
		panic("Unsupported OS")
	}

	switch runtime.GOARCH {
	case "386":
		p.Arch = X86_32
	case "amd64":
		p.Arch = X86_64
	case "arm":
		p.Arch = ARM_32
	case "arm64":
		p.Arch = ARM_64
	default:
		panic("Unsupported architecture")
	}

	return p
}
