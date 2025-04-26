package riff

import (
	"bytes"
	"io"
	"testing"
)

func Test(t *testing.T) {
	var writeAt bytesWriterAt
	lw, err := NewWriter(&writeAt, FourCC{'T', 'E', 'S', 'T'})
	if err != nil {
		t.Fatal(err)
	}
	{ // LIST typ1
		w, err := lw.Next(LIST)
		if err != nil {
			t.Fatal(err)
		}
		lw, err := NewListWriter(w, FourCC{'t', 'y', 'p', '1'})
		if err != nil {
			t.Fatal(err)
		}
		{ // ck01
			w, err := lw.Next(FourCC{'c', 'k', '0', '1'})
			if err != nil {
				t.Fatal(err)
			}
			if _, err := w.Write([]byte("Hello world!")); err != nil {
				t.Fatal(err)
			}
		}
		{ // ck02
			w, err := lw.Next(FourCC{'c', 'k', '0', '2'})
			if err != nil {
				t.Fatal(err)
			}
			if _, err := w.Write([]byte("Hello world")); err != nil {
				t.Fatal(err)
			}
		}
		if err := lw.Close(); err != nil {
			t.Fatal(err)
		}
	}
	{ // ck03
		w, err := lw.Next(FourCC{'c', 'k', '0', '3'})
		if err != nil {
			t.Fatal(err)
		}
		if _, err := w.Write([]byte("Hello world!")); err != nil {
			t.Fatal(err)
		}
	}
	if err := lw.Close(); err != nil {
		t.Fatal(err)
	}

	id, lr, err := NewReader(bytes.NewReader(writeAt.buf))
	if err != nil {
		t.Fatal(err)
	}
	if got, want := id, (FourCC{'T', 'E', 'S', 'T'}); got != want {
		t.Errorf("NewReader() got %v, want %v", got, want)
	}
	{ // LIST typ1
		id, l, r, err := lr.Next()
		if err != nil {
			t.Fatal(err)
		}
		if got, want := id, (LIST); got != want {
			t.Errorf("Next() got %v, want %v", got, want)
		}
		listType, lr, err := NewListReader(l, r)
		if err != nil {
			t.Fatal(err)
		}
		if got, want := listType, (FourCC{'t', 'y', 'p', '1'}); got != want {
			t.Errorf("NewListReader() got %v, want %v", got, want)
		}
		{ // ck01
			id, l, r, err := lr.Next()
			if err != nil {
				t.Fatal(err)
			}
			if got, want := id, (FourCC{'c', 'k', '0', '1'}); got != want {
				t.Errorf("Next() got %v, want %v", got, want)
			}
			b := make([]byte, l)
			if _, err := r.Read(b); err != nil {
				t.Fatal(err)
			}
			if got, want := b, []byte("Hello world!"); !bytes.Equal(got, want) {
				t.Errorf("Read() got %s, want %s", got, want)
			}
		}
		{ // ck02
			id, l, r, err := lr.Next()
			if err != nil {
				t.Fatal(err)
			}
			if got, want := id, (FourCC{'c', 'k', '0', '2'}); got != want {
				t.Errorf("Next() got %v, want %v", got, want)
			}
			b := make([]byte, l)
			if _, err := r.Read(b); err != nil {
				t.Fatal(err)
			}
			if got, want := b, []byte("Hello world"); !bytes.Equal(got, want) {
				t.Errorf("Read() got %s, want %s", got, want)
			}
		}
		if _, _, _, err := lr.Next(); err != io.EOF {
			t.Errorf("Next() got %v, want EOF", err)
		}
	}
	{ // ck03
		id, l, r, err := lr.Next()
		if err != nil {
			t.Fatal(err)
		}
		if got, want := id, (FourCC{'c', 'k', '0', '3'}); got != want {
			t.Errorf("Next() got %v, want %v", got, want)
		}
		b := make([]byte, l)
		if _, err := r.Read(b); err != nil {
			t.Fatal(err)
		}
		if got, want := b, []byte("Hello world!"); !bytes.Equal(got, want) {
			t.Errorf("Read() got %s, want %s", got, want)
		}
	}
	if _, _, _, err := lr.Next(); err != io.EOF {
		t.Errorf("Next() got %v, want EOF", err)
	}
}

type bytesWriterAt struct {
	buf []byte
}

func (w *bytesWriterAt) grow(n int64) {
	if n > int64(cap(w.buf)) {
		m := max(int64(cap(w.buf))*2, 8)
		for m < n {
			m *= 2
		}
		tmp := make([]byte, m)
		copy(tmp, w.buf)
		w.buf = tmp[:n]
	}
	if n > int64(len(w.buf)) {
		w.buf = w.buf[:n]
	}
}

func (b *bytesWriterAt) WriteAt(p []byte, off int64) (n int, err error) {
	b.grow(off + int64(len(p)))
	copy(b.buf[off:off+int64(len(p))], p)
	return len(p), nil
}
