package slice

func Diff(s, t []string) (diff []string) {
	for _, sval := range s {
		if !Contains(t, sval) {
			diff = append(diff, sval)
		}
	}

	for _, tval := range t {
		if !Contains(s, tval) {
			diff = append(diff, tval)
		}
	}
	return
}

func Delete(s []string, toDelete []string) (str []string) {
	for i := range s {
		if !Contains(toDelete, s[i]) {
			str = append(str, s[i])
		}
	}
	return
}

func Contains(s []string, v string) bool {
	for i := range s {
		if s[i] == v {
			return true
		}
	}
	return false
}
