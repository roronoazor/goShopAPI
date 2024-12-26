package validators

import (
	"unicode"
)

type PasswordError struct {
	MinLength   bool
	UpperCase   bool
	LowerCase   bool
	Number      bool
	SpecialChar bool
}

func (e PasswordError) Error() string {
	return "Password validation failed"
}

func ValidatePassword(password string) error {
	var err PasswordError

	if len(password) < 8 {
		err.MinLength = true
	}

	var (
		hasUpper   bool
		hasLower   bool
		hasNumber  bool
		hasSpecial bool
	)

	for _, char := range password {
		switch {
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsLower(char):
			hasLower = true
		case unicode.IsNumber(char):
			hasNumber = true
		case unicode.IsPunct(char) || unicode.IsSymbol(char):
			hasSpecial = true
		}
	}

	if !hasUpper {
		err.UpperCase = true
	}
	if !hasLower {
		err.LowerCase = true
	}
	if !hasNumber {
		err.Number = true
	}
	if !hasSpecial {
		err.SpecialChar = true
	}

	if err.MinLength || err.UpperCase || err.LowerCase || err.Number || err.SpecialChar {
		return err
	}

	return nil
}
