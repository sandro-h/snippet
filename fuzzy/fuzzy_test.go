package fuzzy

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

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
