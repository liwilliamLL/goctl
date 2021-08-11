package generator

import (
	"fmt"
	"william/base/go-zero/core/collection"
	"william/base/go-zero/tools/goctl/util/stringx"
	"path/filepath"
	"strings"

	conf "william/base/go-zero/tools/goctl/config"
	"william/base/go-zero/tools/goctl/rpc/parser"
	"william/base/go-zero/tools/goctl/util"
	"william/base/go-zero/tools/goctl/util/format"
)

const httpTemplate = `{{.head}}

package server

import (
	"context"
	"github.com/gin-gonic/gin"
	"net/http"
	{{.imports}}
)




func (s *{{.server}}Server)CreateRouter(router *gin.Engine,c config.Config)(){
	if c.HttpStatus {
        {{.routers}}
		router.Run(c.HttpPort)
	}
}


{{.funcs}}
`


const httpFuncTemplate = `
{{if .hasComment}}{{.comment}}{{end}}

func (s  *{{.server}}Server) {{.method}}Http (ctx *gin.Context){
	var in {{.request}}
	err := ctx.BindJSON(&in)
	if err!=nil{
		return
	}
	var rsp {{.response}}
	rsp ,err = s.{{.method}}(ctx,in)
	if err!=nil{
		ctx.JSON(http.StatusOK,err)
		return
	}
	ctx.JSON(http.StatusOK, rsp)
	return
}
`

const httpRouterTemplate = `
{{if .hasComment}}{{.comment}}{{end}}
router.Any("/{{.method}}", s.{{.method}}Http)
`

// GenSvc generates the servicecontext.go file, which is the resource dependency of a service,
// such as rpc dependency, model dependency, etc.
func (g *DefaultGenerator) GenHttp(ctx DirContext, proto parser.Proto, cfg *conf.Config) error {
	dir := ctx.GetServer()
	pbImport := fmt.Sprintf(`"%v"`, ctx.GetPb().Package)
	logImport := fmt.Sprintf(`"%v"`, ctx.GetConfig().Package)

	imports := collection.NewSet()
	imports.AddStr(pbImport,logImport)

	head := util.GetHead(proto.Name)
	service := proto.Service
	serverFilename, err := format.FileNamingFormat(cfg.NamingFormat, "http")
	if err != nil {
		return err
	}

	serverFile := filepath.Join(dir.Filename, serverFilename+".go")
	funcList, err := g.genHttpFunctions(proto.PbPackage, service)
	if err != nil {
		return err
	}

	routerList ,err := g.genRoutersFunctions(proto.PbPackage, service)
	if err != nil {
		return err
	}


	text, err := util.LoadTemplate(category, httpTemplateFile, httpTemplate)
	if err != nil {
		return err
	}

	err = util.With("server").GoFmt(true).Parse(text).SaveTo(map[string]interface{}{
		"head":    head,
		"server":  stringx.From(service.Name).ToCamel(),
		"imports": strings.Join(imports.KeysStr(), util.NL),
		"routers": strings.Join(routerList, util.NL),
		"funcs":   strings.Join(funcList, util.NL),
	}, serverFile, true)
	return err
	//
	//
	//
	//
	//svcFilename, err := format.FileNamingFormat(cfg.NamingFormat, "http")
	//if err != nil {
	//	return err
	//}
	//
	//fileName := filepath.Join(dir.Filename, svcFilename+".go")
	//text, err := util.LoadTemplate(category, httpTemplateFile, httpTemplate)
	//if err != nil {
	//	return err
	//}
	//
	//return util.With("svc").GoFmt(true).Parse(text).SaveTo(map[string]interface{}{
	//	"imports": fmt.Sprintf(`"%v"`, ctx.GetConfig().Package),
	//}, fileName, false)
}


func (g *DefaultGenerator) genRoutersFunctions(goPackage string, service parser.Service) ([]string, error) {
	var functionList []string
	for _, rpc := range service.RPC {
		text, err := util.LoadTemplate(category, httpRouterTemplateFile, httpRouterTemplate)
		if err != nil {
			return nil, err
		}

		comment := parser.GetComment(rpc.Doc())
		buffer, err := util.With("func").Parse(text).Execute(map[string]interface{}{
			"server":     stringx.From(service.Name).ToCamel(),
			"logicName":  fmt.Sprintf("%sLogic", stringx.From(rpc.Name).ToCamel()),
			"method":     parser.CamelCase(rpc.Name),
			"request":    fmt.Sprintf("*%s.%s", goPackage, parser.CamelCase(rpc.RequestType)),
			"response":   fmt.Sprintf("*%s.%s", goPackage, parser.CamelCase(rpc.ReturnsType)),
			"hasComment": len(comment) > 0,
			"comment":    comment,
		})
		if err != nil {
			return nil, err
		}

		functionList = append(functionList, buffer.String())
	}
	return functionList, nil
}

func (g *DefaultGenerator) genHttpFunctions(goPackage string, service parser.Service) ([]string, error) {
	var functionList []string
	for _, rpc := range service.RPC {
		text, err := util.LoadTemplate(category, httpFuncTemplateFile, httpFuncTemplate)
		if err != nil {
			return nil, err
		}

		comment := parser.GetComment(rpc.Doc())
		buffer, err := util.With("func").Parse(text).Execute(map[string]interface{}{
			"server":     stringx.From(service.Name).ToCamel(),
			"logicName":  fmt.Sprintf("%sLogic", stringx.From(rpc.Name).ToCamel()),
			"method":     parser.CamelCase(rpc.Name),
			"request":    fmt.Sprintf("*%s.%s", goPackage, parser.CamelCase(rpc.RequestType)),
			"response":   fmt.Sprintf("*%s.%s", goPackage, parser.CamelCase(rpc.ReturnsType)),
			"hasComment": len(comment) > 0,
			"comment":    comment,
		})
		if err != nil {
			return nil, err
		}

		functionList = append(functionList, buffer.String())
	}
	return functionList, nil
}
