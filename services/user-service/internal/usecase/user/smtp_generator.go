package user

import (
	"context"
	"fmt"
	"strings"
	"unicode"

	"github.com/yakonstantine/go-msa-lab/services/user-service/internal/entity"
	"github.com/yakonstantine/go-msa-lab/services/user-service/internal/usecase"
	"golang.org/x/text/unicode/norm"
)

const (
	defaultDomain = "co-group.com"
	maxRetries    = 100
)

func generatePrimarySMTP(ctx context.Context, smtpRepo usecase.SMTPRepository, up *entity.UserProfile) (entity.Email, error) {
	fn := sanitizeName(up.FirstName)
	ln := sanitizeName(up.LastName)
	un := fmt.Sprintf("%v.%v", fn, ln)
	if up.CountryCode == "KZ" || up.CountryCode == "RU" {
		un = fmt.Sprintf("%v.%v", ln, fn)
	}

	domain := calcDomain(up.CountryCode, up.DepartmentCode)

	eml := fmt.Sprintf("%v@%v", un, domain)

	for i := range maxRetries {
		smtp, err := smtpRepo.GetByEmail(ctx, entity.Email(strings.ToLower(eml)))
		if err != nil {
			return "", err
		}
		if smtp == nil || smtp.Identity == string(up.CorpKey) {
			return entity.Email(strings.ToLower(eml)), nil
		}
		eml = fmt.Sprintf("%v.%v@%v", un, (i + 1), domain)
	}

	return "", fmt.Errorf("all email addresses are occupied for %v %v", up.FirstName, up.LastName)
}

func sanitizeName(n entity.Name) entity.Name {
	r := strings.NewReplacer(
		"`", "",
		"'", "",
		" - ", "-",
		" -", "-",
		"- ", "-",
		" ", ".",
		". ", ".",
		"..", ".",
	)

	input := strings.TrimSpace(string(n))
	input = r.Replace(input)
	input = removeDiacritics(input)

	return entity.Name(strings.ToLower(input))
}

func removeDiacritics(str string) string {
	r := strings.NewReplacer(
		"ł", "l", "Ł", "L",
		"đ", "d", "Đ", "D",
		"ð", "d", "Ð", "D",
		"þ", "t", "Þ", "T",
		"ø", "o", "Ø", "O",
		"ı", "i", "İ", "i",
		"œ", "o", "Œ", "O",
		"ĳ", "i", "Ĳ", "I",
		"æ", "ae", "Æ", "Ae",
		"ß", "ss",
	)
	input := r.Replace(str)
	normalized := norm.NFD.String(input)

	var builder strings.Builder
	builder.Grow(len(normalized))

	for _, c := range normalized {
		if unicode.Is(unicode.Mn, c) {
			continue
		}
		builder.WriteRune(c)
	}

	return norm.NFC.String(builder.String())
}

func calcDomain(cc entity.CountryCode, dc string) string {
	switch cc {
	case "KZ":
		return "co.kz"
	case "RU":
		if strings.HasPrefix(dc, "FOO") {
			return "foo.ru"
		}
		return "co.ru"
	default:
		return defaultDomain
	}
}
