package utils

func ValidateKey(key string) bool {
	switch len(key) {
	case 16, 24, 32:
		return true
	default:
		return false
	}
}
