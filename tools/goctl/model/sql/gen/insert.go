package gen

import (
	"strings"

	"william/base/go-zero/core/collection"
	"william/base/go-zero/tools/goctl/model/sql/template"
	"william/base/go-zero/tools/goctl/util"
	"william/base/go-zero/tools/goctl/util/stringx"
)

func genInsert(table Table, withCache bool) (string, string, error) {
	keySet := collection.NewSet()
	keyVariableSet := collection.NewSet()
	for _, key := range table.UniqueCacheKey {
		keySet.AddStr(key.DataKeyExpression)
		keyVariableSet.AddStr(key.KeyLeft)
	}

	expressions := make([]string, 0)
	expressionValues := make([]string, 0)
	for _, field := range table.Fields {
		camelName := field.Name.ToCamel()
		if camelName == "CreateTime" || camelName == "UpdateTime" {
			continue
		}

		if field.Name.Source() == table.PrimaryKey.Name.Source() {
			if table.PrimaryKey.AutoIncrement {
				continue
			}
		}

		expressions = append(expressions, "?")
		expressionValues = append(expressionValues, "data."+camelName)
	}

	camelName := table.Name.ToCamel()
	text, err := util.LoadTemplate(category, insertTemplateFile, template.Insert)
	if err != nil {
		return "", "", err
	}

	output, err := util.With("insert").
		Parse(text).
		Execute(map[string]interface{}{
			"withCache":             withCache,
			"containsIndexCache":    table.ContainsUniqueCacheKey,
			"upperStartCamelObject": camelName,
			"lowerStartCamelObject": stringx.From(camelName).Untitle(),
			"expression":            strings.Join(expressions, ", "),
			"expressionValues":      strings.Join(expressionValues, ", "),
			"keys":                  strings.Join(keySet.KeysStr(), "\n"),
			"keyValues":             strings.Join(keyVariableSet.KeysStr(), ", "),
		})
	if err != nil {
		return "", "", err
	}

	// interface method
	text, err = util.LoadTemplate(category, insertTemplateMethodFile, template.InsertMethod)
	if err != nil {
		return "", "", err
	}

	insertMethodOutput, err := util.With("insertMethod").Parse(text).Execute(map[string]interface{}{
		"upperStartCamelObject": camelName,
	})
	if err != nil {
		return "", "", err
	}

	return output.String(), insertMethodOutput.String(), nil
}
//
//func genFactoryFunc(pkg string, table map[string]*model.Table) (string, error) {
//	fieldsString, err := genFactoryFuncFields(table)
//	if err != nil {
//		return "", err
//	}
//
//	text, err := util.LoadTemplate(Factory, factoryFuncFile, template.FactoryFunc)
//	if err != nil {
//		return "", err
//	}
//
//	output, err := util.With("insert").
//		Parse(text).
//		Execute(map[string]interface{}{
//			//"upperStartCamelObject": table.Name.ToCamelTable(),
//			//"upperStartModelObject": table.Name.ToCamelTable(),
//			"upperStartCamelObject": fmt.Sprintf("%s%s", UpdateUpper(pkg), "Dao"),
//			"fields":                fieldsString,
//		})
//	if err != nil {
//		return "", err
//	}
//
//	return output.String(), nil
//}
