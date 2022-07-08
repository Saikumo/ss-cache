package consistenthash

import (
	"strconv"
	"testing"
)

func TestHash(t *testing.T) {
	hash := NewMap(3, func(data []byte) uint32 {
		i, _ := strconv.Atoi(string(data))
		return uint32(i)
	})

	hash.Add("2", "4", "6")

	testCases := map[string]string{
		"20": "2",
		"22": "2",
		"41": "4",
		"43": "6",
		"63": "2",
	}

	for k, v := range testCases {
		node := hash.Get(k)
		if node != v {
			t.Fatalf("%s的hash应该由%s结点来处理，但被%s结点处理", k, v, node)
		}
	}

	hash.Add("8")
	testCases["80"] = "8"
	testCases["63"] = "8"

	for k, v := range testCases {
		node := hash.Get(k)
		if node != v {
			t.Fatalf("%s的hash应该由%s结点来处理，但被%s结点处理", k, v, node)
		}
	}
}
