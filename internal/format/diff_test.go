package format

import (
	"bufio"
	"os"
	"strconv"
	"strings"
	"testing"
)

// TestAgainstJavaReference compares with output of the former Java implementation (tag 4.0.1,
// German locale): testdata/java-reference.txt holds 3000 random and boundary sizes for
// FileSize ("F <bytes> <text>") and the network size ("N <megabytes> <text>").
func TestAgainstJavaReference(t *testing.T) {
	path := "../../testdata/java-reference.txt"
	f, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	sc := bufio.NewScanner(f)
	n := 0
	for sc.Scan() {
		kind, rest, _ := strings.Cut(sc.Text(), " ")
		arg, want, _ := strings.Cut(rest, " ")
		var got string
		switch kind {
		case "F":
			v, _ := strconv.ParseInt(arg, 10, 64)
			got = FileSize(v)
		case "N":
			v, _ := strconv.ParseFloat(arg, 64)
			// Java renders values below 2 MB and from 2^40 MB on wrongly: Double.toString
			// switches to exponent notation there and the digits are cut at the wrong place.
			if v < 2 || v >= 1<<40 {
				continue
			}
			got = NetworkShareSize(v, FactorAuto)
		}
		if got != want {
			t.Errorf("%s %s: go %q, java %q", kind, arg, got, want)
		}
		n++
	}
	if n < 5000 {
		t.Errorf("only %d reference lines", n)
	}
}
