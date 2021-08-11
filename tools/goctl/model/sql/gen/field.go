package gen

import (
	"william/base/go-zero/tools/goctl/model/sql/model"
	"strings"

	"william/base/go-zero/tools/goctl/model/sql/parser"
	"william/base/go-zero/tools/goctl/model/sql/template"
	"william/base/go-zero/tools/goctl/util"
)

func genFields(fields []*parser.Field, primaryKey *parser.Field) (string, error) {
	var list []string

	for _, field := range fields {
		result, err := genField(field, primaryKey)
		if err != nil {
			return "", err
		}

		list = append(list, result)
	}

	return strings.Join(list, "\n"), nil
}

func genField(field *parser.Field, primaryKey *parser.Field) (string, error) {
	tag, err := genTag(field.Name.Source(), field.ForeignKey, field.References, primaryKey.Name == field.Name)
	name := field.Name.ToCamel()
	typeName := field.DataType
	if name == "DeletedAt" {
		typeName = "gorm.DeletedAt"
	}
	if err != nil {
		return "", err
	}

	text, err := util.LoadTemplate(category, fieldTemplateFile, template.Field)
	if err != nil {
		return "", err
	}

	output, err := util.With("types").
		Parse(text).
		Execute(map[string]interface{}{
			"name":       field.Name.ToCamel(),
			"type":       typeName,
			"tag":        tag,
			"hasComment": field.Comment != "",
			"comment":    field.Comment,
		})
	if err != nil {
		return "", err
	}

	return output.String(), nil
}

func genFactoryFuncFields(tables map[string]*model.Table, merge string) ([]map[string]interface{}, error) {
	var list []map[string]interface{}

	for _, table := range tables {
		t, err := parser.ConvertDataType(table)
		if err != nil {
			return nil, err
		}
		result := map[string]interface{}{
			"name":       util.ToCamelTable(t.Name, t.MultiModelCfg, ""),
			"structName": util.ToCamelTable(t.Name, t.MultiModelCfg, merge),
			"tableName":  t.Name.Source(),
			"withRedis":  table.MultiModelCfg.Table == "",
			"comment":    t.Comment.Source(),
		}

		list = append(list, result)
	}

	return list, nil
}
