package gen

import (
	"william/base/go-zero/tools/goctl/model/sql/template"
	"william/base/go-zero/tools/goctl/util"
	"william/base/go-zero/tools/goctl/util/stringx"
)

func genFindOne(table Table, withCache bool) (string, string,bool, error) {
	var status bool
	camelName := table.Name.ToCamel()
	text, err := util.LoadTemplate(category, findOneTemplateFile, template.FindOne)
	if err != nil {
		return "", "",status, err
	}

	if table.PrimaryKey.DataType =="int64"||table.PrimaryKey.DataType == "int32" {
		status=true
	}

	output, err := util.With("findOne").
		Parse(text).
		Execute(map[string]interface{}{
			"withCache":                 withCache,
			"upperStartCamelObject":     camelName,
			"lowerStartCamelObject":     stringx.From(camelName).Untitle(),
			"originalPrimaryKey":        wrapWithRawString(table.PrimaryKey.Name.Source()),
			"lowerStartCamelPrimaryKey": stringx.From(table.PrimaryKey.Name.ToCamel()).Untitle(),
			"uperStartCamelPrimaryKey":  table.PrimaryKey.Name.ToCamel(),
			"dataType":                  table.PrimaryKey.DataType,
			"cacheKey":                  table.PrimaryCacheKey.KeyExpression,
			"cacheKeyVariable":          table.PrimaryCacheKey.KeyLeft,
			"status"  :                  status,
		})
	if err != nil {
		return "", "",status, err
	}

	text, err = util.LoadTemplate(category, findOneMethodTemplateFile, template.FindOneMethod)
	if err != nil {
		return "", "",status, err
	}

	findOneMethod, err := util.With("findOneMethod").
		Parse(text).
		Execute(map[string]interface{}{
			"upperStartCamelObject":     camelName,
			"lowerStartCamelPrimaryKey": stringx.From(table.PrimaryKey.Name.ToCamel()).Untitle(),
			"uperStartCamelPrimaryKey":  table.PrimaryKey.Name.ToCamel(),
			"dataType":                  table.PrimaryKey.DataType,
		})
	if err != nil {
		return "", "",status, err
	}

	return output.String(), findOneMethod.String(),status,nil
}
