// The following directive is necessary to make the package coherent:
// This program generates types, It can be invoked by running
// go generate
package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"sort"
	"text/template"
)

type fault struct {
	Name    string   `json:"-"`
	Value   []string `json:"value,omitempty"`
	Message string   `json:"message"`
}

func main() {
	const filePath = "errors.go"

	data, err := ioutil.ReadFile("types.json")
	if err != nil {
		log.Fatal(err)
	}
	types := make(map[string]string)
	err = json.Unmarshal(data, &types)
	if err != nil {
		log.Fatal(err)
	}

	funcs := template.FuncMap{
		"lowerFirst": func(s string) string {
			if len(s) == 0 {
				return ""
			}
			if s[0] < 'A' || s[0] > 'Z' {
				return s
			}
			return string(s[0]+'a'-'A') + s[1:]
		},
		"getType": func(s string) string {
			return types[s]
		},
	}

	data, err = ioutil.ReadFile("faults.json")
	if err != nil {
		log.Fatal(err)
	}
	faults := make(map[string]*fault)
	err = json.Unmarshal(data, &faults)
	if err != nil {
		log.Fatal(err)
	}

	for k, v := range faults {
		sort.Strings(faults[k].Value)

		v.Name = k
	}

	// Format input tasks.json
	data, err = json.MarshalIndent(faults, "", "  ")
	if err != nil {
		log.Fatal(err)
	}
	err = ioutil.WriteFile("faults.json", data, 0664)
	if err != nil {
		log.Fatal(err)
	}

	f, err := os.Create(filePath)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	err = template.Must(template.New("").Funcs(funcs).Parse(tmpl)).Execute(f, faults)
	if err != nil {
		log.Fatal(err)
	}
}

var tmpl = `// Code generated by go generate; DO NOT EDIT.
package types

import (
	"fmt"
)

{{- range $_, $fault := . }}
type {{$fault.Name}} struct {
	err error

{{- range $k, $v := $fault.Value }}
	{{$v}}
{{- end }}
}

func (f *{{$fault.Name}}) Error() string {
	return fmt.Sprintf({{$fault.Message}})
}

func (f *{{$fault.Name}}) Unwrap() error {
	return f.err
}

func NewErr{{$fault.Name}}(err error{{- range $k, $v := $fault.Value }},{{$v | lowerFirst}} {{$v | getType}}{{- end }}) error {
	f := &{{$fault.Name}}{}
	f.err=err
{{- range $k, $v := $fault.Value }}
	f.Set{{$v}}({{$v | lowerFirst}})
{{- end }}

	return f
}
{{- end }}
`
