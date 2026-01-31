
package tools

import (
	"github.com/stretchr/testify/assert"
	
	"testing"
)

func TestQAStringEmail (t *testing.T) {

	tests := map[string]struct {
		in, out string
		expected bool
	} {
		"empty": {in: "", out: "", expected: false},
		"hello world": {in: "Hello, world", out: "", expected: false},
		"partial": {in: "nathan", out: "", expected: false},
		"basic": {in: "a@b.co", out: "a@b.co", expected: true},
		"basic spaces": {in: " a@b.co ", out: "a@b.co", expected: true},
		"no end": {in: "nathan@example", out: "", expected: false},
		"no beginning": {in: "@example.com", out: "", expected: false},
		"no @": {in: "nathan.example.com", out: "", expected: false},
		"perfect": {in: "nathan@example.com", out: "nathan@example.com", expected: true},
		"longperfect": {in: "nathan@example.combo", out: "nathan@example.combo", expected: true},
		"phone": {in: "16175433004", out: "", expected: false},
		"shortGmail": {in: "asdfg@gmail.com", out: "", expected: false},
		"longGmail": {in: "asdfgh@gmail.com", out: "asdfgh@gmail.com", expected: true},
	}

	for name, data := range tests {
		t.Run (name, func(t *testing.T) {
			var str String
			str.Set(data.in)

			assert.Equal (t, data.expected, str.Email(), name + " is email")

			if data.expected {
				assert.Equal (t, data.out, str.String(), name + " string match")
			}
		})
	}
}
