package conv

func PString(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func PBool(b bool) *bool {
	return &b
}
