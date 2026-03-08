package user

import (
	"context"
	"fmt"
	"testing"

	"github.com/yakonstantine/go-msa-lab/services/user-service/internal/entity"
	"github.com/yakonstantine/go-msa-lab/services/user-service/internal/usecase/mock"
)

func TestSanitizeName(t *testing.T) {
	tests := []struct {
		name entity.Name
		want entity.Name
	}{
		{name: "Foo", want: "foo"},
		{name: "Foo-Bar", want: "foo-bar"},
		{name: "Foo- Bar", want: "foo-bar"},
		{name: "Foo -Bar", want: "foo-bar"},
		{name: "Foo - Bar", want: "foo-bar"},
		{name: "foo Bar", want: "foo.bar"},
		{name: "van der Foo - Bar", want: "van.der.foo-bar"},
		{name: "DR. Bar", want: "dr.bar"},
		{name: "'t Foo", want: "t.foo"},
		{name: "`t Foo", want: "t.foo"},
		{name: "Țăâîș", want: "taais"},
		{name: "șț ÎȚÎ", want: "st.iti"},
		{name: "Áéíö", want: "aeio"},
		{name: "Üúűó", want: "uuuo"},
		{name: "Ñíéá", want: "niea"},
		{name: "sław", want: "slaw"},
		{name: "Đđe", want: "dde"},
		{name: "Vić", want: "vic"},
		{name: "Øys", want: "oys"},
		{name: "Søren", want: "soren"},
		{name: "Ævar", want: "aevar"},
		{name: "Þór", want: "tor"},
		{name: "Mül", want: "mul"},
		{name: "iß", want: "iss"},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("case %v", tt.name), func(t *testing.T) {
			got := sanitizeName(tt.name)
			if got != tt.want {
				t.Errorf("Sanitize(%v) = %v, want %v", tt.name, got, tt.want)
			}
		})
	}
}

func TestCalcDomain(t *testing.T) {
	tests := []struct {
		name           string
		countryCode    entity.CountryCode
		departmentCode string
		want           string
	}{
		{name: "default domain - CH", countryCode: "CH", departmentCode: "ABC-123", want: defaultDomain},
		{name: "default domain - US", countryCode: "US", departmentCode: "ABC-123", want: defaultDomain},
		{name: "default domain - NL", countryCode: "NL", departmentCode: "ABC-123", want: defaultDomain},
		{name: "specific domain - KZ", countryCode: "KZ", departmentCode: "ABC-123", want: "co.kz"},
		{name: "specific domain - RU", countryCode: "RU", departmentCode: "ABC-123", want: "co.ru"},
		{name: "specific domain - RU Foo", countryCode: "RU", departmentCode: "FOO-123", want: "foo.ru"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := calcDomain(tt.countryCode, tt.departmentCode)
			if got != tt.want {
				t.Errorf("calcDomain() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGeneratePrimarySMTP(t *testing.T) {
	ctx := context.Background()
	tests := []struct {
		name       string
		up         *entity.UserProfile
		duplicates int
		want       string
		wantErr    bool
	}{
		{name: "get smtp by email error",
			up: &entity.UserProfile{
				FirstName:      "foo",
				LastName:       "bar",
				CountryCode:    "CH",
				DepartmentCode: "ABC-123",
			},
			wantErr: true,
		},
		{name: "ru smtp pattern",
			up: &entity.UserProfile{
				FirstName:      "foo",
				LastName:       "bar",
				CountryCode:    "RU",
				DepartmentCode: "ABC-123",
			},
			duplicates: 0,
			want:       "bar.foo@co.ru",
			wantErr:    false,
		},
		{name: "kz smtp pattern",
			up: &entity.UserProfile{
				FirstName:      "foo",
				LastName:       "bar",
				CountryCode:    "KZ",
				DepartmentCode: "ABC-123",
			},
			duplicates: 0,
			want:       "bar.foo@co.kz",
			wantErr:    false,
		},
		{name: "smtp is not in use",
			up: &entity.UserProfile{
				FirstName:      "Áéíö",
				LastName:       "van der Ñíéá",
				CountryCode:    "NL",
				DepartmentCode: "ABC-123",
			},
			duplicates: 0,
			want:       "aeio.van.der.niea@" + defaultDomain,
			wantErr:    false,
		},
		{name: "smtp is in use one time",
			up: &entity.UserProfile{
				FirstName:      "Mül",
				LastName:       "Bar",
				CountryCode:    "US",
				DepartmentCode: "ABC-123",
			},
			duplicates: 1,
			want:       "mul.bar.1@" + defaultDomain,
			wantErr:    false,
		},
		{name: "smtp is in use two times",
			up: &entity.UserProfile{
				FirstName:      "DR. Bar",
				LastName:       "Foo - Bar",
				CountryCode:    "CH",
				DepartmentCode: "ABC-123",
			},
			duplicates: 2,
			want:       "dr.bar.foo-bar.2@" + defaultDomain,
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		index := 0
		mockSMTPRepo := &mock.SMTPRepository{
			GetByEmailFn: func(context.Context, entity.Email) (*entity.SMTPAddress, error) {
				if tt.wantErr {
					return nil, fmt.Errorf("error")
				}
				if tt.duplicates > index {
					index++
					return &entity.SMTPAddress{}, nil
				}
				return nil, nil
			},
		}

		t.Run(tt.name, func(t *testing.T) {
			got, gotErr := generatePrimarySMTP(ctx, mockSMTPRepo, tt.up)
			if tt.wantErr {
				if gotErr == nil {
					t.Errorf("generatePrimarySMTP error nil, want error")
				}
				return
			}

			if gotErr != nil {
				t.Fatalf("generatePrimarySMTP error %v, want no error", gotErr)
			}
			if got != entity.Email(tt.want) {
				t.Errorf("generatePrimarySMTP = %v, want %v", got, tt.want)
			}
		})
	}
}
