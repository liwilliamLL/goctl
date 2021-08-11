package util

var typeArr = map[string]string{
	"sql.NullString": "string",
	"sql.NullTime": "string",
	"time.Time": "string",
	"sql.NullInt64": "int64",
	"sql.NullFloat64": "double",
	"float64": "double",
}

func DataType2ProtoType(dataType string) string {

	t, ok := typeArr[dataType]
	if !ok {
		return dataType
	}
	return t
}