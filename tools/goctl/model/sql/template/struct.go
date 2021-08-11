package template

// Struct defines an struct template
var Struct = `package {{.pkg}}
{{range .fields -}}
// {{.field}} {{.comment}}
const {{.field}} = "{{.name}}"
{{ end }}
`
