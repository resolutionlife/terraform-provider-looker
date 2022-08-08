package conv

import (
	"errors"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func PString(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func PBool(b bool) *bool {
	return &b
}

func SchemaSetToSliceString(i interface{}) ([]string, error) {
	set, ok := i.(*schema.Set)
	if !ok {
		return nil, errors.New("interface{} is not of type *schema.Set")
	}

	slice := make([]string, set.Len())
	for i, v := range set.List() {
		str, ok := v.(string)
		if !ok {
			return nil, errors.New("set contains a non-string element")
		}
		slice[i] = str
	}

	return slice, nil
}
