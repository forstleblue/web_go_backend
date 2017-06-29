package utils

//PosIntOnly extracts a positive integer from a string use: n := strings.TrimFunc(s, PosIntOnly)
func PosIntOnly(r int) bool {
	return r < '0' || '9' < r
}
