package core

import (
	"fmt"
	"strings"
)

// NormalizeUGPhoneNumber validates Uganda phone numbers and returns local UG format.
// Accepted examples:
// - 07XXXXXXXX
// - 03XXXXXXXX
// - 2567XXXXXXXX
// - 2563XXXXXXXX
// - +2567XXXXXXXX
// - +2563XXXXXXXX
func NormalizeUGPhoneNumber(raw string) (string, error) {
	sanitized := strings.TrimSpace(raw)
	sanitized = strings.ReplaceAll(sanitized, " ", "")
	sanitized = strings.ReplaceAll(sanitized, "-", "")

	if sanitized == "" {
		return "", fmt.Errorf("phone number is required")
	}

	// +2567XXXXXXXX / +2563XXXXXXXX => 13 chars
	if strings.HasPrefix(sanitized, "+256") && len(sanitized) == 13 {
		if sanitized[4] == '7' || sanitized[4] == '3' {
			return "0" + sanitized[4:], nil
		}
	}

	// 2567XXXXXXXX / 2563XXXXXXXX => 12 chars
	if strings.HasPrefix(sanitized, "256") && len(sanitized) == 12 {
		if sanitized[3] == '7' || sanitized[3] == '3' {
			return "0" + sanitized[3:], nil
		}
	}

	// 07XXXXXXXX / 03XXXXXXXX => 10 chars
	if len(sanitized) == 10 && sanitized[0] == '0' {
		if sanitized[1] == '7' || sanitized[1] == '3' {
			return "+256" + sanitized[1:], nil
		}
	}

	return "", fmt.Errorf("invalid Uganda phone number format")
}

func UGPhoneCandidates(normalized string) []string {
	if len(normalized) == 10 && normalized[0] == '0' && (normalized[1] == '7' || normalized[1] == '3') {
		withCountry := "256" + normalized[1:]
		withPlus := "+" + withCountry
		return []string{normalized, withCountry, withPlus}
	}

	if !strings.HasPrefix(normalized, "+256") || len(normalized) != 13 {
		return []string{normalized}
	}
	noPlus := strings.TrimPrefix(normalized, "+")
	local := "0" + normalized[4:]
	return []string{normalized, noPlus, local}
}

