package main

import "testing"

func TestCheckFormat(t *testing.T) {
	for _, tc := range []struct {
		name, source string
		valid        bool
	}{
		{"LF", "package sample\n\nvar value = 1\n", true},
		{"CRLF", "package sample\r\n\r\nvar value = 1\r\n", true},
		{"formatting", "package sample\n\nvar value=1\n", false},
		{"syntax", "package sample\nfunc (\n", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if err := checkFormat([]byte(tc.source)); (err == nil) != tc.valid {
				t.Fatalf("valid=%v, got %v", tc.valid, err)
			}
		})
	}
}
