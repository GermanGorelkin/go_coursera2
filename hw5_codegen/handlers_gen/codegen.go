package main

import (
	"fmt"
	"go/token"
	"go/parser"
	"log"
	"go/ast"
	"reflect"
	"strings"
	"encoding/json"
	"os"
	"text/template"
	"bytes"
	"io"
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

//--------
type tmpField struct {
	Name string
	Type string
	Required string
	Min string
	Max string
}

var (
	gmethods map[string]*GenFunc
	gtype map[string]*GenStruct
	recvType map[string]struct{}
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
	recvType[gfn.Recv]= struct{}{}

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
			ftags := ParseTags(tags)
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

func ParseTags(tags string) map[string]interface{} {
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

func ParseApi(file string){
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, file, nil, parser.ParseComments)
	if err != nil {
		log.Fatal(err)
	}

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
}

func GenHeader(out io.Writer){
	fmt.Fprintln(out, hTmp)
}

func GenServeHTTP(out io.Writer){
	for typeName := range recvType {
		var cases []string

		for _, fun := range gmethods {
			if fun.Recv != typeName {
				continue
			}
			if fun.ApiParams.Method == "" {
				var tpl bytes.Buffer
				data := struct {
					Url  string
					Name string
				}{
					Url:  fun.ApiParams.Url,
					Name: fun.Name,
				}
				if err := serveHTTPCase1Tpl.Execute(&tpl, data); err != nil {
					panic(err)
				}
				cases = append(cases, tpl.String())
			} else {
				var tpl bytes.Buffer
				data := struct {
					Url    string
					Name   string
					Method string
				}{
					Url:    fun.ApiParams.Url,
					Name:   fun.Name,
					Method: fun.ApiParams.Method,
				}
				if err := serveHTTPCase2Tpl.Execute(&tpl, data); err != nil {
					panic(err)
				}
				cases = append(cases, tpl.String())
			}

		}

		data := struct {
			Name  string
			Cases []string
		}{
			Name:  typeName,
			Cases: cases,
		}

		fmt.Fprintln(out)
		serveHTTPTpl.Execute(out, data)
	}
}

func GenWrappers(out io.Writer){
	for _, method := range gmethods {
		data := struct {
			Name   string
			Recv   string
			Auth   bool
			Fields []tmpField
		}{
			Name: method.Name,
			Recv: method.Recv,
			Auth: method.ApiParams.Auth,
		}

		//Valid
		par := method.Params[1]
		gt := gtype[par.Type]
		var fields []tmpField
		for _, f := range gt.Fileds {
			fname := strings.ToLower(f.Name)

			if v, ok := f.Tags["paramname"]; ok {
				fname = v.(string)
			}

			filed := tmpField{
				Name: fname,
				Type: f.Type,
			}
			var tpl bytes.Buffer

			// Required
			if _, ok := f.Tags["required"]; ok {
				data := struct {
					Name string
				}{
					Name: fname,
				}
				if err := validRequiredTpl.Execute(&tpl, data); err != nil {
					panic(err)
				}
				filed.Required = tpl.String()
				tpl.Reset()
			}

			// Min
			if val, ok := f.Tags["min"]; ok {
				data := struct {
					Name  string
					Value string
				}{
					Name:  fname,
					Value: val.(string),
				}
				if err := validMinTpl.Execute(&tpl, data); err != nil {
					panic(err)
				}
				filed.Min = tpl.String()
				tpl.Reset()
			}
			// Max
			if val, ok := f.Tags["max"]; ok {
				data := struct {
					Name  string
					Value string
				}{
					Name:  fname,
					Value: val.(string),
				}
				if err := validMaxTpl.Execute(&tpl, data); err != nil {
					panic(err)
				}
				filed.Max = tpl.String()
				tpl.Reset()
			}
			// Enum
			if val, ok := f.Tags["enum"]; ok {
				//map[default:warrior enum:[warrior sorcerer rouge]]

				df := f.Tags["default"].(string)

				data := struct {
					Name    string
					Default string
					Enum    []string
				}{
					Name:    fname,
					Default: df,
					Enum:    val.([]string),
				}

				if err := validEnumTpl.Execute(&tpl, data); err != nil {
					panic(err)
				}
				//fmt.Println(tpl.String())
				filed.Max = tpl.String()
				tpl.Reset()
			}
			//

			fields = append(fields, filed)
		}

		data.Fields = fields
		wrapperMethodTpl.Execute(out, data)
	}
}

func main() {
	gtype = make(map[string]*GenStruct)
	gmethods = make(map[string]*GenFunc)
	recvType = make(map[string]struct{})

	ParseApi(os.Args[1])

	out, _ := os.Create(os.Args[2])

	GenHeader(out)
	GenServeHTTP(out)
	GenWrappers(out)


	//fmt.Println("--types")
	//for _, v := range gtype{
	//	v.Print()
	//}
	//
	//fmt.Println("--methods")
	//for _, v := range gmethods{
	//	v.Print()
	//	fmt.Println()
	//}
}




var(

	serveHTTPTpl = template.Must(template.New("serveHTTPTpl").Parse(`
func (srv *{{.Name}}) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch {
	{{range .Cases}}
        {{.}}
    {{end}}
	default:
		{
			w.WriteHeader(http.StatusNotFound)
			w.Write(apiResponse("", fmt.Errorf("unknown method")))
		}
	}
}
`))

	serveHTTPCase1Tpl = template.Must(template.New("serveHTTPCase1Tpl").Parse(
		`case r.URL.Path == "{{.Url}}":
		srv.wrapper{{.Name}}(w, r)`))

	serveHTTPCase2Tpl = template.Must(template.New("serveHTTPCase2Tpl").Parse(
		`case r.URL.Path == "{{.Url}}":
		if r.Method == http.MethodPost {
			srv.wrapper{{.Name}}(w, r)
		} else {
			w.WriteHeader(http.StatusNotAcceptable)
			w.Write(apiResponse("", fmt.Errorf("bad method")))
		}`))

	validRequiredTpl = template.Must(template.New("validRequiredTpl").Parse(
	`if err := apiParRequired({{.Name}}, "{{.Name}}"); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write(apiResponse("", err))

		return
		}`))

	validMinTpl = template.Must(template.New("validMinTpl").Parse(
	`if err := apiParMin({{.Name}}, "{{.Name}}", {{.Value}}); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write(apiResponse("", err))

		return
		}`))

	validMaxTpl = template.Must(template.New("validMaxTpl").Parse(
	`if err := apiParMax({{.Name}}, "{{.Name}}", {{.Value}}); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write(apiResponse("", err))

		return
		}`))

	validEnumTpl = template.Must(template.New("validEnumTpl").Parse(
		`l{{.Name}} := map[string]struct{}{
	{{range .Enum}}
		"{{.}}":      {},
	{{end}}
	}
	{{.Name}} := r.FormValue("{{.Name}}")
	if {{.Name}} == "" {
		{{.Name}} = "{{.Default}}"
	}
	_, ok := l{{.Name}}[{{.Name}}]
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		w.Write(apiResponse("", fmt.Errorf("{{.Name}} must be one of {{.Enum}}")))
	
		return
	}`))

	wrapperMethodTpl = template.Must(template.New("wrapperMethodTpl").Parse(`
    func (srv *{{.Recv}}) wrapper{{.Name}}(w http.ResponseWriter, r *http.Request) {

	// auth
	{{if .Auth}}
	if r.Header.Get("X-Auth") != "100500" {
		w.WriteHeader(http.StatusForbidden)
		w.Write(apiResponse("", fmt.Errorf("unauthorized")))

		return
	}
    {{end}}


	// valid
	{{range .Fields}}
		{{if eq .Type "string"}}
            {{.Name}} := r.FormValue("{{.Name}}")
        {{else}}
            {{.Name}}, err := strconv.Atoi(r.FormValue("{{.Name}}"))
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				w.Write(apiResponse("", fmt.Errorf("{{.Name}} must be int")))

				return
			}
        {{end}}

        

		//Required
		{{.Required}}
		//Min
		{{.Min}}
		//Max
		{{.Max}}

    {{end}}

	// bl
	//ctx, _ := context.WithCancel(r.Context())
	//
	//in := CreateParams{
	//	Login:  login,
	//	Age:    age,
	//	Name:   full_name,
	//	Status: status,
	//}
	//u, err := srv.Create(ctx, in)
	//
	//if err != nil {
	//	switch ar := err.(type) {
	//	case ApiError:
	//		w.WriteHeader(ar.HTTPStatus)
	//	default:
	//		w.WriteHeader(http.StatusInternalServerError)
	//	}
	//
	//	w.Write(apiResponse("", err))
	//	return
	//}
	//
	//w.Write(apiResponse(u, err))
}`))

)

const
(hTmp = `package main

import (
	"net/http"
	"fmt"
	"encoding/json"
	"context"
	"strconv"
)


func apiResponse(data interface{}, err error) []byte {
	m := make(map[string]interface{})
	if err != nil {
		m["error"] = err.Error()
	} else
	{
		m["error"] = ""
		m["response"] = data
	}

	b, _ := json.Marshal(m)
	return b
}

func apiParRequired(val, name string) error {
	if val == "" {
		return fmt.Errorf("%s must me not empty", name)
	}
	return nil
}

func apiParMin(val interface{}, name string, num int) error {
	switch v := val.(type) {
	case string:
		{
			if len([]rune(v)) < num {
				return fmt.Errorf("%s len must be >= %d", name, num)
			}
		}
	case int:
		{
			if v < num {
				return fmt.Errorf("%s must be >= %d", name, num)
			}
		}
	}
	return nil
}

func apiParMax(val interface{}, name string, num int) error {
	switch v := val.(type) {
	case string:
		{
			if len([]rune(v)) > num {
				return fmt.Errorf("%s len must be <= %d", name, num)
			}
		}
	case int:
		{
			if v > num {
				return fmt.Errorf("%s must be <= %d", name, num)
			}
		}
	}
	return nil
}`)