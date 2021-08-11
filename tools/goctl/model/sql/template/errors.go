package template

// Error defines an error template
var Error = `package {{.pkg}}

import "william/base/go-zero/core/stores/sqlx"

var ErrNotFound = sqlx.ErrNotFound
`
