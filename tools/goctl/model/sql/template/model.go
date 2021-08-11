package template

// Model defines a template for model
var Model = `package {{.pkg}}
{{.comment}}
{{.imports}}
{{.vars}}
{{.types}}
{{.new}}
{{.insert}}
{{.find}}
{{.update}}
{{.delete}}
{{.extraMethod}}
`


var Factory = ``

var Proto = ``