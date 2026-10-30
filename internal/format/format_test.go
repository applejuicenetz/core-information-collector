package format

import (
	"math"
	"testing"
)

func TestFileSize(t *testing.T) {
	const k = int64(1024)
	tests := []struct {
		in   int64
		want string
	}{
		{0, "0"}, {-1, "0"}, {math.MinInt64, "0"},
		{1, "1B"}, {1023, "1.023B"},
		{k, "1kB"}, {1536, "1,5kB"}, {k*k - 1, "1.024kB"}, {k * k, "1MB"},
		{k * k * k, "1GB"}, {2 * k * k * k, "2GB"}, {4 * k * k * k, "4GB"},
		{k * k * k * k, "1TB"}, {8 * k * k * k * k * k, "8PB"},
		{15*k*k*k + k*k*k/2, "15,5GB"},
		{1234567, "1,18MB"},
	}
	for _, tc := range tests {
		if got := FileSize(tc.in); got != tc.want {
			t.Errorf("FileSize(%d) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

func TestFileSizeLargestLongStaysInLastUnit(t *testing.T) {
	// 2^63-1 bytes is just below 8192 PB.
	if got, want := FileSize(math.MaxInt64), "8.192PB"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestFileSizeRoundsHalfEven(t *testing.T) {
	// 1.125 kB (1152 bytes) is an exact tie, 1.375 kB (1408) as well.
	if got := FileSize(1152); got != "1,12kB" {
		t.Errorf("1152: %q", got)
	}
	if got := FileSize(1408); got != "1,38kB" {
		t.Errorf("1408: %q", got)
	}
}

func TestInteger(t *testing.T) {
	tests := map[int64]string{
		0: "0", 7: "7", 999: "999", 1000: "1.000", 36: "36", 3182468: "3.182.468",
		-1234: "-1.234", math.MaxInt64: "9.223.372.036.854.775.807",
	}
	for in, want := range tests {
		if got := Integer(in); got != want {
			t.Errorf("Integer(%d) = %q, want %q", in, got, want)
		}
	}
}

func TestNetworkShareSizeAuto(t *testing.T) {
	tests := []struct {
		in   float64
		want string
	}{
		{0, "0,00 MB"},
		{1536, "1,5GB"},
		{2.25 * 1024 * 1024, "2,25TB"},
		{1449.57, "1,41GB"},
		{798 * 1024 * 1024 * 1.0, "798,0TB"},
		{1023, "0,99GB"},
		{1024, "1,0GB"},
		{1024 * 1023, "1023,0GB"},
		{1024 * 1024, "1,0TB"},
		{1024 * 1024 * 1024, "1,0PB"},
		{1024 * 1024 * 1024 * 1024, "1099511627776,0MB"}, // beyond 1024 PB the value stays in MB
	}
	for _, tc := range tests {
		if got := NetworkShareSize(tc.in, FactorAuto); got != tc.want {
			t.Errorf("NetworkShareSize(%v) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

func TestNetworkShareSizeCutsInsteadOfRounding(t *testing.T) {
	if got := NetworkShareSize(1.999*1024, FactorGB); got != "1,99GB" {
		t.Errorf("got %q", got)
	}
}

func TestNetworkShareSizeExplicitFactors(t *testing.T) {
	tests := []struct {
		share  float64
		factor int64
		want   string
	}{
		{2, FactorMB, "2,0MB"},
		{2048, FactorGB, "2,0GB"},
		{2 * 1024 * 1024, FactorTB, "2,0TB"},
		{2 * 1024 * 1024 * 1024, FactorPB, "2,0PB"},
		{1, 7, "0,14??"},
	}
	for _, tc := range tests {
		if got := NetworkShareSize(tc.share, tc.factor); got != tc.want {
			t.Errorf("NetworkShareSize(%v, %d) = %q, want %q", tc.share, tc.factor, got, tc.want)
		}
	}
}
