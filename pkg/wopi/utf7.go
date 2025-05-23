// Copyright 2013 The Go-IMAP Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

/*
This package modified from:
https://github.com/mxk/go-imap/blob/master/imap/utf7.go
https://github.com/mxk/go-imap/blob/master/imap/utf7_test.go
IMAP specification uses modified UTF-7. Following are the differences:
 1. Printable US-ASCII except & (0x20 to 0x25 and 0x27 to 0x7e) MUST represent by themselves.
 2. '&' is used to shift modified BASE64 instead of '+'.
 3. Can NOT use superfluous null shift (&...-&...- should be just &......-).
 4. ',' is used in BASE64 code instead of '/'.
 5. '&' is represented '&-'. You can have many '&-&-&-&-'.
 6. No implicit shift from BASE64 to US-ASCII. All BASE64 must end with '-'.

Actual UTF-7 specification:
Rule 1: direct characters: 62 alphanumeric characters and 9 symbols: ' ( ) , - . / : ?
Rule 2: optional direct characters: all other printable characters in the range
U+0020–U+007E except ~ \ + and space. Plus sign (+) may be encoded as +-
(special case). Plus sign (+) mean the start of 'modified Base64 encoded UTF-16'.
The end of this block is indicated by any character not in the modified Base64.
If character after modified Base64 is a '-' then it is consumed.

Example:

	"1 + 1 = 2" is encoded as "1 +- 1 +AD0 2" //+AD0 is the '=' sign.
	"£1" is encoded as "+AKM-1" //+AKM- is the '£' sign where '-' is consumed.

A "+" character followed immediately by any character other than members
of modified Base64 or "-" is an ill-formed sequence. Convert to Unicode code
point then apply modified BASE64 (rfc2045) to it. Modified BASE64 do not use
padding instead add extra bits. Lines should never be broken in the middle of
a UTF-7 shifted sequence. Rule 3: Space, tab, carriage return and line feed may
also be represented directly as single ASCII bytes. Further content transfer
encoding may be needed if using in email environment.
*/
package wopi

import (
	"bytes"
	"encoding/base64"
	"errors"
	"io/ioutil"
	"unicode/utf16"
	"unicode/utf8"

	"golang.org/x/text/encoding"
	"golang.org/x/text/transform"
)

const (
	uRepl = '\uFFFD' // Unicode replacement code point
	u7min = 0x20     // Minimum self-representing UTF-7 value
	u7max = 0x7E     // Maximum self-representing UTF-7 value
)

// copy from golang.org/x/text/encoding/internal
type simpleEncoding struct {
	Decoder transform.Transformer
	Encoder transform.Transformer
}

func (e *simpleEncoding) NewDecoder() *encoding.Decoder {
	return &encoding.Decoder{Transformer: e.Decoder}
}

func (e *simpleEncoding) NewEncoder() *encoding.Encoder {
	return &encoding.Encoder{Transformer: e.Encoder}
}

var (
	UTF7 encoding.Encoding = &simpleEncoding{
		utf7Decoder{},
		utf7Encoder{},
	}
)

// ErrBadUTF7 is returned to indicate invalid modified UTF-7 encoding.
var ErrBadUTF7 = errors.New("utf7: bad utf-7 encoding")

// Base64 codec for code points outside of the 0x20-0x7E range.
const modifiedbase64 = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/"

var u7enc = base64.NewEncoding(modifiedbase64)

func isModifiedBase64(r byte) bool {
	if r >= 'A' && r <= 'Z' {
		return true
	} else if r >= 'a' && r <= 'z' {
		return true
	} else if r >= '0' && r <= '9' {
		return true
	} else if r == '+' || r == '/' {
		return true
	}
	return false
	// bs := []byte(modifiedbase64)
	// for _, b := range bs {
	// 	if b == r {
	// 		return true
	// 	}
	// }
	// return false
}

type utf7Decoder struct {
	transform.NopResetter
}

func (d utf7Decoder) Transform(dst, src []byte, atEOF bool) (nDst, nSrc int, err error) {
	var implicit bool
	var tmp int

	nd, n := len(dst), len(src)
	if n == 0 && !atEOF {
		return 0, 0, transform.ErrShortSrc
	}
	for ; nSrc < n; nSrc++ {
		if nDst >= nd {
			return nDst, nSrc, transform.ErrShortDst
		}
		if c := src[nSrc]; ((c < u7min || c > u7max) &&
			c != '\t' && c != '\r' && c != '\n') ||
			c == '~' || c == '\\' {
			return nDst, nSrc, ErrBadUTF7 // Illegal code point in ASCII mode
		} else if c != '+' {
			dst[nDst] = c // character can self represent
			nDst++
			continue
		}
		// found '+'
		start := nSrc + 1
		tmp = nSrc // nSrc remain pointing to '+', tmp point to end of BASE64
		// Find the end of the Base64 or "+-" segment
		implicit = false
		for tmp++; tmp < n && src[tmp] != '-'; tmp++ {
			if !isModifiedBase64(src[tmp]) {
				if tmp == start {
					return nDst, tmp, ErrBadUTF7 // '+' next char must modified base64
				}
				// implicit shift back to ASCII - no need '-' character
				implicit = true
				break
			}
		}
		if tmp == start {
			if tmp == n {
				// did not find '-' sign and '+' is last character
				// total nSrc no include '+'
				if atEOF {
					return nDst, nSrc, ErrBadUTF7 // '+' can not at the end
				}
				// '+' can not at the end, so get more data
				return nDst, nSrc, transform.ErrShortSrc
			}
			dst[nDst] = '+' // Escape sequence "+-"
			nDst++
		} else if tmp == n && !atEOF {
			// no end of BASE64 marker and still has data
			// probably the marker at next block of data
			// so go get more data.
			return nDst, nSrc, transform.ErrShortSrc
		} else if b := utf7dec(src[start:tmp]); len(b) > 0 {
			if len(b)+nDst > nd {
				// need more space on dst for the decoded modified BASE64 unicode
				// total nSrc no include '+'
				return nDst, nSrc, transform.ErrShortDst
			}
			copy(dst[nDst:], b) // Control or non-ASCII code points in Base64
			nDst += len(b)
			if implicit {
				if nDst >= nd {
					return nDst, tmp, transform.ErrShortDst
				}
				dst[nDst] = src[tmp] // implicit shift
				nDst++
			}
			if tmp == n {
				return nDst, tmp, nil
			}
		} else {
			return nDst, nSrc, ErrBadUTF7 // bad encoding
		}
		nSrc = tmp
	}
	return
}

type utf7Encoder struct {
	transform.NopResetter
}

func calcExpectedSize(runeSize int) (round int) {
	numerator := runeSize * 17
	round = numerator / 12
	remain := numerator % 12
	if remain >= 6 {
		round++
	}
	return
}

func (e utf7Encoder) Transform(dst, src []byte, atEOF bool) (nDst, nSrc int, err error) {
	var c byte
	var b []byte
	var endminus, needMoreSrc, needMoreDst, foundASCII, hasRuneStart bool
	var tmp, compare, lastRuneStart int
	var currentSize, maxRuneStart int
	var rn rune

	nd, n := len(dst), len(src)
	if n == 0 {
		if !atEOF {
			return 0, 0, transform.ErrShortSrc
		} else {
			return 0, 0, nil
		}
	}
	for nSrc = 0; nSrc < n; {
		if nDst >= nd {
			return nDst, nSrc, transform.ErrShortDst
		}
		c = src[nSrc]
		if canSelf(c) {
			nSrc++
			dst[nDst] = c
			nDst++
			continue
		} else if c == '+' {
			if nDst+2 > nd {
				return nDst, nSrc, transform.ErrShortDst
			}
			nSrc++
			dst[nDst], dst[nDst+1] = '+', '-'
			nDst += 2
			continue
		}
		start := nSrc
		tmp = nSrc // nSrc still point to first non-ASCII
		currentSize = 0
		maxRuneStart = nSrc
		needMoreDst = false
		if utf8.RuneStart(src[nSrc]) {
			hasRuneStart = true
		} else {
			hasRuneStart = false
		}
		foundASCII = true
		for tmp++; tmp < n && !canSelf(src[tmp]) && src[tmp] != '+'; tmp++ {
			// if next printable ASCII code point found the loop stop
			if utf8.RuneStart(src[tmp]) {
				hasRuneStart = true
				lastRuneStart = tmp
				rn, _ = utf8.DecodeRune(src[maxRuneStart:tmp])
				if rn >= 0x10000 {
					currentSize += 4
				} else {
					currentSize += 2
				}
				if calcExpectedSize(currentSize)+2 > nd-nDst {
					needMoreDst = true
				} else {
					maxRuneStart = tmp
				}
			}
		}

		// following to adjust tmp to right pointer as now tmp can not
		// find any good ending (searching end with no result). Adjustment
		// base on another earlier feasible valid rune position.
		needMoreSrc = false
		if tmp == n {
			foundASCII = false
			if !atEOF {
				if !hasRuneStart {
					return nDst, nSrc, transform.ErrShortSrc
				} else {
					//re-adjust tmp to good position to encode
					if !utf8.Valid(src[maxRuneStart:]) {
						if maxRuneStart == start {
							return nDst, nSrc, transform.ErrShortSrc
						}
						needMoreSrc = true
						tmp = maxRuneStart
					}
				}
			}
		}

		endminus = false
		if hasRuneStart && !needMoreSrc {
			// need check if dst enough buffer for transform
			rn, _ = utf8.DecodeRune(src[lastRuneStart:tmp])
			if rn >= 0x10000 {
				currentSize += 4
			} else {
				currentSize += 2
			}
			if calcExpectedSize(currentSize)+2 > nd-nDst {
				// can not use tmp value as transofrmed size too
				// big for dst
				endminus = true
				needMoreDst = true
				tmp = maxRuneStart
			}
		}

		b = utf7enc(src[start:tmp])
		if len(b) < 2 || b[0] != '+' {
			return nDst, nSrc, ErrBadUTF7 // Illegal code point in ASCII mode
		}

		if foundASCII {
			// printable ASCII found - check if BASE64 type
			if isModifiedBase64(src[tmp]) || src[tmp] == '-' {
				endminus = true
			}
		} else {
			endminus = true
		}
		compare = nDst + len(b)
		if endminus {
			compare++
		}
		if compare > nd {
			return nDst, nSrc, transform.ErrShortDst
		}
		copy(dst[nDst:], b)
		nDst += len(b)
		if endminus {
			dst[nDst] = '-'
			nDst++
		}
		nSrc = tmp

		if needMoreDst {
			return nDst, nSrc, transform.ErrShortDst
		}

		if needMoreSrc {
			return nDst, nSrc, transform.ErrShortSrc
		}
	}
	return
}

// UTF7Encode converts a string from UTF-8 encoding to modified UTF-7. This
// encoding is used by the Mailbox International Naming Convention (RFC 3501
// section 5.1.3). Invalid UTF-8 byte sequences are replaced by the Unicode
// replacement code point (U+FFFD).
func UTF7Encode(s string) string {
	return string(UTF7EncodeBytes([]byte(s)))
}

const (
	setD = iota
	setO
	setRule3
	setInvalid
)

// get the set of characters group.
func getSetType(c byte) int {
	if (c >= 44 && c <= ':') || c == '?' {
		return setD
	} else if c == 39 || c == '(' || c == ')' {
		return setD
	} else if c >= 'A' && c <= 'Z' {
		return setD
	} else if c >= 'a' && c <= 'z' {
		return setD
	} else if c == '+' || c == '\\' {
		return setInvalid
	} else if c > ' ' && c < '~' {
		return setO
	} else if c == ' ' || c == '\t' ||
		c == '\r' || c == '\n' {
		return setRule3
	}
	return setInvalid
}

// Check if can represent by themselves.
func canSelf(c byte) bool {
	t := getSetType(c)
	if t == setInvalid {
		return false
	}
	return true
}

// UTF7EncodeBytes converts a byte slice from UTF-8 encoding to modified UTF-7.
func UTF7EncodeBytes(s []byte) []byte {
	input := bytes.NewReader(s)
	reader := transform.NewReader(input, UTF7.NewEncoder())
	output, err := ioutil.ReadAll(reader)
	if err != nil {
		return nil
	}
	return output
}

// utf7enc converts string s from UTF-8 to UTF-16-BE, encodes the result as
// Base64, removes the padding, and adds UTF-7 shifts.
func utf7enc(s []byte) []byte {
	// len(s) is sufficient for UTF-8 to UTF-16 conversion if there are no
	// control code points (see table below).
	b := make([]byte, 0, len(s)+4)
	for len(s) > 0 {
		r, size := utf8.DecodeRune(s)
		if r > utf8.MaxRune {
			r, size = utf8.RuneError, 1 // Bug fix (issue 3785)
		}
		s = s[size:]
		if r1, r2 := utf16.EncodeRune(r); r1 != uRepl {
			//log.Println("surrogate triggered")
			b = append(b, byte(r1>>8), byte(r1))
			r = r2
		}
		b = append(b, byte(r>>8), byte(r))
	}

	// Encode as Base64
	//n := u7enc.EncodedLen(len(b)) + 2 // plus 2 for prefix '+' and suffix '-'
	n := u7enc.EncodedLen(len(b)) + 1 // plus for prefix '+'
	b64 := make([]byte, n)
	u7enc.Encode(b64[1:], b)

	// Strip padding
	n -= 2 - (len(b)+2)%3
	b64 = b64[:n]

	// Add UTF-7 shifts
	b64[0] = '+'
	//b64[n-1] = '-'
	return b64
}

// UTF7Decode converts a string from modified UTF-7 encoding to UTF-8.
func UTF7Decode(u string) (s string, err error) {
	b, err := UTF7DecodeBytes([]byte(u))
	s = string(b)
	return
}

// UTF7DecodeBytes converts a byte slice from modified UTF-7 encoding to UTF-8.
func UTF7DecodeBytes(u []byte) ([]byte, error) {
	input := bytes.NewReader([]byte(u))
	reader := transform.NewReader(input, UTF7.NewDecoder())
	output, err := ioutil.ReadAll(reader)
	if err != nil {
		return nil, err
	}
	return output, nil
}

// utf7dec extracts UTF-16-BE bytes from Base64 data and converts them to UTF-8.
// A nil slice is returned if the encoding is invalid.
func utf7dec(b64 []byte) []byte {
	var b []byte

	// Allocate a single block of memory large enough to store the Base64 data
	// (if padding is required), UTF-16-BE bytes, and decoded UTF-8 bytes.
	// Since a 2-byte UTF-16 sequence may expand into a 3-byte UTF-8 sequence,
	// double the space allocation for UTF-8.
	if n := len(b64); b64[n-1] == '=' {
		return nil
	} else if n&3 == 0 {
		b = make([]byte, u7enc.DecodedLen(n)*3)
	} else {
		n += 4 - n&3
		b = make([]byte, n+u7enc.DecodedLen(n)*3)
		copy(b[copy(b, b64):n], []byte("=="))
		b64, b = b[:n], b[n:]
	}

	// Decode Base64 into the first 1/3rd of b
	n, err := u7enc.Decode(b, b64)
	if err != nil || n&1 == 1 {
		return nil
	}

	// Decode UTF-16-BE into the remaining 2/3rds of b
	b, s := b[:n], b[n:]
	j := 0
	for i := 0; i < n; i += 2 {
		r := rune(b[i])<<8 | rune(b[i+1])
		if utf16.IsSurrogate(r) {
			if i += 2; i == n {
				//log.Println("surrogate error1!")
				return nil
			}
			r2 := rune(b[i])<<8 | rune(b[i+1])
			//log.Printf("surrogate! 0x%04X 0x%04X\n", r, r2)
			if r = utf16.DecodeRune(r, r2); r == uRepl {
				return nil
			}
		}
		j += utf8.EncodeRune(s[j:], r)
	}
	return s[:j]
}

/*
The following table shows the number of bytes required to encode each code point
in the specified range using UTF-8 and UTF-16 representations:

+-----------------+-------+--------+
| Code points     | UTF-8 | UTF-16 |
+-----------------+-------+--------+
| 000000 - 00007F |   1   |   2    |
| 000080 - 0007FF |   2   |   2    |
| 000800 - 00FFFF |   3   |   2    |
| 010000 - 10FFFF |   4   |   4    |
+-----------------+-------+--------+

Source: http://en.wikipedia.org/wiki/Comparison_of_Unicode_encodings
*/
