//Copyright 2018 The axx Authors. All rights reserved.

package guid

import "github.com/google/uuid"

//GUID create GUID
func GUID() string {
	return uuid.New().String()
}
