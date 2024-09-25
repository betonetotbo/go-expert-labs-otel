package utils

import "fmt"

func ConcatFields(m map[string]any, fields ...string) string {
	r := ""
	for _, field := range fields {
		v, found := m[field]
		if found && v != nil {
			str := fmt.Sprintf("%v", v)
			if len(str) > 0 {
				if len(r) > 0 {
					r = fmt.Sprintf("%s,%s", r, str)
				} else {
					r = str
				}
			}
		}
	}
	return r
}
