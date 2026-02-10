package pkg

import "time"

type SandboxSpec struct {
	CPUCores   float64
	MemoryMB   uint64
	DiskGB     uint64
	Image      string
	DefaultTTL time.Duration
}

var SandboxSpecs = map[string]SandboxSpec{
	"small":  {CPUCores: 0.5, MemoryMB: 512, DiskGB: 10, Image: "alpine:latest", DefaultTTL: 6 * time.Hour},
	"medium": {CPUCores: 2.0, MemoryMB: 4096, DiskGB: 20, Image: "alpine:latest", DefaultTTL: 4 * time.Hour},
	"large":  {CPUCores: 4.0, MemoryMB: 8192, DiskGB: 40, Image: "alpine:latest", DefaultTTL: 2 * time.Hour},
	"xlarge": {CPUCores: 8.0, MemoryMB: 16384, DiskGB: 80, Image: "alpine:latest", DefaultTTL: 1 * time.Hour},
}
