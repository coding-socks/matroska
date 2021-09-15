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
)

var header = []byte(`// Code generated by go run make_doctype.go. DO NOT EDIT.

package matroska

import _ "embed"

`)

func main() {
	filename := "doctype.go"
	buf := bytes.NewBuffer(header)

	gen(buf)

	out, err := imports.Process(filename, buf.Bytes(), nil)
	if err != nil {
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
		fmt.Fprint(w, ")\n")
	})
	root.VisitAll(func(node *schema.TreeNode) {
		writeStruct(w, node)
	})
}

func writeID(w io.Writer, node *schema.TreeNode) {
	fmt.Fprintf(w, "\n\tID%s = %q", node.El.Name, node.El.ID)
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
	fmt.Fprintf(w, "type %s struct {", node.El.Name)
	if node.El.Recursive {
		fmt.Fprintf(w, "\n\t%[1]s *%[1]s", node.El.Name)
	}
	node.VisitAll(func(n *schema.TreeNode) {
		if n.El.MaxOccurs.Unbounded() || n.El.MaxOccurs.Val() > 1 {
			fmt.Fprintf(w, "\n\t%s []%s", n.El.Name, schema.ResolveGoType(n.El.Type, n.El.Name))
			return
		}
		// if !n.El.MaxOccurs.Unbounded() && n.El.MinOccurs == 0 && n.El.MaxOccurs.Val() == 1 {
		// 	fmt.Fprintf(w, "\n\t%s *%s", n.El.Name, schema.ResolveGoType(n.El.Type, n.El.Name))
		// 	return
		// }
		if n.El.ID == "0xE7" || n.El.ID == "0x2AD7B1" { // \Segment\Cluster\Timestamp, \Segment\Info\TimestampScale
			fmt.Fprintf(w, "\n\t%s time.Duration", n.El.Name)
			return
		}
		fmt.Fprintf(w, "\n\t%s %s", n.El.Name, schema.ResolveGoType(n.El.Type, n.El.Name))
	})
	fmt.Fprint(w, "\n}\n\n")
	node.VisitAll(func(n *schema.TreeNode) {
		writeStruct(w, n)
	})
}
