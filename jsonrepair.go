package jsonrepair

import (
	"encoding/json"
	"fmt"
)

var (
	controlCharacters = map[string]string{
		"\b": "\\b",
		"\f": "\\f",
		"\n": "\\n",
		"\r": "\\r",
		"\t": "\\t",
	}

	// map with all escape characters
	escapeCharacters = map[string]string{
		`"`:  `"`,
		"\\": "\\",
		"/":  "/",
		"b":  "\b",
		"f":  "\f",
		"n":  "\n",
		"r":  "\r",
		"t":  "\t",
		// note that \u is handled separately in parseString()
	}
)

type (
	RepairText struct {
		text   []rune //string
		i      int
		output []rune
	}
)

func (t *RepairText) CharCode(i int) rune {
	if i >= len(t.text) {
		return -1
	}
	return t.text[i]
}

func (t *RepairText) Char(i int) string {
	if i >= len(t.text) {
		return ""
	}
	return string(t.text[i])
}

func (t *RepairText) Slice(a, b int) []rune {
	return t.text[a:min(b, len(t.text))]
}

func JSONRepair(text string) (string, error) {
	//i := 0

	t := RepairText{
		text:   []rune(text),
		i:      0,
		output: []rune{},
	}
	processedValue, err := t.parseValue()
	if err != nil {
		return "", err
	}
	if !processedValue {
		return "", UnexpectedEndError.At(len(t.text))
	}
	processedComma := t.parseCharacter(codeComma)
	if processedComma {
		t.parseWhitespaceAndSkipComments()
	}
	if t.i < len(t.text) && IsStartOfValue(t.text[t.i]) && EndsWithCommaOrNewline(string(t.output)) {
		if !processedComma {
			t.output = InsertBeforeLastWhitespace(t.output, ",")
		}
		t.parseNewlineDelimitedJSON()
	} else if processedComma {
		t.output = []rune(stripLastOccurrence(string(t.output), ",", false))
	}
	for t.CharCode(t.i) == codeClosingBrace || t.CharCode(t.i) == codeClosingBracket {
		t.i++
		t.parseWhitespaceAndSkipComments()
	}

	if t.i >= len(t.text) {
		return string(t.output), nil
	}
	return "", UnexpectedCharacterError.MessageAppend(fmt.Sprintf(`"%s"`, string(t.text[t.i]))).At(t.i)
}

func (t *RepairText) parseValue() (bool, error) {
	var processed bool
	var err error
	t.parseWhitespaceAndSkipComments()
	defer func() {
		if err == nil {
			t.parseWhitespaceAndSkipComments()
		}
	}()
	if processed, err = t.parseObject(); err != nil {
		return false, err
	} else if processed {
		return true, nil
	}
	if processed, err = t.parseArray(); err != nil {
		return false, err
	} else if processed {
		return true, nil
	}
	if processed, err = t.parseString(false); err != nil {
		return false, err
	} else if processed {
		return true, nil
	}
	if processed, err = t.parseNumber(); err != nil {
		return false, err
	} else if processed {
		return true, nil
	}
	if processed = t.parseKeywords(); processed {
		return true, nil
	}
	if processed, err = t.parseUnquotedString(); err != nil {
		return false, err
	} else if processed {
		return true, nil
	}
	//t.parseWhitespaceAndSkipComments()
	return processed, nil
}

func (t *RepairText) parseWhitespaceAndSkipComments() bool {
	start := t.i

	var changed bool
	changed = t.parseWhitespace()
	for {
		changed = t.parseComment()
		if changed {
			changed = t.parseWhitespace()
		} else {
			break
		}
	}
	return t.i > start
}

func (t *RepairText) parseComment() bool {

	if t.CharCode(t.i) == codeSlash && t.CharCode(t.i+1) == codeAsterisk {
		for t.i < len(t.text) && !t.atEndOfBlockComment() {
			t.i++
		}
		t.i += 2
		return true
	}

	if t.CharCode(t.i) == codeSlash && t.CharCode(t.i+1) == codeSlash {
		for t.i < len(t.text) && t.CharCode(t.i) != codeNewline {
			t.i++
		}
		return true
	}
	return false
}

func (t *RepairText) parseObject() (bool, error) {
	var err error
	if t.CharCode(t.i) == codeOpeningBrace {
		t.output = append(t.output, '{')
		t.i++
		t.parseWhitespaceAndSkipComments()

		var initial = true
		for t.i < len(t.text) && t.CharCode(t.i) != codeClosingBrace {
			var processedComma bool

			if !initial {
				processedComma = t.parseCharacter(codeComma)
				if !processedComma {
					t.output = InsertBeforeLastWhitespace(t.output, ",")
				}
				t.parseWhitespaceAndSkipComments()
			} else {
				processedComma = true
				initial = false
			}

			var processedKey bool
			processedKey, err = t.parseString(false)
			if err != nil {
				return false, err
			}
			if !processedKey {
				processedKey, err = t.parseUnquotedString()
				if err != nil {
					return false, err
				}
			}
			if !processedKey {
				chcode := t.CharCode(t.i)
				if chcode == codeClosingBrace || chcode == codeOpeningBrace ||
					chcode == codeClosingBracket || chcode == codeOpeningBracket ||
					t.i >= len(t.text) || t.i < 0 {
					t.output = []rune(stripLastOccurrence(string(t.output), ",", false))
				} else {
					return false, ObjectKeyExpectedError.At(t.i)
				}
				break
			}
			t.parseWhitespaceAndSkipComments()
			processedColon := t.parseCharacter(codeColon)
			truncatedtext := t.i >= len(t.text)
			if !processedColon {
				if truncatedtext || IsStartOfValue(t.text[t.i]) {
					t.output = InsertBeforeLastWhitespace(t.output, ":")
				} else {
					return false, ColonExpectedError.At(t.i)
				}
			}
			processedValue, err := t.parseValue()
			if err != nil {
				return false, err
			}
			if !processedValue {
				if truncatedtext || processedColon {
					t.output = append(t.output, []rune("null")...)
				} else {
					return false, ColonExpectedError.At(t.i)
				}
			}
		}
		if t.CharCode(t.i) == codeClosingBrace {
			t.output = append(t.output, '}')
			t.i++
		} else {
			t.output = InsertBeforeLastWhitespace(t.output, "}")
		}
		return true, nil
	}
	return false, nil
}

func (t *RepairText) parseArray() (bool, error) {
	if t.CharCode(t.i) == codeOpeningBracket {
		t.output = append(t.output, '[')
		t.i++
		t.parseWhitespaceAndSkipComments()

		initial := true
		for t.i < len(t.text) && t.CharCode(t.i) != codeClosingBracket {
			if !initial {
				processedComma := t.parseCharacter(codeComma)
				if !processedComma {
					t.output = InsertBeforeLastWhitespace(t.output, ",")
				}
			} else {
				initial = false
			}
			processedValue, err := t.parseValue()
			if err != nil {
				return false, err
			}
			if !processedValue {
				t.output = []rune(stripLastOccurrence(string(t.output), ",", false))
				break
			}
		}
		if t.CharCode(t.i) == codeClosingBracket {
			t.output = append(t.output, ']')
			t.i++
		} else {
			t.output = InsertBeforeLastWhitespace(t.output, "]")
		}
		return true, nil
	}
	return false, nil
}

func (t *RepairText) parseUnquotedString() (bool, error) {
	start := t.i
	for t.i < len(t.text) && !IsDelimiter(t.text[t.i]) {
		t.i++
	}
	if t.i > start {
		if t.CharCode(t.i) == codeOpenParenthesis {
			t.i++
			_, err := t.parseValue()
			if err != nil {
				return false, err
			}
			if t.CharCode(t.i) == codeCloseParenthesis {
				t.i++
				if t.CharCode(t.i) == codeSemicolon {
					t.i++
				}
			}
			return true, nil
		} else {
			// repair unquoted string
			// also, repair undefined to null
			for IsWhitespace(t.CharCode(t.i-1)) && t.i > 0 {
				t.i--
			}
			symbol := string(t.Slice(start, t.i))
			if symbol == "undefined" {
				t.output = append(t.output, []rune("null")...)
			} else {
				ss, _ := json.Marshal(symbol) // TODO
				t.output = append(t.output, []rune(string(ss))...)
			}
			if t.CharCode(t.i) == codeDoubleQuote {
				// we had a missing start quote, but now we encountered the end quote, so we can skip that one
				t.i++
			}

			return true, nil
		}
	}
	return false, nil
}

func (t *RepairText) parseCharacter(code rune) bool {
	if t.CharCode(t.i) == code && t.i < len(t.text) {
		t.output = append(t.output, t.text[t.i])
		t.i++
		return true
	}
	return false
}

func (t *RepairText) skipCharacter(code rune) bool {
	if t.CharCode(t.i) == code {
		t.i++
		return true
	}
	return false
}

func (t *RepairText) parseWhitespace() bool {
	var whitespace string
	var normal bool
	for t.i < len(t.text) {
		charCode := t.CharCode(t.i)
		normal = IsWhitespace(charCode)
		if normal || IsSpecialWhitespace(charCode) {
			if normal {
				whitespace += string(t.text[t.i])
			} else {
				// repair special whitespace
				whitespace += " "
			}
			t.i++
		} else {
			break
		}
	}
	if len(whitespace) > 0 {
		t.output = append(t.output, []rune(whitespace)...)
		return true
	}
	return false
}

func (t *RepairText) atEndOfBlockComment() bool {
	return t.CharCode(t.i) == codeAsterisk && t.CharCode(t.i+1) == codeSlash
}

func (t *RepairText) parseString(stopAtDelimiter bool) (bool, error) {
	var skipEscapeChars = t.CharCode(t.i) == codeBackslash
	if skipEscapeChars {
		t.i++
		skipEscapeChars = true
	}
	if IsQuote(t.CharCode(t.i)) {
		var isEndQuote func(rune) bool
		if IsDoubleQuote(t.CharCode(t.i)) {
			isEndQuote = IsDoubleQuote
		} else {
			if isSingleQuote(t.CharCode(t.i)) {
				isEndQuote = isSingleQuote
			} else {
				if IsSingleQuoteLike(t.CharCode(t.i)) {
					isEndQuote = IsSingleQuoteLike
				} else {
					isEndQuote = IsDoubleQuoteLike
				}
			}
		}

		//t.output = append(t.output, '"')
		iBefore := t.i

		tmpOutput := []rune(`"`)
		t.i++
		var isEndofString func(rune) bool
		if stopAtDelimiter {
			isEndofString = IsDelimiter
		} else {
			isEndofString = isEndQuote
		}

		for t.i < len(t.text) && !isEndofString(t.CharCode(t.i)) {
			if t.CharCode(t.i) == codeBackslash {
				char := t.Char(t.i + 1)
				if _, found := escapeCharacters[char]; found {
					tmpOutput = append(tmpOutput, t.Slice(t.i, t.i+2)...)
					t.i += 2
				} else if char == "u" {
					var j = 2
					for j < 6 && IsHex(t.CharCode(t.i+j)) {
						j++
					}
					if j == 6 {
						tmpOutput = append(tmpOutput, t.Slice(t.i, t.i+6)...)
						t.i += 6
					} else if (t.i + j) >= len(t.text) {
						t.i = len(t.text)
					} else {
						return false, InvalidUnicodeCharacter(string(t.Slice(t.i, t.i+6))).At(t.i)
					}
				} else {
					tmpOutput = append(tmpOutput, []rune(char)...)
					t.i += 2
				}
			} else {
				char := t.Char(t.i)
				code := t.CharCode(t.i)
				if code == codeDoubleQuote && t.CharCode(t.i-1) != codeBackslash {
					tmpOutput = append(tmpOutput, []rune("\\"+char)...)
					t.i++
				} else if IsControlCharacter(code) {
					tmpOutput = append(tmpOutput, []rune(controlCharacters[char])...)
					t.i++
				} else {
					if !IsValidStringCharacter(code) {
						return false, InvalidUnicodeCharacter(char).At(t.i)
					}
					tmpOutput = append(tmpOutput, []rune(char)...)
					t.i++
				}
			}
			if skipEscapeChars {
				processed := t.skipEscapeCharacter()
				if processed {
					// repair: skip escape character (nothing to do)
				}
			}
		}

		var hasEndQuote = IsQuote(t.CharCode(t.i))
		var valid = hasEndQuote && ((t.i+1) >= len(t.text) || IsDelimiter(nextNonWhiteSpaceCharacter(t.text, t.i+1)))
		if !valid && !stopAtDelimiter {
			t.i = iBefore
			return t.parseString(true)
		}
		if hasEndQuote {
			tmpOutput = append(tmpOutput, []rune(`"`)...)
			t.i++
		} else {
			tmpOutput = InsertBeforeLastWhitespace(tmpOutput, `"`)
		}

		t.output = append(t.output, tmpOutput...)
		_, err := t.parseConcatenatedString()
		if err != nil {
			return false, fmt.Errorf("failed to parseConcatenatedString: %v", err)
		}
		return true, nil
	}
	return false, nil
}

func (t *RepairText) parseConcatenatedString() (bool, error) {
	var processed bool
	t.parseWhitespaceAndSkipComments()
	for t.CharCode(t.i) == codePlus {
		processed = true
		t.i++
		t.parseWhitespaceAndSkipComments()
		t.output = []rune(stripLastOccurrence(string(t.output), `"`, true))
		start := len(t.output)
		parsedStr, err := t.parseString(false)
		if err != nil {
			return false, err
		}
		if parsedStr {
			t.output = RemoveAtIndex(t.output, start, 1)
		} else {
			t.output = InsertBeforeLastWhitespace(t.output, `"`)
		}
	}
	return processed, nil
}

func (t *RepairText) skipEscapeCharacter() bool {
	return t.skipCharacter(codeBackslash)
}

func (t *RepairText) parseNumber() (bool, error) {
	start := t.i
	if t.CharCode(t.i) == codeMinus {
		t.i++
		if ok, err := t.expectDigitOrRepair(start); err != nil {
			return false, err
		} else if ok {
			return true, nil
		}
	}

	for IsDigit(t.CharCode(t.i)) {
		t.i++
	}

	if t.CharCode(t.i) == codeDot {
		t.i++
		if ok, err := t.expectDigitOrRepair(start); err != nil {
			return false, err
		} else if ok {
			return true, nil
		}
		for IsDigit(t.CharCode(t.i)) {
			t.i++
		}
	}
	if t.CharCode(t.i) == codeLowercaseE || t.CharCode(t.i) == codeUppercaseE {
		t.i++
		if t.CharCode(t.i) == codeMinus || t.CharCode(t.i) == codePlus {
			t.i++
		}
		if ok, err := t.expectDigitOrRepair(start); err != nil {
			return false, err
		} else if ok {
			return true, nil
		}
		for IsDigit(t.CharCode(t.i)) {
			t.i++
		}
	}
	if t.i > start {
		numStr := string(t.Slice(start, t.i))
		if regexNumberWithLeadingZero.MatchString(numStr) {
			t.output = append(t.output, []rune(`"`+numStr+`"`)...)
		} else {
			t.output = append(t.output, []rune(numStr)...)
		}
		return true, nil
	}
	return false, nil
}

func (t *RepairText) expectDigitOrRepair(start int) (bool, error) {
	if t.i >= len(t.text) {
		t.output = append(t.output, append(t.Slice(start, t.i), '0')...)
		return true, nil
	} else {
		err := t.expectDigit(start)
		if err != nil {
			return false, err
		}
		return false, nil
	}
}

func (t *RepairText) expectDigit(start int) error {
	if !IsDigit(t.CharCode(t.i)) && t.i < len(t.text) {
		numSoFar := string(t.Slice(start, t.i))
		return ExpectDigit(numSoFar, string(t.text[t.i])).At(t.i)
	}
	return nil
}

func (t *RepairText) parseKeywords() bool {
	return t.parseKeyword("true", "true") ||
		t.parseKeyword("false", "false") ||
		t.parseKeyword("null", "null") ||
		// repair Python keywords True, False, None
		t.parseKeyword("True", "true") ||
		t.parseKeyword("False", "false") ||
		t.parseKeyword("None", "null")
}

func (t *RepairText) parseKeyword(name, value string) bool {
	if string(t.Slice(t.i, t.i+len(name))) == name {
		t.output = append(t.output, []rune(value)...)
		t.i += len(name)
		return true
	}
	return false
}

func min(a, b int) int {
	if a > b {
		return b
	}
	return a
}

func (t *RepairText) parseNewlineDelimitedJSON() error {
	initial := true
	processedValue := true
	var err error
	for processedValue {
		if !initial {
			processedComma := t.parseCharacter(codeComma)
			if !processedComma {
				t.output = InsertBeforeLastWhitespace(t.output, ",")
			}
		} else {
			initial = false
		}
		processedValue, err = t.parseValue()
		if err != nil {
			return err
		}
	}
	if !processedValue {
		t.output = []rune(stripLastOccurrence(string(t.output), ",", false))
	}
	t.output = append(append([]rune("[\n"), t.output...), []rune("\n]")...)
	return nil
}
