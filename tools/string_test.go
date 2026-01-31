
package tools

import (
	"github.com/stretchr/testify/assert"
	
	"testing"
)

func TestQAStringPhone1 (t *testing.T) {

	tests := map[string]struct {
		in, out string
		expected bool
	} {
		"empty": {in: "", out: "", expected: false},
		"hello world": {in: "Hello, world", out: "", expected: false},
		"partial": {in: "5433004", out: "", expected: false},
		"no 1": {in: "6175433004", out: "16175433004", expected: true},
		"perfect": {in: "16175433004", out: "16175433004", expected: true},
		"spaces": {in: "617 543 3004", out: "16175433004", expected: true},
		"parentheses": {in: "(617) 543-3004", out: "16175433004", expected: true},
		"periods": {in: "617.543.3004", out: "16175433004", expected: true},
		"international": {in: "+1 (617)-543.3004", out: "16175433004", expected: true},
		"email": {in: "nathan@beelineroutes.com", out: "", expected: false},
	}

	for name, data := range tests {
		t.Run (name, func(t *testing.T) {
			var str String
			str.Set(data.in)

			assert.Equal (t, data.expected, str.Phone(), name + " is phone #")

			if data.expected {
				assert.Equal (t, data.out, str.String(), name + " string match")
			}
			
		})
	}

	var str String
	str.Set("16175433004")
	str.PhoneFormat()
	assert.Equal (t, "(617) 543-3004", str.String())
}

func TestQAStringPhone2 (t *testing.T) {
	var str String
	str.Set("+1-808-628-8291")
	assert.Equal (t, true, str.Phone())
	assert.Equal (t, "18086288291", str.String())
}

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
		"no end": {in: "nathan@beelineroutes", out: "", expected: false},
		"no beginning": {in: "@beelineroutes.com", out: "", expected: false},
		"no @": {in: "nathan.beelineroutes.com", out: "", expected: false},
		"perfect": {in: "nathan@beelineroutes.com", out: "nathan@beelineroutes.com", expected: true},
		"longperfect": {in: "nathan@beelineroutes.combo", out: "nathan@beelineroutes.combo", expected: true},
		"phone": {in: "16175433004", out: "", expected: false},
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

func TestQAStringUrl (t *testing.T) {
	tests := map[string]struct {
		in, out string
		expected bool
	} {
		"empty": {in: "", out: "", expected: false},
		"hello world": {in: "Hello, world", out: "", expected: false},
		"phone": {in: "16175433004", out: "", expected: false},
		"email": {in: "nathan@beelineroutes.com", out: "", expected: false},
		"base": {in: "beelineroutes.com", out: "https://beelineroutes.com", expected: true},
		"www": {in: "www.beelineroutes.com", out: "https://www.beelineroutes.com", expected: true},
		"http": {in: "http://beelineroutes.com", out: "http://beelineroutes.com", expected: true},
		"http-www": {in: "http://www.beelineroutes.com", out: "http://www.beelineroutes.com", expected: true},
		"perfect": {in: "https://beelineroutes.com", out: "https://beelineroutes.com", expected: true},
		"perfect-www": {in: "https://www.beelineroutes.com", out: "https://www.beelineroutes.com", expected: true},
	}

	subDirs := []string{"", "/one", "/one/"}
	params := []string{"", "?one=1", "?one=1&two=2"}
	hashTags := []string{"", "#one"}

	for name, data := range tests {
		for _, subdir := range subDirs {
			for _, param := range params {
				for _, hashTag := range hashTags {
					t.Run(name, func(t *testing.T) {
						var str String
						str.Sprintf("%s%s%s%s", data.in, subdir, param, hashTag)

						assert.Equal (t, data.expected, str.Url(), name + " is url")

						if data.expected {
							assert.Equal (t, data.out + subdir + param + hashTag, str.String(), name + " string match")
						}
					})
				}
			}
		}
	}

}

func TestQAStringSafe (t *testing.T) {

	tests := map[string]struct {
		in, safe, superSafe string
	} {
		"empty": {in: "", safe: "", superSafe: ""},
		"hello world": {in: "Hello World", safe: "hello world", superSafe: "helloworld"},
		"email": {in: "nathan@beelineroutes.com", safe: "nathanbeelineroutescom", superSafe: "nathanbeelineroutescom"},
		"phone": {in: "16175433004", safe: "16175433004", superSafe: "16175433004"},
		"url": {in: "https://www.beelineroutes.com/one?one=1#one", safe: "httpswwwbeelineroutescomoneone1one", superSafe: "httpswwwbeelineroutescomoneone1one"},		
	}

	for name, data := range tests {
		t.Run (name, func(t *testing.T) {
			var str String
			str.Set(data.in)

			assert.Equal (t, data.safe, str.SafeString(), name + " safe")
			assert.Equal (t, data.superSafe, str.SuperSafeString(), name + " super safe")
		})
	}
}

func TestQAStringIsUUID (t *testing.T) {

	tests := map[string]bool {
		"": false,
		"hello world": false, 
		"pro_965bb06f940a4bd188c09719bb1537c0": false,
		"0": false,
		"0-0": false,
		"3fccc0d7-d30b-42c1-a33b-d07a795f20b1": true,
		"00000000-0000-0000-0000-000000000000": true,
		"00000000-0000-0000-0000-00000000000": false,
		"0199aaf1-2fe7-7c5d-9aef-0435d25738e8": true,
	}

	for name, data := range tests {
		t.Run ("uuid test", func(t *testing.T) {
			var str String
			str.Set(name)

			assert.Equal (t, data, str.IsUUID())
		})
	}
}

func TestQAStringUUID (t *testing.T) {
	var str String
	str.Set("0199aaf1-2fe7-7c5d-9aef-0435d25738e8")

	assert.Equal (t, true, str.IsUUID())

	u := str.UUID()
	assert.Equal (t, true, u != nil)
	assert.Equal (t, str.String(), u.String())
}

func TestQAStringRemove (t *testing.T) {
	var str String 
	str.Set("message here: ")

	assert.Equal (t, "message here", str.Remove(":? ?$"))

	str.Set("message here: and here: ")

	assert.Equal (t, "message here: and here", str.Remove(":? ?$"))
}

func TestQAStringList1 (t *testing.T) {
	strList := NewStringList ("zero,one,two,three", ",")
	assert.Equal (t, 4, strList.Len())
	assert.Equal (t, "one", strList[1].String())
	
}

