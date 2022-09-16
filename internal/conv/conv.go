package conv

import (
	"errors"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func PString(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func PSlices[T any](s []T) *[]T {
	if len(s) == 0 {
		return nil
	}

	return &s
}

func PBool(b bool) *bool {
	return &b
}

func P[T any](v T) *T {
	return &v
}

func SchemaSetToSliceString(set *schema.Set) ([]string, error) {
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
