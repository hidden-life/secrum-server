package phone

import "strings"

func Normalize(in string) string {
	s := strings.TrimSpace(in)
	// remove +
	s = strings.TrimPrefix(s, "+")
	var out strings.Builder
	for _, r := range s {
		if r >= '0' && r <= '9' {
			out.WriteRune(r)
		}
	}

	return out.String()
}
