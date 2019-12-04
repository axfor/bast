package validate

// import (
// 	"reflect"
// )

// type Validator struct {
// 	v int
// }

// //Struct v
// //data is validate data
// //rules is validate rule such as:
// // key=>required|int|min:1
// // id_orderkey=>required|int|min:1
// // start_timekey=>sometimes|required|date
// func (v *Validator) Struct(data interface{}, rules ...string) {
// 	t := reflect.TypeOf(data)
// 	switch t.Kind() {
// 	case reflect.Struct:
// 	case reflect.Ptr:
// 		t = t.Elem()
// 		// objV = objV.Elem()
// 	default:
// 		return
// 	}

// }
