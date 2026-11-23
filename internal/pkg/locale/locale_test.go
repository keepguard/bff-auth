package locale

import "testing"

func TestNormalize(t *testing.T) {
	cases := map[string]string{
		"":        "",
		"   ":     "",
		"pt-BR":   PtBR,
		"pt":      PtBR,
		"pt-PT":   PtBR,
		"pt_BR":   PtBR,
		"PT-br":   PtBR,
		"en-US":   EnUS,
		"en":      EnUS,
		"es-ES":   EsES,
		"es":      EsES,
		"es-MX":   EsES,
		"it":      EnUS,
		"ja":      EnUS,
		"de":      EnUS,
		"xx-junk": EnUS,
		"<b>x":    EnUS,
	}
	for raw, want := range cases {
		if got := Normalize(raw); got != want {
			t.Errorf("Normalize(%q) = %q, want %q", raw, got, want)
		}
	}
}
