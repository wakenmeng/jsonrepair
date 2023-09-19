package jsonrepair

import (
	"fmt"
)

var (
	UnexpectedCharacterError = NewJSONRepairError("Unexpected character")
	ObjectKeyExpectedError   = NewJSONRepairError("Object key expected")
	ColonExpectedError       = NewJSONRepairError("Colon expected")
	UnexpectedEndError       = NewJSONRepairError("Unexpected end of json string")
)

type (
	JSONRepairError struct {
		Message  string
		Position int
	}

	InvalidUnicodeCharacterError struct {
		JSONRepairError
	}

	ExpectDigitError struct {
		JSONRepairError
		Got string
	}
)

func NewJSONRepairError(msg string) JSONRepairError {
	return JSONRepairError{
		Message: msg,
	}
}

func (e JSONRepairError) MessageAppend(s string) JSONRepairError {
	e.Message += " " + s
	return e
}

func (e JSONRepairError) At(pos int) JSONRepairError {
	e.Position = pos
	return e
}

func (e JSONRepairError) Error() string {
	return fmt.Sprintf("%s at position %d", e.Message, e.Position)
}

func ExpectDigit(numSoFar, got string) JSONRepairError {
	msg := fmt.Sprintf("Invalid number '%s', expecting a digit ", numSoFar)
	if len(msg) > 0 {
		msg += fmt.Sprintf("but got '%s'", got)
	} else {
		msg += "but reached end of input"
	}
	return NewJSONRepairError(msg)
}

func InvalidUnicodeCharacter(ch string) JSONRepairError {
	msg := fmt.Sprintf("Invalid unicode character \"%s\"", ch)
	return NewJSONRepairError(msg)
}
