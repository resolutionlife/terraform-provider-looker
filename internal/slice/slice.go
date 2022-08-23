package slice

func Diff(s, t []string) []string {
	var diff = []string{}

	for _, sval := range s {
		if !contains(t, sval) {
			diff = append(diff, sval)
		}
	}
	for _, tval := range t {
		if !contains(s, tval) {
			diff = append(diff, tval)
		}
	}

	return diff
}

func Delete(s []string, toDelete []string) (str []string) {
	for i := range s {
		if !contains(toDelete, s[i]) {
			str = append(str, s[i])
		}
	}
	return
}

func contains(s []string, v string) bool {
	for i := range s {
		if s[i] == v {
			return true
		}
	}
	return false
}
