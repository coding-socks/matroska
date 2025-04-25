package riff

import (
	"bytes"
	"io"
	"testing"
)

func Test(t *testing.T) {
	var buf bytes.Buffer
	lw, err := NewWriter(&buf, FourCC{'T', 'E', 'S', 'T'})
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
		if err := lw.Close(); err != nil {
			t.Fatal(err)
		}
	}
	{ // ck02
		w, err := lw.Next(FourCC{'c', 'k', '0', '2'})
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

	id, lr, err := NewReader(&buf)
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
		if _, _, _, err := lr.Next(); err != io.EOF {
			t.Errorf("Next() got %v, want EOF", err)
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
		if got, want := b, []byte("Hello world!"); !bytes.Equal(got, want) {
			t.Errorf("Read() got %s, want %s", got, want)
		}
	}
	if _, _, _, err := lr.Next(); err != io.EOF {
		t.Errorf("Next() got %v, want EOF", err)
	}
}
