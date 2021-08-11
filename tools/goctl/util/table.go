package util

import (
	"fmt"
	"william/base/go-zero/tools/goctl/model/sql/model"
	"william/base/go-zero/tools/goctl/util/stringx"
	"strings"
)

func ToCamelTable(n stringx.String, cfg model.MultiModelCfg, merge string) string {

	if merge != "" {
		return stringx.From(merge).ToCamel()
	}

	name := strings.Builder{}
	name.WriteString(n.ToCamel())
	if &cfg != nil {
		for _, t := range cfg.Joins {
			name.WriteString(fmt.Sprintf("__%s", stringx.From(t.Table).ToCamel()))
		}
	}

	return name.String()
}

func ToSource(n stringx.String, cfg model.MultiModelCfg) string {

	name := strings.Builder{}
	name.WriteString(n.Source())
	if &cfg != nil {
		for _, t := range cfg.Joins {
			name.WriteString(fmt.Sprintf("__%s", stringx.From(t.Table).Source()))
		}
	}

	return name.String()
}
