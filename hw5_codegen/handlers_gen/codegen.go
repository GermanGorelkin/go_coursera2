package main

import (
	"fmt"
	"os"
	"go/token"
	"go/parser"
	"log"
	"go/ast"
	"reflect"
	"strings"
	"encoding/json"
)

//----type
type GenStruct struct{
	Name string
	Fileds []*GenStructFiled
}

type GenStructFiled struct {
	Name string
	Type string
	Tags map[string]interface{}
}

func (gs GenStruct) Print(){
	fmt.Printf("Struct name: %s\n", gs.Name)
	for _,v := range gs.Fileds{
		fmt.Printf("\t %s %s %s\n", v.Name, v.Type, v.Tags)
	}
}

//---methods
type GenFuncApiParam struct{
	Url string `json:"url"`
	Auth bool `json:"auth"`
	Method string `json:"method,omitempty"`
}

type GenFuncParam struct {
	Name string
	Type string
}

type GenFunc struct {
	Name string
	Recv string
	Params []*GenFuncParam
	ApiParams *GenFuncApiParam
}

func (gf GenFunc) Print(){
	fmt.Printf("Func name: %s\n", gf.Name)
	fmt.Printf("Func recv: %s\n", gf.Recv)
	fmt.Printf("Func ApiParams: %s\n", gf.ApiParams)
	fmt.Println("Func Params:")
	for _,v := range gf.Params{
		fmt.Printf("\t %s %s\n", v.Name, v.Type)
	}
}

var (
	gmethods map[string]*GenFunc
	gtype map[string]*GenStruct
)

func processFuncDecl(fd *ast.FuncDecl) {

	// name
	//fmt.Println("----------func-------------")
	//fmt.Printf("func name: %s\n", fd.Name.Name)

	gfn := &GenFunc{
		Name:      fd.Name.Name,
		ApiParams: &GenFuncApiParam{},
	}

	// comments
	if fd.Doc == nil {
		//fmt.Printf("func %#v doesnt have comments\n", fd.Name.Name)
		return
	} else {
		//fmt.Printf("func %#v comments:\n", fd.Name.Name)
		for _, comment := range fd.Doc.List {
			//fmt.Printf("%s\n", comment.Text)
			apt := comment.Text[strings.Index(comment.Text, "{"):len(comment.Text)]
			//fmt.Printf("%s\n", apt)
			json.Unmarshal([]byte(apt), gfn.ApiParams)
		}
	}

	// is methods
	if fd.Recv == nil {
		return
		//fmt.Printf("func %#v is not method\n", fd.Name.Name)
	} else {
		//fmt.Printf("method %#v fields:\n", fd.Name.Name)
		for _, fn := range fd.Recv.List {
			switch v := fn.Type.(type) {
			case *ast.StarExpr:
				{
					//fmt.Printf("type(*ast.StarExpr) name: %s\n", v.X)
					gfn.Recv = fmt.Sprintf("%s", v.X)
				}
			case *ast.Ident:
				{
					//fmt.Printf("type(*ast.Ident) name: %s\n", v.Name)
					gfn.Recv = v.Name
				}

			}
		}
	}

	// params
	//fmt.Printf("func %#v params:\n", fd.Name.Name)
	gfn.Params = make([]*GenFuncParam, len(fd.Type.Params.List))
	for i, fn := range fd.Type.Params.List {
		switch v := fn.Type.(type) {
		case *ast.SelectorExpr:
			{
				gfn.Params[i] = &GenFuncParam{
					Name: fn.Names[0].Name,
					Type: v.Sel.Name,
				}
				//fmt.Printf("param name %s type(*ast.SelectorExpr) name: %s\n", fn.Names, v.Sel.Name)
			}
		case *ast.Ident:
			{
				gfn.Params[i] = &GenFuncParam{
					Name: fn.Names[0].Name,
					Type: v.Name,
				}
				//fmt.Printf("param name %s type(*ast.Ident) name: %s\n", fn.Names, v.Name)
			}
		}
	}

	gmethods[gfn.Recv+"."+gfn.Name] = gfn

	//fmt.Println("-----------------------")
	//fmt.Println()
}

func processGenDecl(gd *ast.GenDecl) {
SPECS_LOOP:
	for _, spec := range gd.Specs {
		currType, ok := spec.(*ast.TypeSpec)
		if !ok {
			continue
		}
		currStruct, ok := currType.Type.(*ast.StructType)
		if !ok {
			continue
		}

		gstruct := &GenStruct{
			Name: currType.Name.Name,
		}
		fileds := make([]*GenStructFiled, len(currStruct.Fields.List))

		if len(currStruct.Fields.List) == 0 {
			continue SPECS_LOOP
		}

		for i, field := range currStruct.Fields.List {
			if field.Tag == nil {
				continue SPECS_LOOP
			}

			tag := reflect.StructTag(field.Tag.Value[1 : len(field.Tag.Value)-1])
			tags := tag.Get("apivalidator")
			if tags == "" {
				continue SPECS_LOOP
			}
			ftags := ParserTags(tags)
			fileds[i] = &GenStructFiled{
				Name: field.Names[0].Name,
				Type: fmt.Sprintf("%s", field.Type),
				Tags: ftags,
			}
		}

		gstruct.Fileds = fileds
		gtype[gstruct.Name] = gstruct
	}
}

func ParserTags(tags string) map[string]interface{} {
	var name string
	var val interface{}
	r := make(map[string]interface{})
	for _, v := range  strings.Split(tags, ","){
		l:=strings.Split(v, "=")
		name = l[0]
		if len(l)>1{
			val = l[1]

			if name == "enum"{
				val = strings.Split(val.(string), "|")
			}

		}else{
			val = "true"
		}
		r[name]=val
	}
	return r
}

func main() {
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, os.Args[1], nil, parser.ParseComments)
	if err != nil {
		log.Fatal(err)
	}

	// Print the AST.
	//ast.Print(fset, node)

	//out, _ := os.Create(os.Args[2])
	//defer out.Close()

	gtype = make(map[string]*GenStruct)
	gmethods = make(map[string]*GenFunc)

	for _, f := range node.Decls {
		switch v := f.(type) {
		case *ast.FuncDecl:
			{
				processFuncDecl(v)
			}
		case *ast.GenDecl:
			{
				processGenDecl(v)
			}
		}
	}


	//for _, v := range gtype{
	//	v.Print()
	//}

	for _, v := range gmethods{
		v.Print()
		fmt.Println()
	}

		//g, ok := f.(*ast.GenDecl)
		//if !ok {
		//	fmt.Printf("SKIP %T is not *ast.GenDecl\n", f)
		//	continue
		//}
	//SPECS_LOOP:
	//	for _, spec := range g.Specs {
	//		currType, ok := spec.(*ast.TypeSpec)
	//		if !ok {
	//			fmt.Printf("SKIP %T is not ast.TypeSpec\n", spec)
	//			continue
	//		}
	//
	//		currStruct, ok := currType.Type.(*ast.StructType)
	//		if !ok {
	//			fmt.Printf("SKIP %T is not ast.StructType\n", currStruct)
	//			continue
	//		}

			//if g.Doc == nil {
			//	fmt.Printf("SKIP struct %#v doesnt have comments\n", currType.Name.Name)
			//	continue
			//}
			//
			//needCodegen := false
			//for _, comment := range g.Doc.List {
			//	needCodegen = needCodegen || strings.HasPrefix(comment.Text, "// cgen: binpack")
			//}
			//if !needCodegen {
			//	fmt.Printf("SKIP struct %#v doesnt have cgen mark\n", currType.Name.Name)
			//	continue SPECS_LOOP
			//}

			//fmt.Printf("process struct %s\n", currType.Name.Name)

			//fmt.Fprintln(out, "func (in *"+currType.Name.Name+") Unpack(data []byte) error {")
			//fmt.Fprintln(out, "	r := bytes.NewReader(data)")

		//FIELDS_LOOP:
			//for _, field := range currStruct.Fields.List {

				//if field.Tag != nil {
				//	tag := reflect.StructTag(field.Tag.Value[1 : len(field.Tag.Value)-1])
				//	if tag.Get("cgen") == "-" {
				//		continue FIELDS_LOOP
				//	}
				//}

				//fieldName := field.Names[0].Name
				//fileType := field.Type.(*ast.Ident).Name

				//fmt.Printf("\tgenerating code for field %s.%s\n", currType.Name.Name, fieldName)

				//switch fileType {
				//case "int":
				//	intTpl.Execute(out, tpl{fieldName})
				//case "string":
				//	strTpl.Execute(out, tpl{fieldName})
				//default:
				//	log.Fatalln("unsupported", fileType)
				//}
			//}
		//}
	//}
}