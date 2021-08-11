package gen

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	"william/base/go-zero/tools/goctl/config"
	"william/base/go-zero/tools/goctl/model/sql/model"
	"william/base/go-zero/tools/goctl/model/sql/parser"
	"william/base/go-zero/tools/goctl/model/sql/template"
	modelutil "william/base/go-zero/tools/goctl/model/sql/util"
	"william/base/go-zero/tools/goctl/util"
	"william/base/go-zero/tools/goctl/util/console"
	"william/base/go-zero/tools/goctl/util/format"
	"william/base/go-zero/tools/goctl/util/stringx"
)

const (
	pwd             = "."
	createTableFlag = `(?m)^(?i)CREATE\s+TABLE` // ignore case
)

type (
	defaultGenerator struct {
		// source string
		dir string
		console.Console
		pkg   string
		cfg   *config.Config
		proto string
		ppack string
	}

	// Option defines a function with argument defaultGenerator
	Option func(generator *defaultGenerator)

	code struct {
		importsCode string
		varsCode    string
		typesCode   string
		newCode     string
		insertCode  string
		findCode    []string
		updateCode  string
		deleteCode  string
		cacheExtra  string
	}
)

// NewDefaultGenerator creates an instance for defaultGenerator
func NewDefaultGenerator(dir, proto, ppack string, cfg *config.Config, opt ...Option) (*defaultGenerator, error) {

	for _, d := range []*string{&dir, &proto} {
		if *d == "" {
			*d = pwd
		}
		dirAbs, err := filepath.Abs(*d)
		if err != nil {
			return nil, err
		}
		*d = dirAbs
		err = util.MkdirIfNotExist(*d)
		if err != nil {
			return nil, err
		}
	}

	pkg := filepath.Base(dir)
	generator := &defaultGenerator{dir: dir, proto: proto, ppack: ppack, cfg: cfg, pkg: pkg}
	var optionList []Option
	optionList = append(optionList, newDefaultOption())
	optionList = append(optionList, opt...)
	for _, fn := range optionList {
		fn(generator)
	}

	return generator, nil
}

// WithConsoleOption creates a console option
func WithConsoleOption(c console.Console) Option {
	return func(generator *defaultGenerator) {
		generator.Console = c
	}
}

func newDefaultOption() Option {
	return func(generator *defaultGenerator) {
		generator.Console = console.NewColorConsole()
	}
}

func (g *defaultGenerator) StartFromDDL(source string, withCache bool) error {
	modelList, err := g.genFromDDL(source, withCache)
	if err != nil {
		return err
	}

	return g.createFile(modelList)
}

func (g *defaultGenerator) StartFromInformationSchema(tables map[string]*model.Table, withCache bool, merge string) error {
	m := make(map[string]string)
	for _, each := range tables {

		if merge != "" {
			each.Table = merge
		}

		table, err := parser.ConvertDataType(each)
		if err != nil {
			return err
		}

		code, err := g.genModel(*table, withCache)
		if err != nil {
			return err
		}

		if table.MultiModelCfg.Table == "" && g.proto != "" && g.ppack != "" {
			err = g.genProto(*table, withCache)
			if err != nil {
				return err
			}
		}

		m[util.ToSource(table.Name, table.MultiModelCfg)] = code

		if merge != "" {
			break
		}
	}

	return g.createFile(m)
}

func (g *defaultGenerator) createFile(modelList map[string]string) error {
	dirAbs, err := filepath.Abs(g.dir)
	//log.Println(g.dir,dirAbs,)
	if err != nil {
		return err
	}

	g.dir = dirAbs
	g.pkg = filepath.Base(dirAbs)
	//log.Println(g.pkg)
	err = util.MkdirIfNotExist(dirAbs)
	if err != nil {
		return err
	}

	for tableName, code := range modelList {
		tn := stringx.From(tableName)
		modelFilename, err := format.FileNamingFormat(g.cfg.NamingFormat, fmt.Sprintf("%s_model", tn.Source()))
		if err != nil {
			return err
		}

		name := modelFilename + ".go"
		filename := filepath.Join(dirAbs, name)
		//if util.FileExists(filename) {
		//	g.Warning("%s already exists, ignored.", name)
		//	continue
		//}
		err = ioutil.WriteFile(filename, []byte(code), os.ModePerm)
		if err != nil {
			return err
		}
	}

	// generate error file
	varFilename, err := format.FileNamingFormat(g.cfg.NamingFormat, "vars")
	if err != nil {
		return err
	}

	filename := filepath.Join(dirAbs, varFilename+".go")
	text, err := util.LoadTemplate(category, errTemplateFile, template.Error)
	if err != nil {
		return err
	}

	err = util.With("vars").Parse(text).SaveTo(map[string]interface{}{
		"pkg": g.pkg,
	}, filename, false)
	if err != nil {
		return err
	}

	g.Success("Done.")
	return nil
}

func (g *defaultGenerator) CreateFactory(model string) error {
	dirAbs, err := filepath.Abs(g.dir)
	//log.Println(g.dir,dirAbs,)
	if err != nil {
		log.Println("Abs err", err)
		return err
	}

	g.dir = dirAbs
	g.pkg = filepath.Base(dirAbs)
	//log.Println(g.pkg)
	err = util.MkdirIfNotExist(dirAbs)
	if err != nil {
		log.Println("MkdirIfNotExist err", err)
		return err
	}

	modelFilename, err := format.FileNamingFormat(g.cfg.NamingFormat, "factory")
	if err != nil {
		log.Println("FileNamingFormat err", err)
		return err
	}

	name := modelFilename + ".go"
	filename := filepath.Join(dirAbs, name)

	err = ioutil.WriteFile(filename, []byte(model), os.ModePerm)
	if err != nil {
		return err
	}

	g.Success("%s Done.", name)
	return nil
}

func (g *defaultGenerator) CreateProto(tableName stringx.String, model string) error {
	dirAbs, err := filepath.Abs(g.proto)
	//log.Println(g.dir,dirAbs,)
	if err != nil {
		log.Println("Abs err", err)
		return err
	}

	g.proto = dirAbs

	//log.Println(g.pkg)
	err = util.MkdirIfNotExist(dirAbs)
	if err != nil {
		log.Println("MkdirIfNotExist err", err)
		return err
	}

	modelFilename, err := format.FileNamingFormat(g.cfg.NamingFormat, tableName.ToSnake())
	if err != nil {
		log.Println("FileNamingFormat err", err)
		return err
	}

	name := modelFilename + ".proto"
	filename := filepath.Join(dirAbs, name)

	err = ioutil.WriteFile(filename, []byte(model), os.ModePerm)
	if err != nil {
		return err
	}

	g.Success("%s Done.", tableName.ToSnake())
	return nil
}

func (g *defaultGenerator) CreateStruct(tableName stringx.String, model string) error {
	dirAbs, err := filepath.Abs(fmt.Sprintf("%s/%s", g.dir, tableName.ToSnake()))
	//log.Println(g.dir,dirAbs,)
	if err != nil {
		log.Println("Abs err", err)
		return err
	}

	//g.proto = dirAbs

	//log.Println(g.pkg)
	err = util.MkdirIfNotExist(dirAbs)
	if err != nil {
		log.Println("MkdirIfNotExist err", err)
		return err
	}

	modelFilename, err := format.FileNamingFormat(g.cfg.NamingFormat, tableName.ToSnake())
	if err != nil {
		log.Println("FileNamingFormat err", err)
		return err
	}

	name := modelFilename + ".go"
	filename := filepath.Join(dirAbs, name)

	err = ioutil.WriteFile(filename, []byte(model), os.ModePerm)
	if err != nil {
		return err
	}

	g.Success("%s Done.", tableName.ToSnake())
	return nil
}

// ret1: key-table name,value-code
func (g *defaultGenerator) genFromDDL(source string, withCache bool) (map[string]string, error) {
	ddlList := g.split(source)
	m := make(map[string]string)
	for _, ddl := range ddlList {
		table, err := parser.Parse(ddl)
		if err != nil {
			return nil, err
		}

		code, err := g.genModel(*table, withCache)
		if err != nil {
			return nil, err
		}

		if table.MultiModelCfg.Table == "" && g.proto != "" && g.ppack != "" {
			err = g.genProto(*table, withCache)
			if err != nil {
				return nil, err
			}
		}

		m[table.Name.Source()] = code
	}

	return m, nil
}

// Table defines mysql table
type Table struct {
	parser.Table
	PrimaryCacheKey        Key
	UniqueCacheKey         []Key
	ContainsUniqueCacheKey bool
}

func (g *defaultGenerator) GenFactory(option string, tables map[string]*model.Table, isMulti bool, merge string) error {
	dirAbs, err := filepath.Abs(g.dir)
	if err != nil {
		return err
	}

	g.dir = dirAbs
	g.pkg = filepath.Base(dirAbs)

	err = g.executeFactory(g.pkg, tables, isMulti, merge)
	if err != nil {
		log.Println("executeFactory err", err)
		return err
	}

	return err
}

func (g *defaultGenerator) genModel(in parser.Table, withCache bool) (string, error) {
	if len(in.PrimaryKey.Name.Source()) == 0 {
		return "", fmt.Errorf("table %s: missing primary key", in.Name.Source())
	}

	primaryKey, uniqueKey := genCacheKeys(in)

	var table Table
	table.Table = in
	table.PrimaryCacheKey = primaryKey
	table.UniqueCacheKey = uniqueKey
	table.ContainsUniqueCacheKey = len(uniqueKey) > 0

	if in.MultiModelCfg.Table != "" {

		findCode := make([]string, 0)
		findMultiCode, findMultiCodeMethod, status, err := genFindMulti(table, withCache)
		if err != nil {
			log.Println("genFindOne err", err)
			return "", err
		}

		importsCode, err := genImports(table, withCache, in.ContainsTime(), status)
		if err != nil {
			log.Println("genImports err", err)
			return "", err
		}

		var list []string
		list = append(list, findMultiCodeMethod)
		typesCode, err := genMultiTypes(table, strings.Join(modelutil.TrimStringSlice(list), util.NL), table.Comment.Source(), withCache)
		if err != nil {
			log.Println("genTypes err", err)
			return "", err
		}

		newCode, err := genNew(table, withCache)
		if err != nil {
			log.Println("genNew err", err)
			return "", err
		}

		ret, err := genFindOneByField(table, withCache)
		if err != nil {
			log.Println("genFindOneByField err", err)
			return "", err
		}
		findCode = append(findCode, findMultiCode, ret.findOneMethod)

		code := &code{
			importsCode: importsCode,
			typesCode:   typesCode,
			newCode:     newCode,
			findCode:    findCode,
		}
		//log.Println(code)
		output, err := g.executeModel(code)
		if err != nil {
			log.Printf("executeModel %s err %s \n", table.Table.Name.ToSnake(), err)
			return "", err
		}

		//err = g.executeStruct(table)
		//if err != nil {
		//	log.Println("executeStruct err", err)
		//	return "", err
		//}
		return output.String(), nil

	} else {

		varsCode, err := genVars(table, withCache)
		if err != nil {
			log.Println("genVars err", err)
			return "", err
		}

		insertCode, insertCodeMethod, err := genInsert(table, withCache)
		if err != nil {
			log.Println("genInsert err", err)
			return "", err
		}

		findCode := make([]string, 0)
		findOneCode, findOneCodeMethod, status, err := genFindOne(table, withCache)
		if err != nil {
			log.Println("genFindOne err", err)
			return "", err
		}

		importsCode, err := genImports(table, withCache, in.ContainsTime(), status)
		if err != nil {
			log.Println("genImports err", err)
			return "", err
		}

		ret, err := genFindOneByField(table, withCache)
		if err != nil {
			log.Println("genFindOneByField err", err)
			return "", err
		}

		findCode = append(findCode, findOneCode, ret.findOneMethod)
		updateCode, updateCodeMethod, err := genUpdate(table, withCache)
		if err != nil {
			log.Println("genUpdate err", err)
			return "", err
		}

		deleteCode, deleteCodeMethod, err := genDelete(table, withCache)
		if err != nil {
			log.Println("genDelete err", err)
			return "", err
		}

		//protoCode,

		var list []string
		list = append(list, insertCodeMethod, findOneCodeMethod, ret.findOneInterfaceMethod, updateCodeMethod, deleteCodeMethod)
		typesCode, err := genTypes(table, strings.Join(modelutil.TrimStringSlice(list), util.NL), table.Comment.Source(), withCache)
		if err != nil {
			log.Println("genTypes err", err)
			return "", err
		}

		newCode, err := genNew(table, withCache)
		if err != nil {
			log.Println("genNew err", err)
			return "", err
		}

		code := &code{
			importsCode: importsCode,
			varsCode:    varsCode,
			typesCode:   typesCode,
			newCode:     newCode,
			insertCode:  insertCode,
			findCode:    findCode,
			updateCode:  updateCode,
			deleteCode:  deleteCode,
			cacheExtra:  ret.cacheExtra,
		}
		//log.Println(code)
		output, err := g.executeModel(code)
		if err != nil {
			log.Printf("executeModel %s err %s \n", table.Table.Name.ToSnake(), err)
			return "", err
		}

		err = g.executeStruct(table)
		if err != nil {
			log.Println("executeStruct err", err)
			return "", err
		}
		return output.String(), nil
	}
	return "", nil
}

func (g *defaultGenerator) genProto(in parser.Table, withCache bool) error {

	var table Table
	table.Table = in

	messageDetail := make([]map[string]interface{}, 0)
	for i, field := range table.Fields {
		messageDetail = append(messageDetail, map[string]interface{}{
			"TypeName": modelutil.DataType2ProtoType(field.DataType),
			"AttrName": field.Name.ToSnake(),
			"Comment":  field.Comment,
			"Num":      fmt.Sprintf("%d", i+1),
			"NotKey":   table.PrimaryKey.Name.Lower() != field.Name.Lower(),
		})
	}

	messageList := make([]map[string]interface{}, 0)
	messageList = append(messageList, map[string]interface{}{
		"Name":          table.Name.ToCamel(),
		"Comment":       table.Comment.Source(),
		"MessageDetail": messageDetail,
	})

	err := g.executeProto(table.Name, g.ppack, messageList)
	if err != nil {
		log.Printf("executeModel %s err %s \n", table.Table.Name.ToSnake(), err)
		return err
	}

	return nil
}

func (g *defaultGenerator) executeProto(tableName stringx.String, pkg string, messageList []map[string]interface{}) error {

	text, err := util.LoadTemplate(category, protoTemplateFile, template.Proto)
	if err != nil {
		log.Println("LoadTemplate err", err)
		return err
	}
	//log.Println(importsCode, typesCode, funcCode)
	t := util.With("proto").
		Parse(text).
		GoFmt(false)
	output, err := t.Execute(map[string]interface{}{
		"TableName":   tableName.ToSnake(),
		"Models":      pkg,
		"MessageList": messageList,
	})
	//log.Println(output)
	if err != nil {
		log.Println("Execute err", err)
		return err
	}
	err = g.CreateProto(tableName, output.String())
	return err
}

func (g *defaultGenerator) executeModel(code *code) (*bytes.Buffer, error) {
	text, err := util.LoadTemplate(category, modelTemplateFile, template.Model)
	if err != nil {
		log.Println("LoadTemplate err", err)
		return nil, err
	}
	t := util.With("model").
		Parse(text).
		GoFmt(true)
	output, err := t.Execute(map[string]interface{}{
		"pkg":     g.pkg,
		"imports": code.importsCode,
		//"vars":        code.varsCode,
		"types":       code.typesCode,
		"new":         code.newCode,
		"insert":      code.insertCode,
		"find":        strings.Join(code.findCode, "\n"),
		"update":      code.updateCode,
		"delete":      code.deleteCode,
		"extraMethod": code.cacheExtra,
	})
	if err != nil {
		return nil, err
	}
	return output, nil
}

func (g *defaultGenerator) executeFactory(pkg string, tables map[string]*model.Table, isMulti bool, merge string) error {

	text, err := util.LoadTemplate(category, factoryFile, template.Factory)
	if err != nil {
		log.Println("LoadTemplate err", err)
		return err
	}
	//log.Println(importsCode, typesCode, funcCode)
	//def_fields, err := genFactoryFields(tables, merge)
	//if err != nil {
	//	return err
	//}

	fields, err := genFactoryFuncFields(tables, merge)
	if err != nil {
		return err
	}

	t := util.With("factory").
		Parse(text).
		GoFmt(true)
	output, err := t.Execute(map[string]interface{}{
		"pkg":                   g.pkg,
		"upperStartCamelObject": stringx.From(pkg).ToCamel(),
		//"def_fields":            def_fields,
		"fields":    fields,
		"withRedis": !isMulti,
		"withMerge": merge != "" && len(fields) > 0,
	})
	//log.Println(output)
	if err != nil {
		log.Println("Execute err", err)
		return err
	}
	err = g.CreateFactory(output.String())
	return err
}

func (g *defaultGenerator) executeStruct(table Table) error {

	text, err := util.LoadTemplate(Factory, factoryFile, template.Struct)
	if err != nil {
		log.Println("LoadTemplate err", err)
		return err
	}
	//log.Println(importsCode, typesCode, funcCode)
	t := util.With("struct").
		Parse(text).
		GoFmt(true)

	fields := make([]map[string]interface{}, 0)
	for _, field := range table.Fields {
		fields = append(fields, map[string]interface{}{
			"field":   field.Name.ToCamel(),
			"name":    field.Name.ToSnake(),
			"comment": field.Comment,
		})
	}

	output, err := t.Execute(map[string]interface{}{
		"pkg":    table.Name.ToSnake(),
		"fields": fields,
	})
	//log.Println(output)
	if err != nil {
		log.Println("Execute err", err)
		return err
	}
	err = g.CreateStruct(table.Name, output.String())
	return err
}

func wrapWithRawString(v string) string {
	if v == "`" {
		return v
	}

	if !strings.HasPrefix(v, "`") {
		v = "`" + v
	}

	if !strings.HasSuffix(v, "`") {
		v = v + "`"
	} else if len(v) == 1 {
		v = v + "`"
	}

	return v
}
