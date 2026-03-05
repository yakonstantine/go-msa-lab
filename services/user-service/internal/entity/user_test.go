package entity

import (
	"errors"
	"maps"
	"testing"
)

func TestCorpKeyValidate(t *testing.T) {
	tests := []struct {
		name    string
		corpKey CorpKey
		wantErr bool
	}{
		{name: "valid capital case", corpKey: "ABC123", wantErr: false},
		{name: "invalid mixed case", corpKey: "aBc123", wantErr: true},
		{name: "invalid lower case", corpKey: "abc123", wantErr: true},
		{name: "invalid not sanitized", corpKey: " ABC123 ", wantErr: true},
		{name: "invalid 2 letters", corpKey: "ab123", wantErr: true},
		{name: "invalid 4 letters", corpKey: "abcd123", wantErr: true},
		{name: "invalid 2 numbers", corpKey: "abc12", wantErr: true},
		{name: "invalid 4 numbers", corpKey: "abc1234", wantErr: true},
		{name: "invalid 2 letters and numbers", corpKey: "ab12", wantErr: true},
		{name: "invalid 4 letters and numbers", corpKey: "abcd1234", wantErr: true},
		{name: "invalid random chars", corpKey: "a1b2c3", wantErr: true},
		{name: "invalid underscore", corpKey: "a_b123", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotErr := tt.corpKey.Validate()
			if gotErr != nil {
				if !errors.Is(gotErr, ErrInvalid) {
					t.Fatalf("Validate = %v, want ErrInvalid", gotErr)
				}
				if !tt.wantErr {
					t.Errorf("Validate = %v, want no error", gotErr)
				}
			} else if tt.wantErr {
				t.Errorf("Validate no error, want validation error")
			}
		})
	}
}

func TestCountryCodeValidate(t *testing.T) {
	tests := []struct {
		name        string
		countryCode CountryCode
		wantErr     bool
	}{
		{name: "valid capital case", countryCode: "KZ", wantErr: false},
		{name: "invalid mixed case", countryCode: "kZ", wantErr: true},
		{name: "invalid lower case", countryCode: "kz", wantErr: true},
		{name: "invalid not sanitized", countryCode: " KZ ", wantErr: true},
		{name: "invalid not in the list", countryCode: "AA", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotErr := tt.countryCode.Validate()
			if gotErr != nil {
				if !errors.Is(gotErr, ErrInvalid) {
					t.Fatalf("Validate = %v, want ErrInvalid", gotErr)
				}
				if !tt.wantErr {
					t.Errorf("Validate = %v, want no error", gotErr)
				}
			} else if tt.wantErr {
				t.Errorf("Validate no error, want validation error")
			}
		})
	}
}

func TestUserProfileValidate(t *testing.T) {
	tests := []struct {
		name string
		user UserProfile
		want map[string]error
	}{
		{name: "valid",
			user: UserProfile{
				CorpKey:        "ABC123",
				CountryCode:    "KZ",
				FirstName:      "John",
				LastName:       "Doe",
				FullName:       "Doe, J.P. (John)",
				DepartmentCode: "FOO-987",
			},
			want: map[string]error{},
		},
		{name: "invalid CorpKey",
			user: UserProfile{
				CorpKey:        "123abc",
				CountryCode:    "KZ",
				FirstName:      "John",
				LastName:       "Doe",
				FullName:       "Doe, J.P. (John)",
				DepartmentCode: "FOO-987",
			},
			want: map[string]error{
				"CorpKey": ErrInvalid,
			},
		},
		{name: "invalid CountryCode",
			user: UserProfile{
				CorpKey:        "ABC123",
				CountryCode:    "AA",
				FirstName:      "John",
				LastName:       "Doe",
				FullName:       "Doe, J.P. (John)",
				DepartmentCode: "FOO-987",
			},
			want: map[string]error{
				"CountryCode": ErrInvalid,
			},
		},
		{name: "invalid FullName too short",
			user: UserProfile{
				CorpKey:        "ABC123",
				CountryCode:    "KZ",
				FirstName:      "John",
				LastName:       "Doe",
				FullName:       "JD",
				DepartmentCode: "FOO-987",
			},
			want: map[string]error{
				"FullName": ErrInvalid,
			},
		},
		{name: "invalid FullName too long",
			user: UserProfile{
				CorpKey:     "ABC123",
				CountryCode: "KZ",
				FirstName:   "John",
				LastName:    "Doe",
				FullName: "Long-Long, N.M.G (Name) Long-Long, N.M.G (Name) Long-Long, N.M.G (Name) Long-Long, N.M.G (Name)" +
					" Long-Long, N.M.G (Name) Long-Long, N.M.G (Name) Long-Long, N.M.G (Name) Long-Long, N.M.G (Name)" +
					" Long-Long, N.M.G (Name) Long-Long, N.M.G (Name) Long-Long, N.M.G (Name) Long-Long, N.M.G (Name)",
				DepartmentCode: "FOO-987",
			},
			want: map[string]error{
				"FullName": ErrInvalid,
			},
		},
		{name: "invalid DepartmentCode too short",
			user: UserProfile{
				CorpKey:        "ABC123",
				CountryCode:    "KZ",
				FirstName:      "John",
				LastName:       "Doe",
				FullName:       "Doe, J.P. (John)",
				DepartmentCode: "K9",
			},
			want: map[string]error{
				"DepartmentCode": ErrInvalid,
			},
		},
		{name: "invalid DepartmentCode too long",
			user: UserProfile{
				CorpKey:        "ABC123",
				CountryCode:    "KZ",
				FirstName:      "John",
				LastName:       "Doe",
				FullName:       "Doe, J.P. (John)",
				DepartmentCode: "ABCDEF-123456",
			},
			want: map[string]error{
				"DepartmentCode": ErrInvalid,
			},
		},
		{name: "invalid DepartmentCode contains spaces",
			user: UserProfile{
				CorpKey:        "ABC123",
				CountryCode:    "KZ",
				FirstName:      "John",
				LastName:       "Doe",
				FullName:       "Doe, J.P. (John)",
				DepartmentCode: "FOO 123",
			},
			want: map[string]error{
				"DepartmentCode": ErrInvalid,
			},
		},
		{name: "invalid all properties",
			user: UserProfile{
				CorpKey:        "A1B2C3",
				CountryCode:    "AA",
				FirstName:      "",
				LastName:       "",
				FullName:       "JD",
				DepartmentCode: "FOO 987",
			},
			want: map[string]error{
				"CorpKey":        ErrInvalid,
				"CountryCode":    ErrInvalid,
				"FirstName":      ErrInvalid,
				"LastName":       ErrInvalid,
				"FullName":       ErrInvalid,
				"DepartmentCode": ErrInvalid,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.user.Validate()
			if got == nil {
				if len(tt.want) > 0 {
					t.Fatalf("Validate no error, want %v", tt.want)
				}
				return
			}
			if len(tt.want) == 0 {
				t.Fatalf("Validate = %v, want no error", got)
			}
			var gotValErr *ValidationError
			if !errors.As(got, &gotValErr) {
				t.Fatalf("Validate = %v, want ValidationError", got)
			}
			if !maps.EqualFunc(gotValErr.Fields, tt.want, errors.Is) {
				t.Errorf("Validate = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestUserProfileSanitize(t *testing.T) {
	user := UserProfile{
		CorpKey:        " abc123 ",
		CountryCode:    "kz ",
		FirstName:      " Sław ",
		LastName:       " der Foo - Bar",
		FullName:       " Doe, J.D. (John) ",
		DepartmentCode: " foo-123 ",
	}
	want := UserProfile{
		CorpKey:        "ABC123",
		CountryCode:    "KZ",
		FirstName:      "Sław",
		LastName:       "der Foo - Bar",
		FullName:       "Doe, J.D. (John)",
		DepartmentCode: "FOO-123",
	}

	user.Sanitize()

	if user.CorpKey != want.CorpKey {
		t.Errorf("CorpKey Sanitize = %v, want %v", user.CorpKey, want.CorpKey)
	}
	if user.CountryCode != want.CountryCode {
		t.Errorf("CountryCode Sanitize = %v, want %v", user.CountryCode, want.CountryCode)
	}
	if user.FirstName != want.FirstName {
		t.Errorf("FirstName Sanitize = %v, want %v", user.FullName, want.FullName)
	}
	if user.LastName != want.LastName {
		t.Errorf("FullName Sanitize = %v, want %v", user.FullName, want.FullName)
	}
	if user.FullName != want.FullName {
		t.Errorf("FullName Sanitize = %v, want %v", user.FullName, want.FullName)
	}
	if user.DepartmentCode != want.DepartmentCode {
		t.Errorf("DepartmentCode Sanitize = %v, want %v", user.DepartmentCode, want.DepartmentCode)
	}
}
