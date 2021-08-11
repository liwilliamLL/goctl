package gen

import (
	"fmt"
	"william/base/go-zero/tools/goctl/model/sql/template"
	"william/base/go-zero/tools/goctl/util"
)

func genTag(in, foreignKey, references string, isPrimaryKey bool) (string, error) {
	if in == "" {
		return in, nil
	}

	text, err := util.LoadTemplate(category, tagTemplateFile, template.Tag)
	if err != nil {
		return "", err
	}

	if foreignKey != "" {
		in = ""
		foreignKey = fmt.Sprintf(";foreignKey:%s", foreignKey)
	}

	if references != "" {
		references = fmt.Sprintf(";references:%s", references)
	}

	output, err := util.With("tag").Parse(text).Execute(map[string]interface{}{
		"field":        in,
		"isPrimaryKey": isPrimaryKey,
		"foreignKey":   foreignKey,
		"references":   references,
	})
	if err != nil {
		return "", err
	}

	return output.String(), nil
}
