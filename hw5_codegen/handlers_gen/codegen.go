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
)

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
	ValidData Validator
}

//Validator ...
type Validator struct {
	RequiredField string
	MinReqField   int
	Enum          []string
	Default       string
	Min           int
	Max           int
}

// код писать тут
func main() {
	structs := make(map[string][]*StructInfo)

	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, os.Args[1], nil, parser.ParseComments)
	if err != nil {
		panic(err)
	}

	out, _ := os.Create(os.Args[2])

	fmt.Fprintln(out, `package `+node.Name.Name)
	fmt.Fprintln(out) // empty line
	fmt.Fprintln(out, `import (`+"\n\t"+`"context"`+"\n\t"+`"encoding/json"`+"\n\t"+`"net/http"`+"\n\t"+`"strconv"`+"\n"+`)`)
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

				for _, field := range currStruct.Fields.List {
					if field.Tag != nil {
						tag := reflect.StructTag(field.Tag.Value[1 : len(field.Tag.Value)-1])
						if tag.Get("apivalidator") != "" {
							//fmt.Println(currStruct.Fields.List[0].Names[0].Name)
							for _, list := range currStruct.Fields.List {
								for _, name := range list.Names {
									fmt.Println(currType.Name, name.Name)
								}
							}
						}
					}
				}
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
						parametr.Type = "context.Context"
					} else {
						parametr.Type = indt.Name

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
		}
	}
}
