package utils

import (
	"strconv"
	"strings"
)

func ParseFee(feeText string) (float64, bool) {
	feeTextLower := strings.ToLower(strings.TrimSpace(feeText))
	isLoan := false

	switch {
	case feeTextLower == "" || feeTextLower == "-" || feeTextLower == "?":
		return 0, false
	case strings.Contains(feeTextLower, "free transfer"):
		return 0, false
	case strings.Contains(feeTextLower, "loan fee"):
		isLoan = true
		feeText = strings.ReplaceAll(feeTextLower, "loan fee:", "")
	case strings.Contains(feeTextLower, "loan transfer"):
		isLoan = true
		return 0, isLoan
	}

	feeText = strings.ReplaceAll(feeText, "€", "")
	feeText = strings.ReplaceAll(feeText, "$", "")
	feeText = strings.TrimSpace(feeText)

	multiplier := 1.0
	if strings.Contains(feeText, "th.") || strings.Contains(feeText, "k") {
		multiplier = 1_000
		feeText = strings.ReplaceAll(feeText, "th.", "")
		feeText = strings.ReplaceAll(feeText, "k", "")
	} else if strings.Contains(feeText, "m") {
		multiplier = 1_000_000
		feeText = strings.ReplaceAll(feeText, "m", "")
	} else if strings.Contains(feeText, "bn") || strings.Contains(feeText, "b") {
		multiplier = 1_000_000_000
		feeText = strings.ReplaceAll(feeText, "bn", "")
		feeText = strings.ReplaceAll(feeText, "b", "")
	}

	feeText = strings.ReplaceAll(feeText, ",", ".")
	feeText = strings.ReplaceAll(feeText, " ", "")
	if feeText == "" {
		return 0, isLoan
	}

	value, err := strconv.ParseFloat(feeText, 64)
	if err != nil {
		return 0, isLoan
	}

	return value * multiplier, isLoan
}
