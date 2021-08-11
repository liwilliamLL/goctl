package gen

import (
	"fmt"
	"william/base/go-zero/tools/goctl/model/sql/template"
	"william/base/go-zero/tools/goctl/util"
	"william/base/go-zero/tools/goctl/util/stringx"
	"strings"
)

func genFindMulti(table Table, withCache bool) (string, string, bool, error) {
	var status bool
	camelName := util.ToCamelTable(table.Name, table.MultiModelCfg, "")
	text, err := util.LoadTemplate(category, findMultiTemplateFile, template.FindOne)
	if err != nil {
		return "", "", status, err
	}

	if table.PrimaryKey.DataType == "int64" || table.PrimaryKey.DataType == "int32" {
		status = true
	}

	joins := make([]map[string]interface{}, 0)
	for _, join := range table.MultiModelCfg.Joins {

		camel := stringx.From(join.Table).ToCamel()
		if join.Alias != "" {
			camel = stringx.From(join.Alias).ToCamel()
		}

		joins = append(joins, map[string]interface{}{
			"upperStartCamelObject": camel,
			"foreignKey":            join.ForeignKey,
			"references":            join.References,
			"joinType":              strings.ToUpper(join.JoinType),
			"lowerStartCamelObject": stringx.From(join.Table).Lower(),
			"snakeStartCamelObject": stringx.From(join.Table).ToSnake(),
			"dataType":              strings.Join([]string{join.AliasPackage, stringx.From(join.Table).ToCamel()}, "."),
		})
	}

	group, distinct := "", ""
	if table.MultiModelCfg.Group != "" {
		group = strings.Join([]string{stringx.From(table.MultiModelCfg.Table).ToSnake(), table.MultiModelCfg.Group}, ".")
		group = fmt.Sprintf(`Group("%s").`, group)

		distinct = strings.Join([]string{stringx.From(table.MultiModelCfg.Table).ToSnake(), table.MultiModelCfg.Group}, ".")
		distinct = fmt.Sprintf(`Distinct("%s").`, distinct)
	} else if len(table.MultiModelCfg.Joins) > 0 {

		for _, join := range table.MultiModelCfg.Joins {
			if join.Group != "" {

				camel := stringx.From(join.Table).ToCamel()
				if join.Alias != "" {
					camel = stringx.From(join.Alias).ToCamel()
				}

				group = strings.Join([]string{camel, table.MultiModelCfg.Group}, ".")
				group = fmt.Sprintf(`Group("%s").`, group)

				distinct = strings.Join([]string{camel, table.MultiModelCfg.Group}, ".")
				distinct = fmt.Sprintf(`Distinct("%s").`, distinct)
				break
			}
		}
	}

	output, err := util.With("findOne").
		Parse(text).
		Execute(map[string]interface{}{
			"withCache":                 withCache,
			"upperStartCamelObject":     camelName,
			"lowerStartCamelObject":     stringx.From(camelName).Untitle(),
			"snakeStartCamelObject":     table.Name.ToSnake(),
			"originalPrimaryKey":        wrapWithRawString(table.PrimaryKey.Name.Source()),
			"lowerStartCamelPrimaryKey": stringx.From(table.PrimaryKey.Name.ToCamel()).Untitle(),
			"uperStartCamelPrimaryKey":  table.PrimaryKey.Name.ToCamel(),
			"dataType":                  table.PrimaryKey.DataType,
			"cacheKey":                  table.PrimaryCacheKey.KeyExpression,
			"cacheKeyVariable":          table.PrimaryCacheKey.KeyLeft,
			"status":                    status,
			"joins":                     joins,
			"group":                     group,
			"distinct":                  distinct,
		})
	if err != nil {
		return "", "", status, err
	}

	text, err = util.LoadTemplate(category, findOneMethodTemplateFile, template.FindOneMethod)
	if err != nil {
		return "", "", status, err
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
		return "", "", status, err
	}

	return output.String(), findOneMethod.String(), status, nil
}
