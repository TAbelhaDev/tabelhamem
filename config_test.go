package main

import (
	"testing"
)

func TestNormalizeDefault(t *testing.T) {
	c := defaultConfig()
	got := normalize(c)
	if got.Layout.SidebarWidthShare != 1 {
		t.Errorf("SidebarWidthShare = %d, want 1", got.Layout.SidebarWidthShare)
	}
	if got.Layout.RightWidthShare != 4 {
		t.Errorf("RightWidthShare = %d, want 4", got.Layout.RightWidthShare)
	}
}

func TestNormalizeFloor(t *testing.T) {
	c := config{
		Layout: layoutConfig{
			SidebarWidthShare: 0,
			RightWidthShare:   0,
			StatsHeightShare:  0,
			MemoryHeightShare: 0,
		},
	}
	got := normalize(c)
	if got.Layout.SidebarWidthShare < 1 {
		t.Errorf("SidebarWidthShare = %d, want >= 1", got.Layout.SidebarWidthShare)
	}
	if got.Layout.RightWidthShare < 1 {
		t.Errorf("RightWidthShare = %d, want >= 1", got.Layout.RightWidthShare)
	}
	if got.Layout.StatsHeightShare < 1 {
		t.Errorf("StatsHeightShare = %d, want >= 1", got.Layout.StatsHeightShare)
	}
	if got.Layout.MemoryHeightShare < 1 {
		t.Errorf("MemoryHeightShare = %d, want >= 1", got.Layout.MemoryHeightShare)
	}
}

func TestNormalizePreservesPositive(t *testing.T) {
	c := config{
		Layout: layoutConfig{
			SidebarWidthShare: 3,
			RightWidthShare:   5,
			StatsHeightShare:  2,
			MemoryHeightShare: 6,
		},
	}
	got := normalize(c)
	if got.Layout.SidebarWidthShare != 3 {
		t.Errorf("SidebarWidthShare = %d, want 3", got.Layout.SidebarWidthShare)
	}
	if got.Layout.RightWidthShare != 5 {
		t.Errorf("RightWidthShare = %d, want 5", got.Layout.RightWidthShare)
	}
}

func TestBridgeGlyph(t *testing.T) {
	tests := []struct {
		claude, opencode bool
		want             string
	}{
		{true, true, "✓"},
		{true, false, "●"},
		{false, false, "○"},
		{false, true, "○"},
	}
	for _, tt := range tests {
		got := bridgeGlyph(tt.claude, tt.opencode)
		if got == "" {
			t.Errorf("bridgeGlyph(%v, %v) is empty", tt.claude, tt.opencode)
		}
	}
}

func TestStatusDot(t *testing.T) {
	if got := statusDot(true); got == "" {
		t.Error("statusDot(true) is empty")
	}
	if got := statusDot(false); got == "" {
		t.Error("statusDot(false) is empty")
	}
}
