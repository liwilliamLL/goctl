package model

import (
	"bytes"
	"william/base/go-zero/core/mapping"
	"io/ioutil"
	"os"
	"path/filepath"
)

type (
	MultiModelJoinCfg struct {
		Table        string   `json:"table" yaml:"table"`
		Group        string   `json:"group" yaml:"group"`
		Alias        string   `json:"alias" yaml:"alias"`
		JoinType     string   `json:"joinType" yaml:"joinType"`
		Package      string   `json:"package" yaml:"package"`
		AliasPackage string   `json:"aliasPackage" yaml:"aliasPackage"`
		ForeignKey   string   `json:"foreignKey" yaml:"foreignKey"`
		References   string   `json:"references" yaml:"references"`
		Fields       []string `json:"fields" yaml:"fields"`
		Excludes     []string `json:"excludes" yaml:"excludes"`
	}

	MultiModelCfg struct {
		Table        string              `json:"table" yaml:"table"`
		Group        string              `json:"group" yaml:"group"`
		Package      string              `json:"package" yaml:"package"`
		AliasPackage string              `json:"aliasPackage" yaml:"aliasPackage"`
		Fields       []string            `json:"fields" yaml:"fields"`
		Excludes     []string            `json:"excludes" yaml:"excludes"`
		Joins        []MultiModelJoinCfg `json:"joins" yaml:"joins"`
	}

	Cfg struct {
		Models []MultiModelCfg `json:"models" yaml:"models"`
	}
)

func NewMultiModelCfg(filename string) ([]MultiModelCfg, error) {

	var pattern, fold string
	abs, _ := filepath.Abs(filename)
	fileInfo, err := os.Stat(abs)
	if err == nil {
		if fileInfo.IsDir() {
			pattern = "*.yaml"
			fold = filename
		}else{
			pattern = filepath.Base(filename)
			fold = filepath.Dir(filename)
		}
	}else{
		f := filepath.Dir(filename)
		_, err := os.Stat(f)
		if err != nil {
			return nil, err
		}

		pattern = filepath.Base(filename)
		fold = filepath.Dir(filename)
	}

	//
	//fold := filepath.Dir(filename)
	//_, err := os.Stat(fold)
	//if err != nil {
	//	return nil, err
	//}

	modelCfgs := make([]MultiModelCfg, 0)
	dir, err := ioutil.ReadDir(fold)
	if err != nil {
		return nil, err
	}

	var cfg Cfg
	var content []byte
	//pattern := filepath.Base(filename)
	for _, f := range dir {

		if f.IsDir() {
			continue
		}

		if matched, err := filepath.Match(pattern, f.Name()); !matched || err != nil {
			continue
		}

		content, err = ioutil.ReadFile(filepath.Join(fold, f.Name()))
		if err != nil {
			return nil, err
		}

		err = mapping.UnmarshalYamlReader(bytes.NewReader(content), &cfg)
		if err != nil {
			return nil, err
		}

		modelCfgs = append(modelCfgs, cfg.Models...)
	}
	return modelCfgs, nil
}
