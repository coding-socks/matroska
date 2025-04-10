//go:build codegen
// +build codegen

// This program generates doctype.go.

package main

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"github.com/coding-socks/ebml/schema"
	"golang.org/x/tools/imports"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"unicode"
)

var header = []byte(`// Code generated by go run make_doctype.go. DO NOT EDIT.

package matroska

import (
	_ "embed"
	"time"

	"github.com/coding-socks/ebml/schema"
)

`)

func main() {
	filename := "doctype.go"
	buf := bytes.NewBuffer(header)

	gen(buf)

	out, err := imports.Process(filename, buf.Bytes(), nil)
	if err != nil {
		io.Copy(os.Stdout, bytes.NewReader(buf.Bytes()))
		log.Fatal(err)
	}

	err = ioutil.WriteFile(filename, out, 0666)
	if err != nil {
		log.Fatal(err)
	}
}

func gen(w io.Writer) {
	root := schema.NewTreeNode(schema.Element{
		Type: schema.TypeMaster,
		Name: "Document",
	})
	fp := filepath.Join(".", "ebml_matroska.xml")
	var s schema.Schema
	func() {
		f, err := os.Open(fp)
		if err != nil {
			log.Fatal(err)
		}
		defer f.Close()
		if err := xml.NewDecoder(f).Decode(&s); err != nil {
			log.Fatal(err)
		}
	}()
	for _, el := range s.Elements {
		if strings.HasPrefix(el.Path, `\EBML`) {
			continue
		}
		p := strings.Split(el.Path, `\`)[1:]
		branch := root
		lastIndex := len(p) - 1
		for _, s := range p[:lastIndex] {
			node := branch.Get(s)
			if node == nil {
				node = schema.NewTreeNode(el)
				branch.Put(s, node)
			}
			branch = node
		}
		branch.Put(p[lastIndex], schema.NewTreeNode(el))
	}
	fmt.Fprint(w, "//go:embed ebml_matroska.xml\n")
	fmt.Fprint(w, "var docType []byte\n")
	root.VisitAll(func(node *schema.TreeNode) {
		fmt.Fprint(w, "\nvar (")
		writeID(w, node)
		fmt.Fprint(w, "\n)\n\n")
	})
	root.VisitAll(func(node *schema.TreeNode) {
		writeStruct(w, node)
	})
}

func writeID(w io.Writer, node *schema.TreeNode) {
	fmt.Fprintf(w, "\n\tID%s schema.ElementID = %v", node.El.Name, node.El.ID)
	if node.El.Type != schema.TypeMaster {
		return
	}
	node.VisitAll(func(n *schema.TreeNode) {
		writeID(w, n)
	})
}

func writeStruct(w io.Writer, node *schema.TreeNode) {
	if node.El.Type != schema.TypeMaster {
		return
	}
	node.VisitAll(func(n *schema.TreeNode) {
		if n.El.ID == 0x63CA || n.El.ID == 0x68CA {
			// TODO: figure out what TargetType and TargetTypeValue represents.
			//  https://www.ietf.org/archive/id/draft-ietf-cellar-matroska-14.html#name-targettypevalue-element
			//  https://www.ietf.org/archive/id/draft-ietf-cellar-matroska-14.html#name-targettype-element
			return
		}
		if n.El.Restriction != nil {
			fmt.Fprint(w, "const (")
			for _, enum := range n.El.Restriction.Enum {
				val := enum.Value
				if n.El.Type == "string" {
					val = fmt.Sprintf("%q", val)
				}
				for _, label := range strings.Split(enum.Label, " / ") {
					fmt.Fprintf(w, "\n\t%s%s = %s", n.El.Name, studlyCase(label), val)
				}
			}
			fmt.Fprint(w, "\n)\n\n")
		}
	})
	fmt.Fprintf(w, "type %s struct {", node.El.Name)
	if node.El.Recursive {
		fmt.Fprintf(w, "\n\t%[1]s *%[1]s", node.El.Name)
	}
	node.VisitAll(func(n *schema.TreeNode) {
		if n.El.MaxOccurs.Unbounded() || n.El.MaxOccurs.Val() > 1 {
			fmt.Fprintf(w, "\n\t%s []%s", n.El.Name, schema.ResolveGoType(n.El.Type, n.El.Name))
			return
		}
		if n.El.ID == 0xE7 || n.El.ID == 0x2AD7B1 { // \Segment\Cluster\Timestamp, \Segment\Info\TimestampScale
			fmt.Fprintf(w, "\n\t%s time.Duration", n.El.Name)
			return
		}
		if n.El.ID == 0x53AB { // \Segment\SeekHead\Seek\SeekID
			fmt.Fprintf(w, "\n\t%s schema.ElementID", n.El.Name)
			return
		}
		if !n.El.MaxOccurs.Unbounded() && n.El.MinOccurs == 0 && n.El.MaxOccurs.Val() == 1 && n.El.Default == nil {
			fmt.Fprintf(w, "\n\t%s *%s", n.El.Name, schema.ResolveGoType(n.El.Type, n.El.Name))
			return
		}
		fmt.Fprintf(w, "\n\t%s %s", n.El.Name, schema.ResolveGoType(n.El.Type, n.El.Name))
	})
	fmt.Fprint(w, "\n}\n\n")
	node.VisitAll(func(n *schema.TreeNode) {
		writeStruct(w, n)
	})
}

// studlyCase converts a value to studly caps case.
func studlyCase(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return s
	}

	var n strings.Builder
	capNext := true
	for _, r := range s {
		if unicode.IsLetter(r) {
			if capNext && unicode.IsLower(r) {
				r = unicode.ToUpper(r)
			}
			n.WriteRune(r)
			capNext = false
		} else if unicode.IsNumber(r) {
			n.WriteRune(r)
			capNext = true
		} else {
			capNext = true
		}
	}
	return n.String()
}

func isNumeric(s string) bool {
	if len(s) > 2 && len(s)%2 == 0 && strings.HasPrefix(s, "0x") {
		for _, r := range s {
			if !unicode.IsNumber(r) || ('a' <= r && r <= 'f') {
				return false
			}
		}
		return true
	}
	for _, r := range s {
		if !unicode.IsNumber(r) {
			return false
		}
	}
	return true
}
