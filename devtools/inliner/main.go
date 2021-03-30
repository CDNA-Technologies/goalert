package main

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"flag"
	"log"
	"os"
	"path/filepath"
	"sort"
	"text/template"

	"golang.org/x/tools/imports"
)

const typesTmplStr = `// Code generated by inliner DO NOT EDIT.

package {{.PackageName}}

import (
	"crypto/sha256"
	"encoding/hex"
	"sync"
)

type File struct {
	Name string
	Data func() []byte
	hash string
	hashCalc sync.Once
}

func (f *File) Hash256() string {
	f.hashCalc.Do(func(){
		h := sha256.New()
		h.Write(f.Data())
		f.hash = hex.EncodeToString(h.Sum(nil))
	})
	return f.hash
}

var Files []*File
`

const dataTmplStr = `// Code generated by inliner DO NOT EDIT.

package {{.PackageName}}

{{if .Files}}
func init() {
	var parseOnce sync.Once
	var data []byte

	dataStr := {{.Encoded}}
	dataRange := func(start, end int) (func() []byte) {
		return func() []byte {
			parseOnce.Do(func(){
				dec := base64.NewDecoder(
					base64.URLEncoding,
					bytes.NewBufferString(strings.Replace(dataStr, "\n", "", -1)),
				)
				r, err := gzip.NewReader(dec)
				if err != nil {
					panic(err)
				}
				defer r.Close()

				buf := new(bytes.Buffer)
				buf.Grow({{.Data.Len}})

				_, err = io.Copy(buf, r)
				if err != nil {
					panic(err)
				}

				data = buf.Bytes()
			})

			return data[start:end]
		}
	}


	Files = []*File{
		{{- range .Files}}
		{Data: dataRange({{.DataStart}},{{.DataEnd}}), Name: {{printf "%q" .Path}}},
		{{- end}}
	}
}
{{else}}
// No files
{{end}}
`

var (
	typesTmpl = template.Must(template.New("types").Parse(typesTmplStr))
	dataTmpl  = template.Must(template.New("data").Parse(dataTmplStr))
)

type file struct {
	Path      string
	DataStart int
	DataLen   int
}

func (f file) DataEnd() int {
	return f.DataStart + f.DataLen
}

func (dm *dataMap) Encoded() string {
	return "`\n" + dm.Data.String() + "`"
}

type dataMap struct {
	PackageName string
	Files       []*file
	Data        bytes.Buffer
}

func main() {
	packageName := flag.String("pkg", "", "Set the package name of the output file.")
	flag.Parse()

	if *packageName == "" {
		log.Fatal("pkg is required")
	}

	var m dataMap
	m.PackageName = *packageName

	for _, pattern := range flag.Args() {
		matches, err := filepath.Glob(pattern)
		if err != nil {
			log.Fatalf("glob pattern '%s': %v", pattern, err)
		}
		for _, match := range matches {
			f, err := os.Stat(match)
			if err != nil {
				log.Fatalf("stat file '%s': %v", match, err)
			}
			if f.IsDir() {
				continue
			}
			m.Files = append(m.Files, &file{
				Path: match,
			})
		}
	}
	sort.Slice(m.Files, func(i, j int) bool { return m.Files[i].Path < m.Files[j].Path })

	var n int
	lb := &lineBreaker{out: &m.Data}
	enc := base64.NewEncoder(base64.URLEncoding, lb)
	w := gzip.NewWriter(enc)
	for _, file := range m.Files {
		data, err := os.ReadFile(file.Path)
		if err != nil {
			log.Fatalf("read file '%s': %v", file.Path, err)
		}
		file.DataStart = n
		file.DataLen = len(data)
		n += file.DataLen
		w.Write(data)
	}
	w.Close()
	enc.Close()
	lb.Close()

	gen := func(filename string, tmpl *template.Template) {

		buf := bytes.NewBuffer(nil)
		err := tmpl.Execute(buf, &m)
		if err != nil {
			log.Fatal("render:", err)
		}

		data, err := imports.Process(filename, buf.Bytes(), nil)
		if err != nil {
			log.Fatal("format:", err)
		}

		err = os.WriteFile(filename, data, 0644)
		if err != nil {
			log.Fatal("save:", err)
		}
	}

	gen("inline_types_gen.go", typesTmpl)
	gen("inline_data_gen.go", dataTmpl)
}
