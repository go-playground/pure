package pure

import (
	"mime"
	"path/filepath"
)

func min(a, b int) int {

	if a <= b {
		return a
	}
	return b
}

func countParams(path string) uint8 {

	var n uint // add one just as a buffer

	for i := 0; i < len(path); i++ {
		if path[i] == paramByte || path[i] == wildByte {
			n++
		}
	}

	if n >= 255 {
		panic("too many parameters defined in path, max is 255")
	}

	return uint8(n)
}

func detectContentType(filename string) (t string) {
	if t = mime.TypeByExtension(filepath.Ext(filename)); t == "" {
		t = OctetStream
	}
	return
}
