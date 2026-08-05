package main

import "testing"

func TestProfanityCheck(t *testing.T) {
	cases := []struct {
		input    string
		expected string
	}{
		{"I need a Kerfuffle!", "I need a Kerfuffle!"},
		{"I need a Kerfuffle! with Kerfuffle", "I need a Kerfuffle! with ****"},
		{"I need a Kerfuffle! with kerfuffle", "I need a Kerfuffle! with ****"},
	}

	for _, c := range cases {
		result := profanityCensor(c.input)
		if result != c.expected {
			t.Errorf("input %q: expected %q, got %q", c.input, c.expected, result)
		}
	}
}
