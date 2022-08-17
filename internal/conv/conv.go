package conv

import "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

func PString(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func PSlices(s *schema.Set) *[]string {
	if s == nil {
		return nil
	}

	slice := make([]string, 0, s.Len())
	for _, v := range s.List() {
		slice = append(slice, v.(string))
	}

	return &slice
}

func PBool(b bool) *bool {
	return &b
}
