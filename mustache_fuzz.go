//go:build gofuzz
// +build gofuzz

package mustache

// Fuzzing code for use with github.com/dvyukov/go-fuzz
//
// To use, in the main project directory do:
//
//   go get -u github.com/dvyukov/go-fuzz/go-fuzz github.com/dvyukov/go-fuzz/go-fuzz-build
//   go-fuzz-build
//   go-fuzz

func Fuzz(data []byte) int {
	_, err := ParseString(string(data))
	if err == nil {
		return 1
	}
	return 0
}
