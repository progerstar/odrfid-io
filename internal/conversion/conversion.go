package conversion

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"strings"
)

// ParseUID accepts the complete uppercase hexadecimal UID emitted by HU*.
// The reader currently emits identifiers of 4 to 10 bytes.
func ParseUID(line string) ([]byte, error) {
	if len(line) < 8 || len(line) > 20 || len(line)%2 != 0 {
		return nil, fmt.Errorf("ожидается UID длиной от 4 до 10 байт")
	}
	for _, c := range line {
		if (c < '0' || c > '9') && (c < 'A' || c > 'F') {
			return nil, fmt.Errorf("UID должен содержать только заглавные HEX-символы")
		}
	}
	return hex.DecodeString(line)
}

type part struct {
	literal string
	field   string
}

// Formatter expands a checked template for one complete UID.
type Formatter struct {
	parts []part
	sep   string
}

var modes = map[string]string{
	"id":          "{id}",
	"hex":         "{uid}",
	"hex-reverse": "{uid-reverse}",
	"decimal":     "{decimal}",
	"decimal-le":  "{decimal-le}",
	"wiegand26":   "{wiegand26}",
	"em4100":      "{em4100}",
	// Aliases for scripts using the original command line.
	"default": "{uid}",
	"conv1":   "{decimal-le}",
	"conv2":   "{wiegand26}",
}

var fields = map[string]bool{
	"id": true, "uid": true, "uid-reverse": true,
	"decimal": true, "decimal-le": true,
	"wiegand26": true, "em4100": true, "length": true,
}

// NewFormatter uses template when nonempty; otherwise it uses the named mode.
func NewFormatter(mode, template, sep string) (*Formatter, error) {
	if template == "" {
		var ok bool
		template, ok = modes[mode]
		if !ok {
			return nil, fmt.Errorf("неизвестный режим %q", mode)
		}
	}
	if strings.ContainsAny(sep, "\r\n") {
		return nil, fmt.Errorf("разделитель не может содержать перевод строки")
	}
	f := &Formatter{sep: sep}
	for template != "" {
		open := strings.IndexByte(template, '{')
		if open == -1 {
			if strings.ContainsRune(template, '}') {
				return nil, fmt.Errorf("лишняя закрывающая скобка в шаблоне")
			}
			f.parts = append(f.parts, part{literal: template})
			break
		}
		if strings.ContainsRune(template[:open], '}') {
			return nil, fmt.Errorf("лишняя закрывающая скобка в шаблоне")
		}
		if open > 0 {
			f.parts = append(f.parts, part{literal: template[:open]})
		}
		template = template[open+1:]
		close := strings.IndexByte(template, '}')
		if close == -1 || strings.ContainsRune(template[:close], '{') {
			return nil, fmt.Errorf("незакрытое поле в шаблоне")
		}
		field := template[:close]
		if !fields[field] {
			return nil, fmt.Errorf("неизвестное поле {%s}", field)
		}
		f.parts = append(f.parts, part{field: field})
		template = template[close+1:]
	}
	if len(f.parts) == 0 {
		return nil, fmt.Errorf("пустой шаблон")
	}
	return f, nil
}

func reverse(data []byte) []byte {
	reversed := append([]byte(nil), data...)
	for i, j := 0, len(reversed)-1; i < j; i, j = i+1, j-1 {
		reversed[i], reversed[j] = reversed[j], reversed[i]
	}
	return reversed
}

func (f *Formatter) Format(uid []byte) (string, error) {
	if len(uid) < 4 || len(uid) > 10 {
		return "", fmt.Errorf("ожидается UID длиной от 4 до 10 байт")
	}
	var out strings.Builder
	for _, p := range f.parts {
		if p.field == "" {
			out.WriteString(p.literal)
			continue
		}
		switch p.field {
		case "id":
			// The two-digit byte count preserves leading zero bytes and separates lengths.
			fmt.Fprintf(&out, "%02d%s", len(uid), new(big.Int).SetBytes(uid).String())
		case "uid":
			out.WriteString(strings.ToUpper(hex.EncodeToString(uid)))
		case "uid-reverse":
			out.WriteString(strings.ToUpper(hex.EncodeToString(reverse(uid))))
		case "decimal":
			out.WriteString(new(big.Int).SetBytes(uid).String())
		case "decimal-le":
			out.WriteString(new(big.Int).SetBytes(reverse(uid)).String())
		case "length":
			fmt.Fprintf(&out, "%d", len(uid))
		case "wiegand26":
			if len(uid) != 4 {
				return "", fmt.Errorf("{wiegand26} требует UID длиной 4 байта, получено %d", len(uid))
			}
			fmt.Fprintf(&out, "%03d%s%05d", uid[2], f.sep, uint16(uid[1])<<8|uint16(uid[0]))
		case "em4100":
			if len(uid) != 5 {
				return "", fmt.Errorf("{em4100} требует UID длиной 5 байт, получено %d", len(uid))
			}
			fmt.Fprintf(&out, "%03d%s%05d", uid[2], f.sep, uint16(uid[3])<<8|uint16(uid[4]))
		}
	}
	return out.String(), nil
}
