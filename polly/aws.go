package polly

import (
	"strings"
)

func def(x string) *string {
	if x != "" {
		return &x
	}
	return nil
}

// to_ssml converts text to SSML (AWS version), replacing reserved characters by XML entities.
func to_ssml(text, rate string) string {
	sb := &strings.Builder{}
	sb.WriteString("<speak><amazon:auto-breaths>")
	sb.WriteString("<prosody rate=\"" + rate + "\">")
	for _, r := range text {
		switch r {
		case '<':
			sb.WriteString("&lt;")
		case '&':
			sb.WriteString("&amp;")
		default:
			sb.WriteRune(r)
		}
	}
	sb.WriteString("</prosody>")
	sb.WriteString("</amazon:auto-breaths>")
	sb.WriteString("</speak>")
	return sb.String()
}
