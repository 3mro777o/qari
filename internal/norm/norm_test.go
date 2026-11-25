package norm

import "testing"

func TestFold(t *testing.T) {
	cases := map[string]string{
		"مِشَارِي العفاسيّ": "مشاري العفاسي",
		"إبراهيم الأخضر":    "ابراهيم الاخضر",
		"Al-Fātiḥah ":       "al fatihah",
		"سورة البقرة":       "سوره البقره",
		"Yusuf  (12)":       "yusuf 12",
	}
	for in, want := range cases {
		if got := Fold(in); got != want {
			t.Errorf("Fold(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestMatch(t *testing.T) {
	h := Fold("Mishary Alafasy مشاري العفاسي")
	for _, q := range []string{"afasy", "عفاسي", "العَفاسي", "mish afa", "al afasy", ""} {
		if !Match(h, q) {
			t.Errorf("Match(%q) = false", q)
		}
	}
	if Match(h, "sudais") {
		t.Error("Match(sudais) = true")
	}
}
