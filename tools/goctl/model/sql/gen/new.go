package gen

import (
	"william/base/go-zero/tools/goctl/model/sql/template"
	"william/base/go-zero/tools/goctl/util"
)

func genNew(table Table, withCache bool) (string, error) {
	text, err := util.LoadTemplate(category, modelNewTemplateFile, template.New)
	if err != nil {
		return "", err
	}

	camelName := util.ToCamelTable(table.Name, table.MultiModelCfg, "")

	output, err := util.With("new").
		Parse(text).
		Execute(map[string]interface{}{
			"table":                 wrapWithRawString(table.Name.Source()),
			"withCache":             withCache,
			"upperStartCamelObject": camelName,
			"originTable":           table.Name.Source(),
			"withRedis":             table.Table.MultiModelCfg.Table == "",
		})
	if err != nil {
		return "", err
	}

	return output.String(), nil
}
