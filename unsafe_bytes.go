package omnitoken

import "unsafe"

func unsafeStringBytes(s string) []byte {
	if len(s) == 0 {
		return nil
	}
	return unsafe.Slice(unsafe.StringData(s), len(s))
}

func unsafeBytesString(b []byte) string {
	if len(b) == 0 {
		return ""
	}
	// The caller must keep b immutable for the lifetime of the returned string.
	return unsafe.String(unsafe.SliceData(b), len(b))
}
