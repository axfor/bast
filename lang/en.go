//Copyright 2018 The axx Authors. All rights reserved.

package lang

//en
var en = map[string]string{
	"v.required":   "The {0} field is required",
	"v.date":       "The {0} is not a valid date",
	"v.int":        "The {0} must be an integer",
	"v.max.string": "The {0} must be less than {1} characters",
	"v.max.int":    "The {0} must be less than {1}",
	"v.min.string": "The {0} must must be greater than {1} characters",
	"v.min.int":    "The {0} must must be greater than {1}",
	"v.email":      "The {0} must be a valid email address",
	"v.ip":         "The {0} must be a valid ip address",
	"v.match":      "The {0} is a invalid format",
}

func init() {
	Register("en", en)
}
