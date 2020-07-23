package uuid

import (
	"fmt"
	goouid "github.com/google/uuid"
	"reflect"
	"strings"
	"testing"
	"time"
	"unsafe"
)

type test struct {
	in      string
	version goouid.Version
	variant goouid.Variant
	isuuid  bool
}

var timeNow = time.Now // for testing

var tests = []test{
	{"f47ac10b-58cc-0372-8567-0e02b2c3d479", 0, goouid.RFC4122, true},
	{"f47ac10b-58cc-1372-8567-0e02b2c3d479", 1, goouid.RFC4122, true},
	{"f47ac10b-58cc-2372-8567-0e02b2c3d479", 2, goouid.RFC4122, true},
	{"f47ac10b-58cc-3372-8567-0e02b2c3d479", 3, goouid.RFC4122, true},
	{"f47ac10b-58cc-4372-8567-0e02b2c3d479", 4, goouid.RFC4122, true},
	{"f47ac10b-58cc-5372-8567-0e02b2c3d479", 5, goouid.RFC4122, true},
	{"f47ac10b-58cc-6372-8567-0e02b2c3d479", 6, goouid.RFC4122, true},
	{"f47ac10b-58cc-7372-8567-0e02b2c3d479", 7, goouid.RFC4122, true},
	{"f47ac10b-58cc-8372-8567-0e02b2c3d479", 8, goouid.RFC4122, true},
	{"f47ac10b-58cc-9372-8567-0e02b2c3d479", 9, goouid.RFC4122, true},
	{"f47ac10b-58cc-a372-8567-0e02b2c3d479", 10, goouid.RFC4122, true},
	{"f47ac10b-58cc-b372-8567-0e02b2c3d479", 11, goouid.RFC4122, true},
	{"f47ac10b-58cc-c372-8567-0e02b2c3d479", 12, goouid.RFC4122, true},
	{"f47ac10b-58cc-d372-8567-0e02b2c3d479", 13, goouid.RFC4122, true},
	{"f47ac10b-58cc-e372-8567-0e02b2c3d479", 14, goouid.RFC4122, true},
	{"f47ac10b-58cc-f372-8567-0e02b2c3d479", 15, goouid.RFC4122, true},

	{"urn:uuid:f47ac10b-58cc-4372-0567-0e02b2c3d479", 4, goouid.Reserved, true},
	{"URN:UUID:f47ac10b-58cc-4372-0567-0e02b2c3d479", 4, goouid.Reserved, true},
	{"f47ac10b-58cc-4372-0567-0e02b2c3d479", 4, goouid.Reserved, true},
	{"f47ac10b-58cc-4372-1567-0e02b2c3d479", 4, goouid.Reserved, true},
	{"f47ac10b-58cc-4372-2567-0e02b2c3d479", 4, goouid.Reserved, true},
	{"f47ac10b-58cc-4372-3567-0e02b2c3d479", 4, goouid.Reserved, true},
	{"f47ac10b-58cc-4372-4567-0e02b2c3d479", 4, goouid.Reserved, true},
	{"f47ac10b-58cc-4372-5567-0e02b2c3d479", 4, goouid.Reserved, true},
	{"f47ac10b-58cc-4372-6567-0e02b2c3d479", 4, goouid.Reserved, true},
	{"f47ac10b-58cc-4372-7567-0e02b2c3d479", 4, goouid.Reserved, true},
	{"f47ac10b-58cc-4372-8567-0e02b2c3d479", 4, goouid.RFC4122, true},
	{"f47ac10b-58cc-4372-9567-0e02b2c3d479", 4, goouid.RFC4122, true},
	{"f47ac10b-58cc-4372-a567-0e02b2c3d479", 4, goouid.RFC4122, true},
	{"f47ac10b-58cc-4372-b567-0e02b2c3d479", 4, goouid.RFC4122, true},
	{"f47ac10b-58cc-4372-c567-0e02b2c3d479", 4, goouid.Microsoft, true},
	{"f47ac10b-58cc-4372-d567-0e02b2c3d479", 4, goouid.Microsoft, true},
	{"f47ac10b-58cc-4372-e567-0e02b2c3d479", 4, goouid.Future, true},
	{"f47ac10b-58cc-4372-f567-0e02b2c3d479", 4, goouid.Future, true},

	{"f47ac10b158cc-5372-a567-0e02b2c3d479", 0, goouid.Invalid, false},
	{"f47ac10b-58cc25372-a567-0e02b2c3d479", 0, goouid.Invalid, false},
	{"f47ac10b-58cc-53723a567-0e02b2c3d479", 0, goouid.Invalid, false},
	{"f47ac10b-58cc-5372-a56740e02b2c3d479", 0, goouid.Invalid, false},
	{"f47ac10b-58cc-5372-a567-0e02-2c3d479", 0, goouid.Invalid, false},
	{"g47ac10b-58cc-4372-a567-0e02b2c3d479", 0, goouid.Invalid, false},

	{"{f47ac10b-58cc-0372-8567-0e02b2c3d479}", 0, goouid.RFC4122, true},
	{"{f47ac10b-58cc-0372-8567-0e02b2c3d479", 0, goouid.Invalid, false},
	{"f47ac10b-58cc-0372-8567-0e02b2c3d479}", 0, goouid.Invalid, false},

	{"f47ac10b58cc037285670e02b2c3d479", 0, goouid.RFC4122, true},
	{"f47ac10b58cc037285670e02b2c3d4790", 0, goouid.Invalid, false},
	{"f47ac10b58cc037285670e02b2c3d47", 0, goouid.Invalid, false},
}

var constants = []struct {
	c    interface{}
	name string
}{
	{goouid.Person, "Person"},
	{goouid.Group, "Group"},
	{goouid.Org, "Org"},
	{goouid.Invalid, "Invalid"},
	{goouid.RFC4122, "RFC4122"},
	{goouid.Reserved, "Reserved"},
	{goouid.Microsoft, "Microsoft"},
	{goouid.Future, "Future"},
	{goouid.Domain(17), "Domain17"},
	{goouid.Variant(42), "BadVariant42"},
}

func testTest(t *testing.T, in string, tt test) {
	uuid, err := Parse(in)
	if ok := (err == nil); ok != tt.isuuid {
		t.Errorf("Parse(%s) got %v expected %v\b", in, ok, tt.isuuid)
	}
	if err != nil {
		return
	}

	if v := uuid.Variant(); v != tt.variant {
		t.Errorf("Variant(%s) got %d expected %d\b", in, v, tt.variant)
	}
	if v := uuid.Version(); v != tt.version {
		t.Errorf("Version(%s) got %d expected %d\b", in, v, tt.version)
	}
}

func testBytes(t *testing.T, in []byte, tt test) {
	uuid, err := ParseBytes(in)
	if ok := (err == nil); ok != tt.isuuid {
		t.Errorf("ParseBytes(%s) got %v expected %v\b", in, ok, tt.isuuid)
	}
	if err != nil {
		return
	}
	suuid, _ := Parse(string(in))
	if !reflect.DeepEqual(uuid, suuid) {
		t.Errorf("ParseBytes(%s) got %v expected %v\b", in, uuid, suuid)
	}
}

func TestUUID(t *testing.T) {
	for _, tt := range tests {
		testTest(t, tt.in, tt)
		testTest(t, strings.ToUpper(tt.in), tt)
		testBytes(t, []byte(tt.in), tt)
	}
}

func TestFromBytes(t *testing.T) {
	b := []byte{
		0x7d, 0x44, 0x48, 0x40,
		0x9d, 0xc0,
		0x11, 0xd1,
		0xb2, 0x45,
		0x5f, 0xfd, 0xce, 0x74, 0xfa, 0xd2,
	}
	uuid, err := goouid.FromBytes(b)
	if err != nil {
		t.Fatalf("%s", err)
	}
	for i := 0; i < len(uuid); i++ {
		if b[i] != uuid[i] {
			t.Fatalf("FromBytes() got %v expected %v\b", uuid[:], b)
		}
	}
}

func TestConstants(t *testing.T) {
	for x, tt := range constants {
		v, ok := tt.c.(fmt.Stringer)
		if !ok {
			t.Errorf("%x: %v: not a stringer", x, v)
		} else if s := v.String(); s != tt.name {
			v, _ := tt.c.(int)
			t.Errorf("%x: Constant %T:%d gives %q, expected %q", x, tt.c, v, s, tt.name)
		}
	}
}

func TestRandomUUID(t *testing.T) {
	m := make(map[string]bool)
	for x := 1; x < 32; x++ {
		uuid := New()
		s := uuid.String()
		if m[s] {
			t.Errorf("NewRandom returned duplicated UUID %s", s)
		}
		m[s] = true
		if v := uuid.Version(); v != 4 {
			t.Errorf("Random UUID of version %s", v)
		}
		if uuid.Variant() != goouid.RFC4122 {
			t.Errorf("Random UUID is variant %d", uuid.Variant())
		}
	}
}

func TestNew(t *testing.T) {
	m := make(map[UUID]bool)
	for x := 1; x < 32; x++ {
		s := New()
		if m[s] {
			t.Errorf("New returned duplicated UUID %s", s)
		}
		m[s] = true
		uuid, err := Parse(s.String())
		if err != nil {
			t.Errorf("New.String() returned %q which does not decode", s)
			continue
		}
		if v := uuid.Version(); v != 4 {
			t.Errorf("Random UUID of version %s", v)
		}
		if uuid.Variant() != goouid.RFC4122 {
			t.Errorf("Random UUID is variant %d", uuid.Variant())
		}
	}
}

func TestCoding(t *testing.T) {
	text := "7d444840-9dc0-11d1-b245-5ffdce74fad2"
	urn := "urn:uuid:7d444840-9dc0-11d1-b245-5ffdce74fad2"
	data := UUID{UUID: goouid.UUID{
		0x7d, 0x44, 0x48, 0x40,
		0x9d, 0xc0,
		0x11, 0xd1,
		0xb2, 0x45,
		0x5f, 0xfd, 0xce, 0x74, 0xfa, 0xd2,
	}}
	if v := data.String(); v != text {
		t.Errorf("%x: encoded to %s, expected %s", data, v, text)
	}
	if v := data.URN(); v != urn {
		t.Errorf("%x: urn is %s, expected %s", data, v, urn)
	}

	uuid, err := Parse(text)
	if err != nil {
		t.Errorf("Parse returned unexpected error %v", err)
	}
	if data != uuid {
		t.Errorf("%s: decoded to %s, expected %s", text, uuid, data)
	}
}

var asString = "f47ac10b-58cc-0372-8567-0e02b2c3d479"
var asBytes = []byte(asString)

func BenchmarkParse(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, err := Parse(asString)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkParseBytes(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, err := ParseBytes(asBytes)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// parseBytesUnsafe is to benchmark using unsafe.
func parseBytesUnsafe(b []byte) (UUID, error) {
	return Parse(*(*string)(unsafe.Pointer(&b)))
}

func BenchmarkParseBytesUnsafe(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, err := parseBytesUnsafe(asBytes)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// parseBytesCopy is to benchmark not using unsafe.
func parseBytesCopy(b []byte) (UUID, error) {
	return Parse(string(b))
}

func BenchmarkParseBytesCopy(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, err := parseBytesCopy(asBytes)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkNew(b *testing.B) {
	for i := 0; i < b.N; i++ {
		New()
	}
}

func BenchmarkUUID_String(b *testing.B) {
	uuid, err := Parse("f47ac10b-58cc-0372-8567-0e02b2c3d479")
	if err != nil {
		b.Fatal(err)
	}
	for i := 0; i < b.N; i++ {
		if uuid.String() == "" {
			b.Fatal("invalid uuid")
		}
	}
}

func BenchmarkUUID_URN(b *testing.B) {
	uuid, err := Parse("f47ac10b-58cc-0372-8567-0e02b2c3d479")
	if err != nil {
		b.Fatal(err)
	}
	for i := 0; i < b.N; i++ {
		if uuid.URN() == "" {
			b.Fatal("invalid uuid")
		}
	}
}
