package entity

import (
	"fmt"
	"strings"
	"time"
)

type CorpKey string

func (ck CorpKey) Validate() error {
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
		return fmt.Errorf("country code '%v': %w", cc, ErrInvalid)
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

func (u *User) Sanitize() {
	u.CorpKey = CorpKey(strings.TrimSpace(strings.ToUpper(string(u.CorpKey))))
	u.FullName = strings.TrimSpace(u.FullName)
	u.CountryCode = CountryCode(strings.TrimSpace(strings.ToUpper(string(u.CountryCode))))
	u.DepartmentCode = strings.TrimSpace(strings.ToUpper(u.DepartmentCode))
}

func (u *User) Validate() map[string]error {
	errs := make(map[string]error)

	if err := u.CorpKey.Validate(); err != nil {
		errs["CorpKey"] = err
	}
	if err := u.CountryCode.Validate(); err != nil {
		errs["CountryCode"] = err
	}
	if len(u.FullName) < 3 && len(u.FullName) > 256 {
		err := fmt.Errorf("full name should should be between 3 and 255 characters, now %v: %w", len(u.FullName), ErrInvalid)
		errs["FullName"] = err
	}

	return errs
}
