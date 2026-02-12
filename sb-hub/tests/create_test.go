package tests

import (
	"net"
	"testing"

	"github.com/NjariaOwen/sb-hub/cmd"
)

func TestFindFreePort_ReturnsFirstAvailable(t *testing.T) {
	used := map[string]bool{}
	port := cmd.FindFreePort(49500, 49510, used)
	if port < 49500 || port > 49510 {
		t.Fatalf("expected port in range 49500-49510, got %d", port)
	}
}

func TestFindFreePort_SkipsUsedPorts(t *testing.T) {
	used := map[string]bool{
		"49500": true,
		"49501": true,
		"49502": true,
	}
	port := cmd.FindFreePort(49500, 49510, used)
	if port == 49500 || port == 49501 || port == 49502 {
		t.Fatalf("expected port to skip used ports, got %d", port)
	}
	if port < 49503 || port > 49510 {
		t.Fatalf("expected port in range 49503-49510, got %d", port)
	}
}

func TestFindFreePort_AllBusy(t *testing.T) {
	// Occupy a small range of ports
	listeners := make([]net.Listener, 0)
	used := map[string]bool{}

	// Use ports in a very narrow range that we also physically listen on
	for p := 49550; p <= 49552; p++ {
		ln, err := net.Listen("tcp", ":0") // Bind to random ports
		if err == nil {
			listeners = append(listeners, ln)
		}
	}
	defer func() {
		for _, ln := range listeners {
			ln.Close()
		}
	}()

	// Mark all ports in range as used
	used["49550"] = true
	used["49551"] = true
	used["49552"] = true

	port := cmd.FindFreePort(49550, 49552, used)
	if port != 0 {
		t.Fatalf("expected 0 when all ports are used, got %d", port)
	}
}

func TestFindFreePort_EmptyRange(t *testing.T) {
	used := map[string]bool{}
	// Start > End should return 0
	port := cmd.FindFreePort(49510, 49500, used)
	if port != 0 {
		t.Fatalf("expected 0 for empty range, got %d", port)
	}
}

func TestFindFreePort_SinglePort(t *testing.T) {
	used := map[string]bool{}
	port := cmd.FindFreePort(49560, 49560, used)
	if port != 49560 {
		t.Fatalf("expected port 49560, got %d", port)
	}
}
