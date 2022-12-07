package parse

func IsAuth(val byte) bool {
	return val == 133 || val == 15
}
