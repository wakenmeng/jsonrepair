package jsonrepair

import (
	"regexp"
	"strings"
)

const (
	codeBackslash               = 0x5c   // "\"
	codeSlash                   = 0x2f   // "/"
	codeAsterisk                = 0x2a   // "*"
	codeOpeningBrace            = 0x7b   // "{"
	codeClosingBrace            = 0x7d   // "}"
	codeOpeningBracket          = 0x5b   // "["
	codeClosingBracket          = 0x5d   // "]"
	codeOpenParenthesis         = 0x28   // "("
	codeCloseParenthesis        = 0x29   // ")"
	codeSpace                   = 0x20   // " "
	codeNewline                 = 0xa    // "\n"
	codeTab                     = 0x9    // "\t"
	codeReturn                  = 0xd    // "\r"
	codeBackspace               = 0x08   // "\b"
	codeFormFeed                = 0x0c   // "\f"
	codeDoubleQuote             = 0x0022 // "
	codePlus                    = 0x2b   // "+"
	codeMinus                   = 0x2d   // "-"
	codeQuote                   = 0x27   // "'"
	codeZero                    = 0x30   // 0
	codeNine                    = 0x39   // 9
	codeComma                   = 0x2c   // ","
	codeDot                     = 0x2e   // "." (dot, period)
	codeColon                   = 0x3a   // ":"
	codeSemicolon               = 0x3b   // ";"
	codeUppercaseA              = 0x41   // "A"
	codeLowercaseA              = 0x61   // "a"
	codeUppercaseE              = 0x45   // "E"
	codeLowercaseE              = 0x65   // "e"
	codeUppercaseF              = 0x46   // "F"
	codeLowercaseF              = 0x66   // "f"
	codeNonBreakingSpace        = 0xa0
	codeEnQuad                  = 0x2000
	codeHairSpace               = 0x200a
	codeNarrowNoBreakSpace      = 0x202f
	codeMediumMathematicalSpace = 0x205f
	codeIdeographicSpace        = 0x3000
	codeDoubleQuoteLeft         = 0x201c // “
	codeDoubleQuoteRight        = 0x201d // ”
	codeQuoteLeft               = 0x2018 // ‘
	codeQuoteRight              = 0x2019 // ’
	codeGraveAccent             = 0x0060 // `
	codeAcuteAccent             = 0x00b4 // ´
) // TODO: sort the codes

func IsHex(code rune) bool {
	return ((code >= codeZero && code <= codeNine) ||
		(code >= codeUppercaseA && code <= codeUppercaseF) ||
		(code >= codeLowercaseA && code <= codeLowercaseF))
}

func IsDigit(code rune) bool {
	return code >= codeZero && code <= codeNine
}

func IsValidStringCharacter(code rune) bool {
	return code >= 0x20 && code <= 0x10ffff
}

var Delimiters = map[rune]bool{
	',':  true,
	':':  true,
	'[':  true,
	']':  true,
	'{':  true,
	'}':  true,
	'(':  true,
	')':  true,
	'\n': true,
}

var regexDelimiter = regexp.MustCompile(`^[,:[\]{}()\n+]$`)

func IsDelimiter(c rune) bool {
	return regexDelimiter.Match([]byte{byte(c)}) || IsQuote(c)
	// return Delimiters[r] || (r != 0 && IsQuote(int(r)))
	//return regexDelimiter.test(char) || (char && IsQuote(char.charCodeAt(0)))
}

var regexStartOfValue = regexp.MustCompile(`^[[{\w-]$`)

var regexNumberWithLeadingZero = regexp.MustCompile(`^0\d`)

func IsStartOfValue(r rune) bool {
	return regexStartOfValue.Match([]byte{byte(r)}) || (r != 0 && IsQuote(r))
}

func IsControlCharacter(code rune) bool {
	return (code == codeNewline ||
		code == codeReturn ||
		code == codeTab ||
		code == codeBackspace ||
		code == codeFormFeed)
}

/**
 * Check if the given character is a whitespace character like space, tab, or
 * newline
 */
func IsWhitespace(code rune) bool {
	return code == codeSpace || code == codeNewline || code == codeTab || code == codeReturn
}

/**
 * Check if the given character is a special whitespace character, some
 * unicode variant
 */
func IsSpecialWhitespace(code rune) bool {
	return (code == codeNonBreakingSpace ||
		(code >= codeEnQuad && code <= codeHairSpace) ||
		code == codeNarrowNoBreakSpace ||
		code == codeMediumMathematicalSpace ||
		code == codeIdeographicSpace)
}

/**
 * Test whether the given character is a quote or double quote character.
 * Also tests for special variants of quotes.
 */
func IsQuote(code rune) bool {
	// the first check double quotes, since that occurs most often
	return IsDoubleQuoteLike(code) || IsSingleQuoteLike(code)
}

func IsDoubleQuoteLike(code rune) bool {
	// the first check double quotes, since that occurs most often
	return code == codeDoubleQuote || code == codeDoubleQuoteLeft || code == codeDoubleQuoteRight
}

/**
 * Test whether the given character is a double quote character.
 * Does NOT test for special variants of double quotes.
 */
func IsDoubleQuote(code rune) bool {
	return code == codeDoubleQuote
}

/**
 * Test whether the given character is a single quote character.
 * Also tests for special variants of single quotes.
 */
func IsSingleQuoteLike(code rune) bool {
	return (code == codeQuote ||
		code == codeQuoteLeft ||
		code == codeQuoteRight ||
		code == codeGraveAccent ||
		code == codeAcuteAccent)
}

/**
 * Test whether the given character is a single quote character.
 * Does NOT test for special variants of single quotes.
 */
func isSingleQuote(code rune) bool {
	return code == codeQuote
}

/**
 * Strip last occurrence of textToStrip from text
 */
func stripLastOccurrence(text, textToStrip string, stripRemainingText bool) string {
	index := strings.LastIndex(text, textToStrip)
	if index != -1 {
		if stripRemainingText {
			return text[:index]
		} else {
			return text[:index] + text[index+1:]
		}
	}
	return text
}

func InsertBeforeLastWhitespace(text []rune, textToInsert string) []rune {
	index := len(text)
	toInsert := []rune(textToInsert)

	if !IsWhitespace(text[index-1]) {
		// no trailing whitespaces
		text = append(text, toInsert...)
		return text
	}

	for IsWhitespace(text[index-1]) {
		index--
	}
	toInsert = append(toInsert, text[index:]...)
	text = append(text[:index], toInsert...)
	return text
}

func RemoveAtIndex(text []rune, start, count int) []rune {
	return append(text[:start], text[start+count:]...)
}

/**
 * Test whether a string ends with a newline or comma character and optional whitespace
 */

var endWithCommaOrNewlineReg = regexp.MustCompile(`[,\n][ \t\r]*$`)

func EndsWithCommaOrNewline(text string) bool {
	return endWithCommaOrNewlineReg.MatchString(text)
}

func nextNonWhiteSpaceCharacter(text []rune, start int) rune {
	var i = start
	for i < len(text) && IsWhitespace(text[i]) {
		i++
	}
	if i >= len(text) {
		return -1
	}
	return text[i]
}
