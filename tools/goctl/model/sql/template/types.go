package template

// Types defines a template for types in model
var Types = `
type (
	{{.upperStartCamelObject}}Model interface{
		{{.method}}
	}

	default{{.upperStartCamelObject}}Model struct {
		{{if .withCache}}sqlc.CachedConn{{else}}conn gorm.DB{{end}}
		table string
	}

	{{.upperStartCamelObject}} struct {
		{{.fields}}
	}
)
`


var FactoryTypes = `
type (
	{{.upperStartCamelObject}} struct {
		{{.fields}}
	}
)
`


var FactoryFunc = `
func New{{.upperStartCamelObject}}(config *mysql.Config) *{{.upperStartCamelObject}} {
	dataSource,err := mysql.NewDataSource(config)
	if err !=nil{
		panic(err)
	}
	return &{{.upperStartCamelObject}}{
		{{.fields}}
	}
}
`
