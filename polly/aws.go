package polly

import (
	"strings"
)

// oggly normalises the name of ogg vorbis format for Polly.
// Other formats are left as-is, and if incorrect, will get an AWS diagnostic.
func oggly(format string) string {
	switch format {
	case "ogg", "oggvorbis", "ogg-vorbis":
		return "ogg_vorbis"
	default:
		return format
	}
}

func def(x string) *string {
	if x != "" {
		return &x
	}
	return nil
}

// toSSML converts text to SSML (AWS version), replacing reserved characters by XML entities.
func toSSML(text, rate string) string {
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
