package stream

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_makeEndpoint(t *testing.T) {
	testCases := []struct {
		url      *apiURL
		format   string
		env      string
		args     []interface{}
		expected string
	}{
		{
			url:      &apiURL{},
			format:   "test-%d-%s",
			args:     []interface{}{42, "asd"},
			expected: "https://api.getstream.io/api/v1.0/test-42-asd/?api_key=test",
		},
		{
			url:      &apiURL{},
			env:      "http://localhost:8000/api/v1.0/",
			format:   "test-%d-%s",
			args:     []interface{}{42, "asd"},
			expected: "http://localhost:8000/api/v1.0/test-42-asd/?api_key=test",
		},
	}

	for _, tc := range testCases {
		os.Setenv("STREAM_API_URL", tc.env)
		c := &Client{url: tc.url, key: "test"}
		assert.Equal(t, tc.expected, c.makeEndpoint(tc.format, tc.args...))
	}
}
