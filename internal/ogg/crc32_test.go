package ogg

import (
	"io"
	"testing"
)

func TestSum32(t *testing.T) {
	c := NewCRC32()
	in := "123456789"
	io.WriteString(c, in)
	s := c.Sum32()
	if out := uint32(0x89a1897f); s != out {
		t.Fatalf("jones crc64(%s) = 0x%x want 0x%x", in, s, out)
	}
}
