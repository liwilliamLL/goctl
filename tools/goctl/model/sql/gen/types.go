package gen

import (
	"william/base/go-zero/tools/goctl/model/sql/parser"
	"william/base/go-zero/tools/goctl/model/sql/template"
	"william/base/go-zero/tools/goctl/util"
	"william/base/go-zero/tools/goctl/util/stringx"
	"strings"
)

func genTypes(table Table, methods, comment string, withCache bool) (string, error) {
	fields := table.Fields
	fieldsString, err := genFields(fields, &table.PrimaryKey.Field)
	if err != nil {
		return "", err
	}

	text, err := util.LoadTemplate(category, typesTemplateFile, template.Types)
	if err != nil {
		return "", err
	}

	camelName := util.ToCamelTable(table.Name, table.MultiModelCfg, "")

	output, err := util.With("types").
		Parse(text).
		Execute(map[string]interface{}{
			"withCache":             withCache,
			"method":                methods,
			"upperStartCamelObject": camelName,
			"fields":                fieldsString,
			"comment":               comment,
			"withRedis":             true,
		})
	if err != nil {
		return "", err
	}

	return output.String(), nil
}

func genMultiTypes(table Table, methods, comment string, withCache bool) (string, error) {
	//fields := table.Fields

	fields := []*parser.Field{
		&parser.Field{
			Name:            stringx.From(""),
			DataBaseType:    table.Name.ToCamel(),
			DataType:        strings.Join([]string{table.MultiModelCfg.AliasPackage, table.Name.ToCamel()}, "."),
			Comment:         table.Comment.Title(),
			SeqInIndex:      0,
			OrdinalPosition: 0,
		},
	}
	for _, join := range table.MultiModelCfg.Joins {

		camel := stringx.From(join.Table)
		if join.Alias != "" {
			camel = stringx.From(join.Alias)
		}

		fields = append(fields, &parser.Field{
			Name:            camel,
			DataBaseType:    join.Table,
			DataType:        strings.Join([]string{join.AliasPackage, stringx.From(join.Table).ToCamel()}, "."),
			Comment:         "",
			SeqInIndex:      0,
			OrdinalPosition: 0,
			ForeignKey:      join.ForeignKey,
			References:      join.References,
		})
	}

	fieldsString, err := genFields(fields, &table.PrimaryKey.Field)
	if err != nil {
		return "", err
	}

	text, err := util.LoadTemplate(category, typesTemplateFile, template.Types)
	if err != nil {
		return "", err
	}

	camelName := util.ToCamelTable(table.Name, table.MultiModelCfg, "")

	output, err := util.With("types").
		Parse(text).
		Execute(map[string]interface{}{
			"withCache":             withCache,
			"method":                methods,
			"upperStartCamelObject": camelName,
			"fields":                fieldsString,
			"comment":               comment,
			"withRedis":             false,
		})
	if err != nil {
		return "", err
	}

	return output.String(), nil
}
//
//func genFactoryTypes(pkg string, tables map[string]*model.Table) (string, error) {
//	//for _,k:=range table {
//	//	tables, err := parser.ConvertDataType(k)
//	//	if err != nil {
//	//		return "", err
//	//	}
//	//}
//
//	fieldsString, err := genFactoryFields(tables)
//	if err != nil {
//		return "", err
//	}
//
//	text, err := util.LoadTemplate(Factory, factoryTypesFile, template.FactoryTypes)
//	if err != nil {
//		return "", err
//	}
//
//	output, err := util.With("types").
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

func UpdateUpper(a string) string {
	vv := []rune(a)
	if len(vv) != 0 && vv[0] >= 97 && vv[0] <= 132 {
		vv[0] = vv[0] - 32
	}
	return string(vv)
}
