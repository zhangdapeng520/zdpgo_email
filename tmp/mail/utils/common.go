package utils

import (
	"golang.org/x/text/encoding/simplifiedchinese"
	"strings"
	"unicode/utf8"
)

type CommonError struct{
	Cause string
}

func (e CommonError) Error() string {
	return e.Cause
}

func IsUTF8String(s string)bool{
	return utf8.ValidString(s)
}

func IsUTF8(data []byte) bool {
	return utf8.Valid(data)
}

func IsGBK(data []byte) bool {
	if IsUTF8(data){
		return false
	}
	length := len(data)
	var i = 0
	for i < length {
		if data[i] <= 0xff {
			//编码小于等于127,只有一个字节的编码，兼容ASCII吗
			i++
			continue
		} else {
			//大于127的使用双字节编码
			if  data[i] >= 0x81 &&
				data[i] <= 0xfe &&
				data[i + 1] >= 0x40 &&
				data[i + 1] <= 0xfe &&
				data[i + 1] != 0xf7 {
				i += 2
				continue
			} else {
				return false
			}
		}
	}
	return true
}

// Convert gbk等转为utf-8 bytes
func ConvertToUTF8(b []byte, charset string) ([]byte, error) {
	charset = strings.ToLower(charset)
	switch charset {
	case "gb18030":
		decodeBytes,err := simplifiedchinese.GB18030.NewDecoder().Bytes(b)
		if err != nil{
			return b,err
		}
		return decodeBytes,err
	// GBK包含的汉字数量比GB2312和BIG5多，gb2312只包括简体汉字
	case "gbk","gb2312","big5":
		decodeBytes,err := simplifiedchinese.GBK.NewDecoder().Bytes(b)
		if err != nil{
			return b,err
		}
		return decodeBytes,nil
	}
	return b,CommonError{Cause: "不支持的编码转换"}
}