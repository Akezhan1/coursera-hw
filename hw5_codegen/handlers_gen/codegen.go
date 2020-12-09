package main

import (
	"encoding/json"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
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
	Name string
	Type string
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

	for _, f := range node.Decls {
		fnc, ok := f.(*ast.FuncDecl)
		if !ok {
			continue
		}

		if fnc.Doc == nil {
			continue
		}
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

				for i, v := range fnc.Type.Params.List {
					parametr := Parametr{}
					parametr.Name = v.Names[0].Name
					if i == 1 {
						parametr.Type = v.Type.(*ast.Ident).Name
					} else {
						parametr.Type = "context.Context"
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
	fmt.Println(len(structs))
}
