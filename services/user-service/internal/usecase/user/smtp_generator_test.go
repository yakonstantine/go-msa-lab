package user

import (
	"fmt"
	"testing"

	"github.com/yakonstantine/go-msa-lab/services/user-service/internal/entity"
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
