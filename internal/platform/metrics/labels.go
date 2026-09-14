package metrics

import "strings"

// labelKey renders alternating key/value pairs into a stable series identity.
// An odd trailing argument is dropped rather than panicking: a metric call must never
// be able to take the process down.
func labelKey(labels []string) string {
	if len(labels) < 2 {
		return ""
	}
	var b strings.Builder
	for i := 0; i+1 < len(labels); i += 2 {
		if b.Len() > 0 {
			b.WriteString(",")
		}
		b.WriteString(labels[i])
		b.WriteString(`="`)
		b.WriteString(escape(labels[i+1]))
		b.WriteString(`"`)
	}
	return b.String()
}

func renderLabels(labels, extra string) string {
	switch {
	case labels == "" && extra == "":
		return ""
	case labels == "":
		return "{" + extra + "}"
	case extra == "":
		return "{" + labels + "}"
	default:
		return "{" + labels + "," + extra + "}"
	}
}

// escape protects the exposition format from label values we do not control.
func escape(s string) string {
	r := strings.NewReplacer(`\`, `\\`, `"`, `\"`, "\n", `\n`)
	return r.Replace(s)
}
