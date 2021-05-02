package main

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
		actual := splitSpecials(c.in, specials)
		assert.Equal(t, c.out, actual)
	}
}
