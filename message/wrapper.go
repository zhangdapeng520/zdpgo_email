package message

import (
	"io"
)

type writer struct {
	Len int

	sepBytes []byte
	w        io.Writer
	i        int
}

func (w *writer) Write(b []byte) (N int, err error) {
	to := w.Len - w.i

	for len(b) > to {
		var n int
		n, err = w.w.Write(b[:to])
		if err != nil {
			return
		}
		N += n
		b = b[to:]

		_, err = w.w.Write(w.sepBytes)
		if err != nil {
			return
		}

		w.i = 0
		to = w.Len
	}

	w.i += len(b)

	n, err := w.w.Write(b)
	if err != nil {
		return
	}
	N += n

	return
}

// NewWrapper 返回一个写入器，该写入器将其输入分成具有相同长度的多个部分，并在这些部分之间添加分隔符。
func NewWrapper(w io.Writer, sep string, l int) io.Writer {
	return &writer{
		Len:      l,
		sepBytes: []byte(sep),
		w:        w,
	}
}

// NewRFC822 创建一个RFC822文本包装器。它添加了一个CRLF(例如。\r\n)每个76个字符。
func NewRFC822(w io.Writer) io.Writer {
	return NewWrapper(w, "\r\n", 76)
}
