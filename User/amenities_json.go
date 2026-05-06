package user

import (
	"encoding/json"
	"strings"

	"gorm.io/datatypes"
)

// AmenitiesDisplayString turns DB JSON (string or []string) into a single display string for APIs.
func AmenitiesDisplayString(j datatypes.JSON) string {
	if len(j) == 0 {
		return ""
	}
	var s string
	if err := json.Unmarshal(j, &s); err == nil {
		return s
	}
	var arr []string
	if err := json.Unmarshal(j, &arr); err == nil {
		return strings.Join(arr, ", ")
	}
	return string(j)
}
