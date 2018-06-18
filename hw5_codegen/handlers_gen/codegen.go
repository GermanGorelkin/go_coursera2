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
	Url string
	Auth bool
	Method string
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

var (
	gmethods map[string]*GenFunc
	gtype map[string]*GenStruct
)

func processFuncDecl(fd *ast.FuncDecl) {
	return
	// name
	fmt.Println("----------func-------------")
	fmt.Printf("func name: %s\n", fd.Name.Name)

	// comments
	if fd.Doc == nil {
		fmt.Printf("func %#v doesnt have comments\n", fd.Name.Name)
	} else {
		fmt.Printf("func %#v comments:\n", fd.Name.Name)
		for _, comment := range fd.Doc.List {
			fmt.Printf("%s\n", comment.Text)
		}
	}

	// is methods
	if fd.Recv == nil {
		fmt.Printf("func %#v is not method\n", fd.Name.Name)
	}else {
		fmt.Printf("method %#v fields:\n", fd.Name.Name)
		for _, fn := range fd.Recv.List{
			switch v:=fn.Type.(type) {
			case *ast.StarExpr:
				{
					fmt.Printf("type(*ast.StarExpr) name: %s\n", v.X)
				}
			case *ast.Ident:
				{
					fmt.Printf("type(*ast.Ident) name: %s\n", v.Name)
				}

			}
		}
	}

	// params
	fmt.Printf("func %#v params:\n", fd.Name.Name)
	for _, fn := range fd.Type.Params.List{
		switch v:=fn.Type.(type) {
		case *ast.SelectorExpr:
			{
				fmt.Printf("param name %s type(*ast.SelectorExpr) name: %s\n", fn.Names, v.Sel.Name)
			}
		case *ast.Ident:
			{
				fmt.Printf("param name %s type(*ast.Ident) name: %s\n", fn.Names, v.Name)
			}
		}
	}

	fmt.Println("-----------------------")
	fmt.Println()
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

		gs := &GenStruct{
			Name: currType.Name.Name,
		}
		lf := make([]*GenStructFiled, len(currStruct.Fields.List))

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
			lf[i] = &GenStructFiled{
				Name: field.Names[0].Name,
				Type: fmt.Sprintf("%s", field.Type),
				Tags: ftags,
			}
		}

		gs.Fileds = lf
		gtype[gs.Name] = gs
	}

	fmt.Println()
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


	for _, v := range gtype{
		v.Print()
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