package zdpgo_email

import (
	"bytes"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"strings"
	"unicode"
	"unicode/utf8"
)

// 重写mime/encodedword.go的DecodeHeader方法

// A WordDecoder decodes MIME headers containing RFC 2047 encoded-words.
type WordDecoder struct {
	// CharsetReader, if non-nil, defines a function to generate
	// charset-conversion readers, converting from the provided
	// charset into UTF-8.
	// Charsets are always lower-case. utf-8, iso-8859-1 and us-ascii charsets
	// are handled by default.
	// One of the CharsetReader's result values must be non-nil.
	CharsetReader func(charset string, input io.Reader) (io.Reader, error)
}

var (
	errInvalidWord = errors.New("mime: invalid RFC 2047 encoded-word")
)

// DecodeHeader decodes all encoded-words of the given string. It returns an
// error if and only if CharsetReader of d returns an error.
func (d *WordDecoder) DecodeHeader(header string) (string, error) {
	// If there is no encoded-word, returns before creating a buffer.
	i := strings.Index(header, "=?")
	if i == -1 {
		return header, nil
	}

	var buf strings.Builder

	buf.WriteString(header[:i])
	header = header[i:]

	betweenWords := false
	for {
		start := strings.Index(header, "=?")
		if start == -1 {
			break
		}
		cur := start + len("=?")

		i := strings.Index(header[cur:], "?")
		if i == -1 {
			break
		}
		charset := header[cur : cur+i]
		cur += i + len("?")

		if len(header) < cur+len("Q??=") {
			break
		}
		encoding := header[cur]
		cur++

		if header[cur] != '?' {
			break
		}
		cur++

		j := strings.Index(header[cur:], "?=")
		if j == -1 {
			break
		}
		text := header[cur : cur+j]
		end := cur + j + len("?=")

		content, err := Decode(encoding, text)
		if err != nil {
			betweenWords = false
			buf.WriteString(header[:start+2])
			header = header[start+2:]
			continue
		}

		// Write characters before the encoded-word. White-space and newline
		// characters separating two encoded-words must be deleted.
		if start > 0 && (!betweenWords || hasNonWhitespace(header[:start])) {
			buf.WriteString(header[:start])
		}

		if err := d.Convert(&buf, charset, content); err != nil {
			return "", err
		}

		header = header[end:]
		betweenWords = true
	}

	if len(header) > 0 {
		buf.WriteString(header)
	}

	return buf.String(), nil
}

func Decode(encoding byte, text string) ([]byte, error) {
	switch encoding {
	case 'B', 'b':
		r, err := base64.StdEncoding.DecodeString(text)
		//这个报错可忽略
		if strings.HasPrefix(fmt.Sprintf("%s", err), "illegal base64 data at input byte") {
			return r, nil
		}
		return r, err
	case 'Q', 'q':
		return QDecode(text)
	default:
		return nil, errInvalidWord
	}
}

func (d *WordDecoder) Convert(buf *strings.Builder, charset string, content []byte) error {
	switch {
	case strings.EqualFold("utf-8", charset):
		buf.Write(content)
	case strings.EqualFold("iso-8859-1", charset):
		for _, c := range content {
			buf.WriteRune(rune(c))
		}
	case strings.EqualFold("us-ascii", charset):
		for _, c := range content {
			if c >= utf8.RuneSelf {
				buf.WriteRune(unicode.ReplacementChar)
			} else {
				buf.WriteByte(c)
			}
		}
	default:
		if d.CharsetReader == nil {
			return fmt.Errorf("mime: unhandled charset %q", charset)
		}
		r, err := d.CharsetReader(strings.ToLower(charset), bytes.NewReader(content))
		if err != nil {
			return err
		}
		if _, err = io.Copy(buf, r); err != nil {
			return err
		}
	}
	return nil
}

// hasNonWhitespace reports whether s (assumed to be ASCII) contains at least
// one byte of non-whitespace.
func hasNonWhitespace(s string) bool {
	for _, b := range s {
		switch b {
		// Encoded-words can only be separated by linear white spaces which does
		// not include vertical tabs (\v).
		case ' ', '\t', '\n', '\r':
		default:
			return true
		}
	}
	return false
}

// QDecode decodes a Q encoded string.
func QDecode(s string) ([]byte, error) {
	dec := make([]byte, len(s))
	n := 0
	for i := 0; i < len(s); i++ {
		switch c := s[i]; {
		case c == '_':
			dec[n] = ' '
		case c == '=':
			if i+2 >= len(s) {
				return nil, errInvalidWord
			}
			b, err := readHexByte(s[i+1], s[i+2])
			if err != nil {
				return nil, err
			}
			dec[n] = b
			i += 2
		case (c <= '~' && c >= ' ') || c == '\n' || c == '\r' || c == '\t':
			dec[n] = c
		default:
			return nil, errInvalidWord
		}
		n++
	}

	return dec[:n], nil
}

// readHexByte returns the byte from its quoted-printable representation.
func readHexByte(a, b byte) (byte, error) {
	var hb, lb byte
	var err error
	if hb, err = fromHex(a); err != nil {
		return 0, err
	}
	if lb, err = fromHex(b); err != nil {
		return 0, err
	}
	return hb<<4 | lb, nil
}

func fromHex(b byte) (byte, error) {
	switch {
	case b >= '0' && b <= '9':
		return b - '0', nil
	case b >= 'A' && b <= 'F':
		return b - 'A' + 10, nil
	// Accept badly encoded bytes.
	case b >= 'a' && b <= 'f':
		return b - 'a' + 10, nil
	}
	return 0, fmt.Errorf("mime: invalid hex byte %#02x", b)
}
