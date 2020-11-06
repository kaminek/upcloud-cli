package ui

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"unicode"
)

type spec struct {
	th     uint
	suffix string
}

var siMultiples = [...]spec{
	{th: 1 * 1000 * 1000 * 1000 * 1000, suffix: "T"},
	{th: 1 * 1000 * 1000 * 1000, suffix: "G"},
	{th: 1 * 1000 * 1000, suffix: "M"},
	{th: 1 * 1000, suffix: "K"},
	{th: 1, suffix: ""},
}

var binaryMultiples = [...]spec{
	{th: 1 * 1024 * 1024 * 1024 * 1024, suffix: "Ti"},
	{th: 1 * 1024 * 1024 * 1024, suffix: "Gi"},
	{th: 1 * 1024 * 1024, suffix: "Mi"},
	{th: 1 * 1024, suffix: "Ki"},
	{th: 1, suffix: ""},
}

func abbrevInt(raw uint, specs [5]spec) string {
	var (
		val    float64
		suffix string
	)
	for _, spec := range specs {
		if spec.th > raw {
			continue
		}
		if float64(raw)/float64(spec.th) < 0 {
			continue
		}
		val = float64(raw) / float64(spec.th)
		suffix = spec.suffix
		break
	}
	if _, frac := math.Modf(val); frac > 0 {
		return fmt.Sprintf("%.2f%s", val, suffix)
	} else {
		return fmt.Sprintf("%.0f%s", val, suffix)
	}
}

func parseAbbrevInt(s string, specs [5]spec) (uint, error) {
	if len(s) == 0 {
		return 0, fmt.Errorf("invalid value %q", s)
	}
	for _, spec := range specs {
		if strings.HasSuffix(s, spec.suffix) {
			if spec.suffix == "" && !unicode.IsDigit(rune(s[len(s)-1])) {
				s = s[0 : len(s)-1]
			}
			v, err := strconv.ParseFloat(strings.TrimRight(s, spec.suffix), 64)
			if err != nil {
				return 0, err
			}
			return uint(v * float64(spec.th)), nil
		}
	}
	return 0, fmt.Errorf("invalid value %q", s)
}

func AbbrevNum(raw uint) string {
	return abbrevInt(raw, siMultiples)
}

func AbbrevNumBinaryPrefix(raw uint) string {
	return abbrevInt(raw, binaryMultiples)
}

func FormatBytes(n int) string {
	return fmt.Sprintf("%sB", AbbrevNumBinaryPrefix(uint(n)))
}

func ParseAbbrevNum(s string) (uint, error) {
	return parseAbbrevInt(s, siMultiples)
}

func ParseAbbrevNumBinaryPrefix(s string) (uint, error) {
	return parseAbbrevInt(s, binaryMultiples)
}
