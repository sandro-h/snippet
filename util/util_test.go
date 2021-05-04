package util

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSplitSpecial(t *testing.T) {
	cases := []struct {
		in  string
		out []string
	}{
		{"hello/world", []string{"hello", "/", "world"}},
		{"/hello/", []string{"/", "hello", "/"}},
		{"hello", []string{"hello"}},
		{"/", []string{"/"}},
		{"", nil},
		{"#!/usr/bin/env bash", []string{"#", "!", "/", "usr", "/", "bin", "/", "env bash"}},
	}

	specials := "/#"

	for _, c := range cases {
		actual := SplitSpecials(c.in, specials)
		assert.Equal(t, c.out, actual)
	}
}

func TestSplitSpecialNoList(t *testing.T) {
	actual := SplitSpecials("hello/world", "")
	assert.Equal(t, []string{"hello/world"}, actual)
}

func TestSearchFuzzy(t *testing.T) {
	cases := []struct {
		source   string
		targets  []string
		expected []string
	}{
		{
			"cert",
			[]string{"docker bash: docker exec -ti container bash", "openssl view cert: openssl x509 -text -noout -in"},
			[]string{"openssl view cert: openssl x509 -text -noout -in", "docker bash: docker exec -ti container bash"},
		},
		{
			"",
			[]string{"banana", "apple", "pear"},
			[]string{"banana", "apple", "pear"},
		},
	}

	for _, c := range cases {
		ranked := SearchFuzzy(c.source, c.targets)
		var actual []string
		for _, r := range ranked {
			actual = append(actual, r.Str)
		}
		assert.Equal(t, c.expected, actual)
	}
}
