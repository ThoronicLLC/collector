package core

import (
	"fmt"
	"github.com/gookit/validate"
	"reflect"
	"strings"
)

func ValidateStruct(requestObject interface{}) error {
	if reflect.ValueOf(requestObject).Kind() != reflect.Ptr {
		return fmt.Errorf("request object must be a pointer")
	}

	// Validate structure
	v := validate.Struct(requestObject)
	if v.Validate() {
		err := v.BindSafeData(requestObject)
		if err != nil {
			return fmt.Errorf("invalid structure")
		}
		return nil
	} else {
		if !v.Errors.Empty() {
			mappedErrors := mapContractErrorsToArray(v.Errors.All())
			return fmt.Errorf("validation failed: %s", strings.Join(mappedErrors, "; "))
		} else {
			return fmt.Errorf("validation error")
		}
	}
}

func mapContractErrorsToArray(err map[string]map[string]string) []string {
	arr := make([]string, 0)
	for _, v := range err {
		for _, v2 := range v {
			arr = append(arr, v2)
		}
	}
	return arr
}
