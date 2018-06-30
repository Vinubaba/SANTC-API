package api

func IsNilOrEmpty(value *string) bool {
	if value == nil {
		return true
	}
	return *value == ""
}
