package tests

import (
	"testing"
	"time"

	"github.com/NjariaOwen/sb-hub/pkg"
)

func TestSandboxSpecs_AllSizesExist(t *testing.T) {
	expectedSizes := []string{"small", "medium", "large", "xlarge"}
	for _, size := range expectedSizes {
		if _, ok := pkg.SandboxSpecs[size]; !ok {
			t.Fatalf("expected size '%s' to exist in SandboxSpecs", size)
		}
	}
}

func TestSandboxSpecs_NoExtraSizes(t *testing.T) {
	validSizes := map[string]bool{"small": true, "medium": true, "large": true, "xlarge": true}
	for size := range pkg.SandboxSpecs {
		if !validSizes[size] {
			t.Fatalf("unexpected size '%s' found in SandboxSpecs", size)
		}
	}
}

func TestSandboxSpecs_ValuesAreValid(t *testing.T) {
	for size, spec := range pkg.SandboxSpecs {
		if spec.CPUCores <= 0 {
			t.Fatalf("size '%s': CPUCores must be > 0, got %f", size, spec.CPUCores)
		}
		if spec.MemoryMB == 0 {
			t.Fatalf("size '%s': MemoryMB must be > 0", size)
		}
		if spec.DiskGB == 0 {
			t.Fatalf("size '%s': DiskGB must be > 0", size)
		}
		if spec.Image == "" {
			t.Fatalf("size '%s': Image must not be empty", size)
		}
		if spec.DefaultTTL <= 0 {
			t.Fatalf("size '%s': DefaultTTL must be > 0, got %v", size, spec.DefaultTTL)
		}
	}
}

func TestSandboxSpecs_ResourceScaling(t *testing.T) {
	sizes := []string{"small", "medium", "large", "xlarge"}

	for i := 1; i < len(sizes); i++ {
		prev := pkg.SandboxSpecs[sizes[i-1]]
		curr := pkg.SandboxSpecs[sizes[i]]

		if curr.CPUCores <= prev.CPUCores {
			t.Fatalf("'%s' CPUCores (%f) should be > '%s' CPUCores (%f)", sizes[i], curr.CPUCores, sizes[i-1], prev.CPUCores)
		}
		if curr.MemoryMB <= prev.MemoryMB {
			t.Fatalf("'%s' MemoryMB (%d) should be > '%s' MemoryMB (%d)", sizes[i], curr.MemoryMB, sizes[i-1], prev.MemoryMB)
		}
		if curr.DiskGB <= prev.DiskGB {
			t.Fatalf("'%s' DiskGB (%d) should be > '%s' DiskGB (%d)", sizes[i], curr.DiskGB, sizes[i-1], prev.DiskGB)
		}
	}
}

func TestSandboxSpecs_TTLInverseScaling(t *testing.T) {
	sizes := []string{"small", "medium", "large", "xlarge"}

	for i := 1; i < len(sizes); i++ {
		prev := pkg.SandboxSpecs[sizes[i-1]]
		curr := pkg.SandboxSpecs[sizes[i]]

		if curr.DefaultTTL >= prev.DefaultTTL {
			t.Fatalf("'%s' TTL (%v) should be < '%s' TTL (%v) — larger sizes should have shorter TTLs",
				sizes[i], curr.DefaultTTL, sizes[i-1], prev.DefaultTTL)
		}
	}
}

func TestSandboxSpecs_SpecificValues(t *testing.T) {
	tests := []struct {
		name   string
		cpu    float64
		memMB  uint64
		diskGB uint64
		image  string
		ttl    time.Duration
	}{
		{"small", 0.5, 512, 10, "alpine:latest", 6 * time.Hour},
		{"medium", 2.0, 4096, 20, "alpine:latest", 4 * time.Hour},
		{"large", 4.0, 8192, 40, "alpine:latest", 2 * time.Hour},
		{"xlarge", 8.0, 16384, 80, "alpine:latest", 1 * time.Hour},
	}

	for _, tt := range tests {
		spec := pkg.SandboxSpecs[tt.name]
		if spec.CPUCores != tt.cpu {
			t.Errorf("size '%s': expected CPUCores %f, got %f", tt.name, tt.cpu, spec.CPUCores)
		}
		if spec.MemoryMB != tt.memMB {
			t.Errorf("size '%s': expected MemoryMB %d, got %d", tt.name, tt.memMB, spec.MemoryMB)
		}
		if spec.DiskGB != tt.diskGB {
			t.Errorf("size '%s': expected DiskGB %d, got %d", tt.name, tt.diskGB, spec.DiskGB)
		}
		if spec.Image != tt.image {
			t.Errorf("size '%s': expected Image '%s', got '%s'", tt.name, tt.image, spec.Image)
		}
		if spec.DefaultTTL != tt.ttl {
			t.Errorf("size '%s': expected TTL %v, got %v", tt.name, tt.ttl, spec.DefaultTTL)
		}
	}
}
