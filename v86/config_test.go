// config.go's parsing helpers are portable stdlib and intentionally carry no
// js/wasm build constraint, so parseFlags can be unit-tested on the host with
// plain `go test` (the wasm test runner is not required).

package main

import "testing"

// TestParseFlagsRelayFlag verifies the primary path: a -relay flag sets the
// top-level network_relay_url option (consumed by v86 at libv86.mjs:276).
// net_device is intentionally left unset so main.go's virtio default applies,
// matching the guest's virtio_net driver (forcing ne2k breaks the virtio-9p
// rootfs in this v86 build).
func TestParseFlagsRelayFlag(t *testing.T) {
	cfg, err := parseFlags([]string{"-relay", "ws://relay.example.com:7654/net"})
	if err != nil {
		t.Fatalf("parseFlags: %v", err)
	}
	if got := cfg["network_relay_url"]; got != "ws://relay.example.com:7654/net" {
		t.Errorf("network_relay_url = %v, want ws://relay.example.com:7654/net", got)
	}
	if _, ok := cfg["net_device"]; ok {
		t.Errorf("net_device must be absent with relay-only (main.go defaults virtio); got %v", cfg["net_device"])
	}
}

// TestParseFlagsRelayEnvOverride verifies the VM_RELAY_URL env path, mirroring
// the existing VM_APPEND precedent. This is how the <wanix-vm relay="...">
// attribute delivers the URL (env survives the whitespace-split arg string).
func TestParseFlagsRelayEnvOverride(t *testing.T) {
	t.Setenv("VM_RELAY_URL", "ws://env.example.com:1234/")
	cfg, err := parseFlags(nil)
	if err != nil {
		t.Fatalf("parseFlags: %v", err)
	}
	if got := cfg["network_relay_url"]; got != "ws://env.example.com:1234/" {
		t.Errorf("network_relay_url = %v, want env value", got)
	}
	if _, ok := cfg["net_device"]; ok {
		t.Errorf("net_device must be absent with relay env (main.go defaults virtio); got %v", cfg["net_device"])
	}
}

// TestParseFlagsExplicitNetdevWins verifies an explicit netdev is not clobbered
// by the relay ne2k default: the two are orthogonal opts keys.
func TestParseFlagsExplicitNetdevWins(t *testing.T) {
	cfg, err := parseFlags([]string{"-relay", "ws://r/", "-netdev", "user,type=virtio"})
	if err != nil {
		t.Fatalf("parseFlags: %v", err)
	}
	if cfg["network_relay_url"] != "ws://r/" {
		t.Errorf("network_relay_url not set alongside explicit netdev: %v", cfg["network_relay_url"])
	}
	if nd := cfg["net_device"].(map[string]any)["type"]; nd != "virtio" {
		t.Errorf("explicit netdev must win; got net_device.type=%v", nd)
	}
}

// TestParseFlagsNoRelayUnchanged verifies that without a relay, parseFlags sets
// neither network_relay_url nor net_device, so main.go's existing virtio
// default (main.go:123) applies unchanged — i.e. the patch is purely additive.
func TestParseFlagsNoRelayUnchanged(t *testing.T) {
	cfg, err := parseFlags(nil)
	if err != nil {
		t.Fatalf("parseFlags: %v", err)
	}
	if _, ok := cfg["network_relay_url"]; ok {
		t.Errorf("network_relay_url must be absent without relay; got %v", cfg["network_relay_url"])
	}
	if _, ok := cfg["net_device"]; ok {
		t.Errorf("net_device must be absent without relay (main.go defaults to virtio); got %v", cfg["net_device"])
	}
}
