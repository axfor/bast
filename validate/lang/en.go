package lang

//en
var en = map[string]string{
	"required":   "The {0} field is required",
	"date":       "The {0} is not a valid date",
	"int":        "The {0} must be an integer",
	"max.string": "The {0} must be less than {1} characters",
	"max.int":    "The {0} must be less than {1}",
	"min.string": "The {0} must must be greater than {1} characters",
	"min.int":    "The {0} must must be greater than {1}",
}

func init() {
	Register("en", en)
}
