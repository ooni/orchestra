package common

import "strings"

// MapToUppercase returns the list with all the strings uppercased
func MapToUppercase(vs []string) []string {
	vso := make([]string, len(vs))
	for i, v := range vs {
		vso[i] = strings.ToUpper(v)
	}
	return vso
}

// ValidateCSVMapStr Check if a CSV map string contains all values inside of a target mapStrStruct
func ValidateCSVMapStr(csvStr string, m mapStrStruct) bool {
	if csvStr == "" {
		// It's ok if the value is missing
		return true
	}
	for _, v := range strings.Split(csvStr, ",") {
		_, present := m[strings.ToUpper(v)]
		if !present {
			return false
		}
	}
	return true
}
