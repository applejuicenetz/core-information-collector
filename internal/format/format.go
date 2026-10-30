// Package format renders the values of the placeholders. Separators are fixed to the
// German convention (comma decimal mark, dot grouping), independent of the host locale.
package format

import (
	"math/bits"
	"strconv"
	"strings"
)

// FileSize renders a byte count with binary units: "0" up to zero, then B, kB, MB, GB, TB or
// PB with at most two decimals ("1,5GB", "1.024kB" is never produced below 1 MB).
func FileSize(size int64) string {
	if size <= 0 {
		return "0"
	}
	units := [...]string{"B", "kB", "MB", "GB", "TB", "PB"}
	group := (bits.Len64(uint64(size)) - 1) / 10
	if group >= len(units) {
		group = len(units) - 1
	}
	return decimal(float64(size)/float64(uint64(1)<<(10*uint(group)))) + units[group]
}

// decimal rounds half-even to two decimals, drops trailing zeros and groups thousands.
func decimal(v float64) string {
	s := strconv.FormatFloat(v, 'f', 2, 64)
	whole, frac, _ := strings.Cut(s, ".")
	frac = strings.TrimRight(frac, "0")
	out := group(whole)
	if frac != "" {
		out += "," + frac
	}
	return out
}

// Integer renders n with dots as thousands separators ("3.182.468").
func Integer(n int64) string {
	return group(strconv.FormatInt(n, 10))
}

func group(digits string) string {
	sign := ""
	if strings.HasPrefix(digits, "-") {
		sign, digits = "-", digits[1:]
	}
	if len(digits) <= 3 {
		return sign + digits
	}
	var b strings.Builder
	first := len(digits) % 3
	if first > 0 {
		b.WriteString(digits[:first])
	}
	for i := first; i < len(digits); i += 3 {
		if b.Len() > 0 {
			b.WriteByte('.')
		}
		b.WriteString(digits[i : i+3])
	}
	return sign + b.String()
}

// Units of NetworkShareSize: the factor is the divisor applied to a value given in MB.
const (
	FactorAuto = 0
	FactorMB   = 1
	FactorGB   = 1024
	FactorTB   = 1024 * 1024
	FactorPB   = 1024 * 1024 * 1024
)

// NetworkShareSize renders the network share size, which the Core reports in MB. With
// FactorAuto the largest unit that keeps the number below 1024 is used. The value is cut
// (not rounded) to two decimals, trailing digits are kept as in "2,0GB" and "1,5TB". A factor
// that is not one of the constants renders the unit "??".
func NetworkShareSize(share float64, factor int64) string {
	if share == 0 {
		return "0,00 MB"
	}
	if factor == FactorAuto {
		switch {
		case share/1024 < 1024:
			factor = FactorGB
		case share/1048576 < 1024:
			factor = FactorTB
		case share/1073741824 < 1024:
			factor = FactorPB
		default:
			factor = FactorMB
		}
	}
	text := strconv.FormatFloat(share/float64(factor), 'f', -1, 64)
	whole, frac, found := strings.Cut(text, ".")
	if !found {
		frac = "0"
	}
	if len(frac) > 2 {
		frac = frac[:2]
	}
	unit := "??"
	switch factor {
	case FactorMB:
		unit = "MB"
	case FactorGB:
		unit = "GB"
	case FactorTB:
		unit = "TB"
	case FactorPB:
		unit = "PB"
	}
	return whole + "," + frac + unit
}
