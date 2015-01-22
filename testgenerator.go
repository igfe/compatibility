package compatibility

import (
	"fmt"
	"strconv"
)

var types []string = []string{"double", "float", "int32", "int64", "uint32", "uint64", "sint32", "sint64", "fixed32", "fixed64", "sfixed32", "sfixed64", "bool", "string", "bytes"}
var cInt []string = []string{"int32", "int64", "uint32", "uint64", "bool"}
var cSInt []string = []string{"sint32", "sint64"}
var cSB []string = []string{"string", "bytes"}
var cFixed []string = []string{"fixed32", "fixed64", "sfixed32", "sfixed64"}

func print(s1, s2 string) bool {
	if s1 == s2 {
		return false
	} else if compatible(s1, s2, cInt) || compatible(s1, s2, cSInt) || compatible(s1, s2, cSB) || compatible(s1, s2, cFixed) {
		return false
	} else {
		return true
	}
}

func compatible(s1, s2 string, sS []string) bool {
	foundS1 := false
	foundS2 := false
	for _, val := range sS {
		if s1 == val {
			foundS1 = true
		}
		if s2 == val {
			foundS2 = true
		}
		if foundS1 && foundS2 {
			return true
		}
	}
	return false
}

func main() {
	string1 := ""
	string2 := ""
	c := 1
	for i := 0; i < len(types); i++ {
		for j := i + 1; j < len(types); j++ {
			if print(types[i], types[j]) {
				string1 = string1 + "  required " + types[i] + " field" + strconv.Itoa(c) + " = " + strconv.Itoa(c) + ";\n"
				string2 = string2 + "  required " + types[j] + " field" + strconv.Itoa(c) + " = " + strconv.Itoa(c) + ";\n"
				c++
			}
		}
	}
	fmt.Println(string1)
	fmt.Println(string2)
}
