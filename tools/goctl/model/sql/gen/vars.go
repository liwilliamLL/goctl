package gen

import (
	"strings"

	"william/base/go-zero/tools/goctl/model/sql/template"
	"william/base/go-zero/tools/goctl/util"
	"william/base/go-zero/tools/goctl/util/stringx"
)

func genVars(table Table, withCache bool) (string, error) {
	keys := make([]string, 0)
	keys = append(keys, table.PrimaryCacheKey.VarExpression)
	for _, v := range table.UniqueCacheKey {
		keys = append(keys, v.VarExpression)
	}

	camelName := table.Name.ToCamel()
	text, err := util.LoadTemplate(category, varTemplateFile, template.Vars)
	if err != nil {
		return "", err
	}

	output, err := util.With("var").Parse(text).
		GoFmt(true).Execute(map[string]interface{}{
		"lowerStartCamelObject": stringx.From(camelName).Untitle(),
		"upperStartCamelObject": camelName,
		"cacheKeys":             strings.Join(keys, "\n"),
		"autoIncrement":         table.PrimaryKey.AutoIncrement,
		"originalPrimaryKey":    wrapWithRawString(table.PrimaryKey.Name.Source()),
		"withCache":             withCache,
	})
	if err != nil {
		return "", err
	}

	return output.String(), nil
}
