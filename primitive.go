package configByTag

import "strconv"

func IsStrInt(s string) bool {
	tmp, err := strconv.Atoi(s)
	if err != nil {
		return false
	}
	return strconv.Itoa(tmp) == s
}

func StrIntAsSecond(s string) string {
	if IsStrInt(s) {
		s += "s"
	}
	return s
}
