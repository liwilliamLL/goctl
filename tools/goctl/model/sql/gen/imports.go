package gen

import (
	"william/base/go-zero/tools/goctl/model/sql/template"
	"william/base/go-zero/tools/goctl/util"
	"strings"
)

func genImports(table Table, withCache, timeImport, status bool) (string, error) {

	sql, importGorm := false, false

	for _, f := range table.Fields {
		if strings.Index(f.DataType, "sql.") != -1 && f.Name.ToCamel() != "DeletedAt" {
			sql = true
		}
		if f.Name.ToCamel() == "DeletedAt" {
			importGorm = true
		}
	}

	if withCache {
		text, err := util.LoadTemplate(category, importsTemplateFile, template.Imports)
		if err != nil {
			return "", err
		}

		buffer, err := util.With("import").Parse(text).Execute(map[string]interface{}{
			"time":      timeImport,
			"status":    status,
			"sql":       sql,
			"gorm":      importGorm,
			"withRedis": table.MultiModelCfg.Table == "",
		})
		if err != nil {
			return "", err
		}

		return buffer.String(), nil
	}

	text, err := util.LoadTemplate(category, importsWithNoCacheTemplateFile, template.ImportsNoCache)
	if err != nil {
		return "", err
	}

	joins := make([]map[string]interface{}, 0)
	if table.MultiModelCfg.Table != "" {
		status, sql, importGorm, timeImport = false, false, false, false
		//joins := []map[string]interface{}{
		//	{"package": table.MultiModelCfg.Package, "aliasPackage": table.MultiModelCfg.AliasPackage},
		//}

		//joins = append(joins, map[string]interface{}{
		//	"package": table.MultiModelCfg.Package, "aliasPackage": table.MultiModelCfg.AliasPackage,
		//})

		packs := map[string]interface{}{
			table.MultiModelCfg.Package: table.MultiModelCfg.AliasPackage,
		}

		for _, join := range table.MultiModelCfg.Joins {
			packs[join.Package] = join.AliasPackage

		}

		for pack, alias := range packs {
			joins = append(joins, map[string]interface{}{
				"package": pack, "aliasPackage": alias,
			})
		}
	}

	buffer, err := util.With("import").Parse(text).Execute(map[string]interface{}{
		"time":      timeImport,
		"status":    status,
		"sql":       sql,
		"gorm":      importGorm,
		"joins":     joins,
		"withRedis": table.MultiModelCfg.Table == "",
	})
	if err != nil {
		return "", err
	}

	return buffer.String(), nil
}

func genFactoryImport(pkg string) (string, error) {
	text, err := util.LoadTemplate(Factory, factoryImportsFile, template.FactoryImport)
	if err != nil {
		return "", err
	}

	buffer, err := util.With("import").Parse(text).Execute(map[string]interface{}{
		"time": false,
		"pkg":  pkg,
	})
	if err != nil {
		return "", err
	}

	return buffer.String(), nil
}
