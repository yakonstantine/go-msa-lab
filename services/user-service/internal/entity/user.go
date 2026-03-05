package entity

import (
	"fmt"
	"regexp"
	"strings"
	"time"
)

var corpKeyPattern = regexp.MustCompile("^([A-Z]{3}[0-9]{3})$")

type CorpKey string

func (ck CorpKey) Validate() error {
	if ok := corpKeyPattern.MatchString(string(ck)); !ok {
		return fmt.Errorf("Corporate Key should be in format 3 capital letters, 3 numbers - '%v': %w", ck, ErrInvalid)
	}
	return nil
}

type CountryCode string

var allCountryCodes = map[CountryCode]struct{}{
	"NL": {},
	"RU": {},
	"CH": {},
	"KZ": {},
	"US": {},
}

func (cc CountryCode) Validate() error {
	if _, ok := allCountryCodes[cc]; !ok {
		return fmt.Errorf("Country Code '%v': %w", cc, ErrInvalid)
	}
	return nil
}

var namePattern = regexp.MustCompile("^(^[\\p{L}\\p{N}\\-().,`' _]+$)$")

type Name string

func (n Name) Validate() error {
	if n == "" {
		return fmt.Errorf("can't be empty: %w", ErrInvalid)
	}
	if len(n) > 128 {
		return fmt.Errorf("should be up to 128 characters: %w", ErrInvalid)
	}
	if !namePattern.MatchString(string(n)) {
		return fmt.Errorf("can only contain any character, including non-Latin characters, "+
			"dashes, parentheses, spaces, dots, commas, and backticks: %w", ErrInvalid)
	}
	return nil
}

type UserProfile struct {
	CorpKey        CorpKey
	FirstName      Name
	LastName       Name
	FullName       string
	CountryCode    CountryCode
	DepartmentCode string
}

func (u *UserProfile) Sanitize() {
	u.CorpKey = CorpKey(strings.TrimSpace(strings.ToUpper(string(u.CorpKey))))
	u.FirstName = Name(strings.TrimSpace(string(u.FirstName)))
	u.LastName = Name(strings.TrimSpace(string(u.LastName)))
	u.FullName = strings.TrimSpace(u.FullName)
	u.CountryCode = CountryCode(strings.TrimSpace(strings.ToUpper(string(u.CountryCode))))
	u.DepartmentCode = strings.TrimSpace(strings.ToUpper(u.DepartmentCode))
}

func (u *UserProfile) Validate() error {
	valErr := ValidationError{Fields: map[string]error{}}

	if err := u.CorpKey.Validate(); err != nil {
		valErr.Fields["CorpKey"] = err
	}
	if err := u.CountryCode.Validate(); err != nil {
		valErr.Fields["CountryCode"] = err
	}
	if err := u.FirstName.Validate(); err != nil {
		valErr.Fields["FirstName"] = err
	}
	if err := u.LastName.Validate(); err != nil {
		valErr.Fields["LastName"] = err
	}
	if len(u.FullName) < 3 || len(u.FullName) > 256 {
		err := fmt.Errorf("Full Name should be between 3 and 256 characters - '%v': %w", len(u.FullName), ErrInvalid)
		valErr.Fields["FullName"] = err
	}
	if len(u.DepartmentCode) < 3 ||
		len(u.DepartmentCode) > 12 ||
		strings.Contains(u.DepartmentCode, " ") {
		valErr.Fields["DepartmentCode"] = fmt.Errorf("Department Code should be between 3 and 12 characters, shouldn't contain spaces - '%v': %w", len(u.DepartmentCode), ErrInvalid)
	}

	if len(valErr.Fields) > 0 {
		valErr.Message = "invalid user profile data"
		return &valErr
	}

	return nil
}

type User struct {
	CorpKey        CorpKey
	FullName       string
	CountryCode    CountryCode
	DepartmentCode string
	PrimarySMTP    Email
	SecondarySMTPs []Email
	CreatedAt      time.Time
	UpdatedAt      time.Time
	Deleted        bool
}
