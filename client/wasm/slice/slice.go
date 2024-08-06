package byteslice

import "unsafe"

func String2ByteSlice(str string, maxL int) []byte {
	if str == "" || maxL == 0 {
		return nil
	}

	if len(str) > maxL {
		str = str[:maxL]
	}

	return unsafe.Slice(unsafe.StringData(str), len(str))
}

func ByteSlice2String(bs []byte, maxL int) string {
	if len(bs) == 0 || maxL == 0 {
		return ""
	}

	if len(bs) > maxL {
		bs = bs[:maxL]
	}

	return unsafe.String(unsafe.SliceData(bs), len(bs))
}
