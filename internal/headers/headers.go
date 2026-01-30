package headers

import (
	"bytes"
	"fmt"
	"strings"
)

type Headers map[string]string

const CRLF = "\r\n"

func NewHeaders() (Headers) {
	return make(Headers)
}

func (h Headers) Get(fieldName string) string {
	return h[strings.ToLower(fieldName)]
}

func (h Headers) Set(fieldName string, value string) {
	key := strings.ToLower(fieldName)
	if _, found := h[key]; found {
		h[key] = fmt.Sprintf("%s, %s", h[key], value);
	} else {
		h[key] = value
	}
}

func (h Headers) Parse(data []byte) (n int, done bool, err error) {
	ind := bytes.Index(data, []byte(CRLF))
	if ind == 0 {
		return 2, true, nil
	}
	if ind == -1 {
		return 0, false, nil
	}

	sepIndex := bytes.IndexByte(data, ':')
	if sepIndex == -1 {
		return 0, false, fmt.Errorf("header does not contain separator")
	}
	if sepIndex == 0 {
		return 0, false, fmt.Errorf("header contains no field-name")
	}
	if data[sepIndex - 1] == ' ' {
		return 0, false, fmt.Errorf("whitespace not allowed between field-name and separator")
	}
	fieldName := data[:sepIndex]
	fieldName = bytes.TrimSpace(fieldName)
	if !isValidFieldName(fieldName) {
		return 0, false, fmt.Errorf("invalid character in field-name")
	}

	fieldValue := data[(sepIndex + 1):ind]
	fieldValue = bytes.TrimSpace(fieldValue)

	h.Set(string(fieldName), string(fieldValue))

  // if (bytes.Index(data[(ind + 2):], []byte(CRLF)) == 0) {
  //   return ind + 2, true, nil
  // }

  // fmt.Printf("found %s: %s\n", string(fieldName), string(fieldValue))

	return ind + 2, false, nil
}

func isValidFieldName(fieldName []byte) bool {
	var allowed = func() [256]bool {
		var a [256]bool
		for c := byte('A'); c <= 'Z'; c++ {
			a[c] = true
		}
		for c := byte('a'); c <= 'z'; c++ {
			a[c] = true
		}
		for c := byte('0'); c <= '9'; c++ {
			a[c] = true
		}
		for _, c := range []byte("!#$%&'*+-.^_`|~") {
			a[c] = true
		}
		return a
	}()

	for _, b := range fieldName {
		if !allowed[b] {
			return false
		}
	}
	return true
}