package main

import (
	"encoding/json"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"reflect"
	"strings"
	"text/template"
)

type tmplData struct {
	structs map[string][]*StructInfo
}

//StructInfo ... for MyApi and OtherApi
type StructInfo struct {
	Name  string
	Type  string
	Funcs []FuncInfo
}

//FuncInfo ... for create and get profile funcs
type FuncInfo struct {
	Name      string
	Parent    *StructInfo
	MarkData  Marks
	Parametrs []Parametr
}

//Marks ... for json marshiling
type Marks struct {
	URL    string `json:"url"`
	Auth   bool   `json:"auth"`
	Method string `json:"method"`
}

//Parametr ...
type Parametr struct {
	Name      string
	Type      string
	ValidData []Validator
}

//Validator ...
type Validator struct {
	Type      string
	Name      string
	Paramname string
	Required  bool
	Enum      []string
	Default   string
	Min       string
	Max       string
}

var (
	serveHTTPtmpl = template.Must(template.New("serveHTTPtmpl").Parse(`{{range $key, $value := .}}
func (srv *{{$key}}) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path { {{range $struct := $value}}{{range $funcInfo := $struct.Funcs}}
	case "{{$funcInfo.MarkData.URL}}":{{if $funcInfo.MarkData.Auth}}
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusNotAcceptable)
			data, _ := json.Marshal(resp{"error": "bad method"})
			w.Write(data)
			return
		}{{else}}
		if !(r.Method == http.MethodGet || r.Method == http.MethodPost) {
			w.WriteHeader(http.StatusNotAcceptable)
			data, _ := json.Marshal(resp{"error": "bad method"})
			w.Write(data)
			return
		}{{end}}
		srv.handle{{$key}}{{$funcInfo.Name}}(w, r){{end}}{{end}}
	default:
		w.WriteHeader(http.StatusNotFound)
		data, _ := json.Marshal(resp{"error": "unknown method"})
		w.Write(data)
		return
	}	
}
	{{end}}`))

	funcsTmpl = template.Must(template.New("funcsTmpl").Parse(`{{range $key, $value := .}}{{range $struct := $value}}{{range $funcInfo := $struct.Funcs}}
func (srv *{{$key}}) handle{{$key}}{{$funcInfo.Name}}(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json"){{if $funcInfo.MarkData.Auth}}
	if at := r.Header.Get("X-Auth"); at != "100500" {
		w.WriteHeader(http.StatusForbidden)
		data, _ := json.Marshal(resp{"error": "unauthorized"})
		w.Write(data)
		return
	}
	{{end}}{{range $parametr := $funcInfo.Parametrs}}{{range $validData := $parametr.ValidData}}{{if eq $validData.Type "int"}}
	{{$validData.Paramname}}_int := r.FormValue("{{$validData.Paramname}}")
	{{$validData.Paramname}},err := strconv.Atoi({{$validData.Paramname}}_int)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		data, _ := json.Marshal(resp{"error": "{{$validData.Paramname}} must be int"})
		w.Write(data)
		return
	}{{else}}
	{{$validData.Paramname}} := r.FormValue("{{$validData.Paramname}}"){{end}}{{if $validData.Required}}{{if eq $validData.Type "string"}}
	if {{$validData.Paramname}} == "" {{else}}if len("{{$validData.Paramname}}") == 0{{end}}{
		w.WriteHeader(http.StatusBadRequest)
		data, _ := json.Marshal(resp{"error": "{{$validData.Paramname}} must me not empty"})
		w.Write(data)
		return
	} 
	{{end}}{{if ne $validData.Default ""}}
	if {{$validData.Paramname}} == "" {
		{{$validData.Paramname}} = "{{$validData.Default}}"
	}
	{{end}}{{if ne $validData.Max ""}}{{if eq $validData.Type "string"}}
	if len({{$validData.Paramname}}){{else}}
	if {{$validData.Paramname}}{{end}} > {{$validData.Max}} {
		w.WriteHeader(http.StatusBadRequest)
		data, _ := json.Marshal(resp{"error": "{{$validData.Paramname}} must be <= {{$validData.Max}}"})
		w.Write(data)
		return
	}
	{{end}}{{if ne $validData.Min ""}}{{if eq $validData.Type "string"}}
	if len({{$validData.Paramname}}){{else}}if {{$validData.Paramname}}{{end}} < {{$validData.Min}} {
		w.WriteHeader(http.StatusBadRequest)
		data, _ := json.Marshal(resp{"error": "{{$validData.Paramname}}{{if eq $validData.Type "string"}} len{{end}} must be >= {{$validData.Min}}"})
		w.Write(data)
		return
	}
	{{end}}{{$len := len $validData.Enum}}{{if ne $len 0}}
	if !({{range $i,$e := $validData.Enum}}{{if ne $i 0}} || {{end}}{{$validData.Paramname}} == "{{$e}}"{{end}}) {
		w.WriteHeader(http.StatusBadRequest)
		data, _ := json.Marshal(resp{"error": "{{$validData.Paramname}} must be one of [{{range $i,$e := $validData.Enum}}{{if ne $i 0}}, {{end}}{{$e}}{{end}}]"})
		w.Write(data)
		return
	}{{end}}{{end}}
	{{$len := len $parametr.ValidData}}{{if ne $len 0}}
	{{$parametr.Name}} := {{$parametr.Type}} { {{range $validData := $parametr.ValidData}}
		{{$validData.Name}}: {{$validData.Paramname}},{{end}}
	}
	user,err := srv.{{$funcInfo.Name}}(context.Background(),{{$parametr.Name}})
	if err != nil {
		if v, ok := err.(ApiError); ok {
			w.WriteHeader(v.HTTPStatus)
			data, _ := json.Marshal(resp{"error": v.Error()})
			w.Write(data)
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		data, _ := json.Marshal(resp{"error": err.Error()})
		w.Write(data)
		return
	}

	response := map[string]interface{}{
		"error":    "",
		"response": user,
	}
	data, _ := json.Marshal(response)
	w.WriteHeader(http.StatusOK)
	w.Write(data)
	return{{end}}{{end}}
	
}{{end}}{{end}}{{end}}`))
)

// код писать тут
func main() {
	structs := make(map[string][]*StructInfo)
	validData := make(map[string][]Validator)

	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, os.Args[1], nil, parser.ParseComments)
	if err != nil {
		panic(err)
	}

	out, _ := os.Create(os.Args[2])

	fmt.Fprintln(out, `package `+node.Name.Name)
	fmt.Fprintln(out) // empty line
	fmt.Fprintln(out, `import (`+"\n\t"+`"context"`+"\n\t"+`"encoding/json"`+"\n\t"+`"net/http"`+"\n\t"+`"strconv"`+"\n)\n\ntype resp map[string]interface{}")
	fmt.Fprintln(out) // empty line

	for _, element := range node.Decls {
		fnc, ok := element.(*ast.FuncDecl)
		if !ok {

			//structs handler
			gen, ok2 := element.(*ast.GenDecl)
			if !ok2 {
				continue
			}

			for _, spec := range gen.Specs {
				currType, ok2 := spec.(*ast.TypeSpec)
				if !ok2 {
					continue
				}

				currStruct, ok2 := currType.Type.(*ast.StructType)
				if !ok2 {
					continue
				}

				var vdatas []Validator
				for _, field := range currStruct.Fields.List {
					if field.Tag != nil {
						tag := reflect.StructTag(field.Tag.Value[1 : len(field.Tag.Value)-1])
						apiv := tag.Get("apivalidator")
						if apiv != "" {
							if apiv == "-" {
								continue
							}

							validator := Validator{}
							validator.Type = field.Type.(*ast.Ident).Name
							validator.Name = field.Names[0].Name
							validator.Paramname = strings.ToLower(validator.Name)

							data := strings.Split(apiv, ",")
							for _, v := range data {
								elem := strings.Split(v, "=")
								switch elem[0] {
								case "required":
									validator.Required = true
								case "enum":
									for _, e := range strings.Split(elem[1], "|") {
										validator.Enum = append(validator.Enum, e)
									}
								case "default":
									validator.Default = elem[1]
								case "paramname":
									validator.Paramname = elem[1]
								case "min":
									validator.Min = elem[1]
								case "max":
									validator.Max = elem[1]
								}
							}
							vdatas = append(vdatas, validator)
						}
					}
				}
				validData[currType.Name.Name] = vdatas
			}
			continue
		}

		if fnc.Doc == nil {
			continue
		}

		//funcs handler
		for _, c := range fnc.Doc.List {
			if strings.HasPrefix(c.Text, "// apigen:api") {
				mark := Marks{}
				marks := strings.TrimLeft(c.Text, "// apigen:api ")
				if err := json.Unmarshal([]byte(marks), &mark); err != nil {
					panic(err)
				}

				structInfo := &StructInfo{}
				funcInfo := FuncInfo{}

				for _, v := range fnc.Recv.List {
					o, ok := v.Type.(*ast.StarExpr)
					if ok {
						structInfo.Name = v.Names[0].Name
					}
					z := o.X.(*ast.Ident)
					structInfo.Type = z.Name
				}

				for _, v := range fnc.Type.Params.List {
					parametr := Parametr{}
					parametr.Name = v.Names[0].Name
					indt, ok := v.Type.(*ast.Ident)
					if !ok {
						parametr.Type = "context.Background()"
					} else {
						parametr.Type = indt.Name
						parametr.ValidData = validData[indt.Name]

					}
					funcInfo.Parametrs = append(funcInfo.Parametrs, parametr)
				}

				funcInfo.Name = fnc.Name.Name
				funcInfo.Parent = structInfo
				funcInfo.MarkData = mark
				structInfo.Funcs = append(structInfo.Funcs, funcInfo)
				structs[structInfo.Type] = append(structs[structInfo.Type], structInfo)
			}
		}
	}

	for key, value := range structs {
		fmt.Printf("key: %v, value: %v\n", key, value)
		for _, v := range value {
			fmt.Printf("struct info: %v\n", v)
			for _, f := range v.Funcs {
				for _, vd := range f.Parametrs {
					fmt.Println(vd.ValidData)
				}
			}
		}
	}

	serveHTTPtmpl.Execute(out, structs)
	funcsTmpl.Execute(out, structs)
}
