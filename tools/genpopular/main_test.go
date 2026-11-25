package main

import "testing"

func TestMatcher(t *testing.T) {
	m := newMatcher([]person{
		{1, "مشاري العفاسي", "Mishary Alafasi"},
		{2, "عبدالباسط عبدالصمد", "Abdulbasit Abdulsamad"},
		{3, "محمد صديق المنشاوي", "Mohammed Siddiq Al-Minshawi"},
		{4, "محمد أيوب", "Mohammad Ayyub"},
	})
	cases := []struct {
		ar, en string
		want   int
	}{
		{"مشاري راشد العفاسي", "", 1},          // extra middle name
		{"عبد الباسط عبد الصمد", "", 2},        // split words
		{"", "Mohammed Siddiq Al Minshawi", 3}, // English only
		{"محمد", "Mohamed", 0},                 // too vague
		{"سعد الغامدي", "Saad Al-Ghamdi", 0},   // not on mp3quran
	}
	for _, c := range cases {
		if got := m.find(c.ar, c.en); got != c.want {
			t.Errorf("find(%q, %q) = %d, want %d", c.ar, c.en, got, c.want)
		}
	}
}
