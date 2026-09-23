package conversion

import (
	"bytes"
	"fmt"
	"math/big"
	"strings"
	"testing"
)

func format(t *testing.T, mode, template, sep, text string) (string, error) {
	t.Helper()
	f, err := NewFormatter(mode, template, sep)
	if err != nil {
		t.Fatal(err)
	}
	uid, err := ParseUID(text)
	if err != nil {
		t.Fatal(err)
	}
	return f.Format(uid)
}

func TestIDRoundTripForReaderUIDLengths(t *testing.T) {
	example, err := format(t, "id", "", ",", "7A403AB9")
	if err != nil || example != "042051029689" {
		t.Fatalf("default ID = %q, %v", example, err)
	}
	for _, original := range []string{
		"00000000", "7A403AB9", "04A1B2C3D4", "010203040506",
		"04A1B2C3D4E5F6", "E004015023FCB2A1", "04A1B2C3D4E5F6081920",
	} {
		t.Run(original, func(t *testing.T) {
			uid, _ := ParseUID(original)
			got, err := format(t, "id", "", ",", original)
			if err != nil {
				t.Fatal(err)
			}
			if !strings.HasPrefix(got, fmt.Sprintf("%02d", len(uid))) {
				t.Fatalf("length prefix missing: %q", got)
			}
			value, ok := new(big.Int).SetString(got[2:], 10)
			if !ok {
				t.Fatalf("invalid decimal ID %q", got)
			}
			restored := value.FillBytes(make([]byte, len(uid)))
			if !bytes.Equal(uid, restored) {
				t.Fatalf("ID %q restores %X, want %X", got, restored, uid)
			}
		})
	}
	short, _ := format(t, "id", "", ",", "00000000")
	long, _ := format(t, "id", "", ",", "0000000000")
	if short == long {
		t.Fatal("different UID lengths have the same ID")
	}
}

func TestDerivedFormatsAndLossyModes(t *testing.T) {
	tests := []struct {
		mode, uid, want string
	}{
		{"hex", "7A403AB9", "7A403AB9"},
		{"hex-reverse", "7A403AB9", "B93A407A"},
		{"decimal", "7A403AB9", "2051029689"},
		{"decimal-le", "7A403AB9", "3107602554"},
		{"decimal-le", "01020304050607", "1976943448883713"},
		{"wiegand26", "7A403AB9", "058,16506"},
		{"em4100", "0102034050", "003,16464"},
		{"conv1", "7A403AB9", "3107602554"},
		{"conv2", "7A403AB9", "058,16506"},
	}
	for _, tt := range tests {
		t.Run(tt.mode, func(t *testing.T) {
			got, err := format(t, tt.mode, "", ",", tt.uid)
			if err != nil || got != tt.want {
				t.Fatalf("got %q, %v; want %q", got, err, tt.want)
			}
		})
	}
	for _, mode := range []string{"wiegand26", "em4100"} {
		uid := "0102030405"
		if mode == "em4100" {
			uid = "01020304"
		}
		if _, err := format(t, mode, "", ",", uid); err == nil {
			t.Errorf("%s accepted incompatible UID length", mode)
		}
	}
	a, _ := format(t, "wiegand26", "", ",", "7A403AB9")
	b, _ := format(t, "wiegand26", "", ",", "7A403A00")
	if a != b {
		t.Fatalf("Wiegand truncation changed: %q != %q", a, b)
	}
}

func TestTemplateAndValidation(t *testing.T) {
	got, err := format(t, "id", "{length}|{uid}|{id}|{decimal-le}", ",", "00010203")
	if err != nil || got != "4|00010203|0466051|50462976" {
		t.Fatalf("template = %q, %v", got, err)
	}
	for _, template := range []string{"{}", "{unknown}", "{uid", "uid}", "{uid}}", "{uid{length}"} {
		if _, err := NewFormatter("id", template, ","); err == nil {
			t.Errorf("accepted invalid template %q", template)
		}
	}
	if _, err := NewFormatter("invalid", "", ","); err == nil {
		t.Error("accepted invalid mode")
	}
	if _, err := NewFormatter("id", "", "\n"); err == nil {
		t.Error("accepted newline in separator")
	}
}

func TestParseUIDRejectsSerialNoise(t *testing.T) {
	for _, line := range []string{"", "OK", "SCAN:+7A403AB9", "SCAN:-7A403AB9", "1234567", "123456789", "7A403ABG", "7a403ab9", "7A403AB9 ", strings.Repeat("A", 22)} {
		if _, err := ParseUID(line); err == nil {
			t.Errorf("accepted %q", line)
		}
	}
}
