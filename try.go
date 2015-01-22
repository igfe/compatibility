package compatibility

import (
	"fmt"
	"github.com/gogo/protobuf/parser"
	descriptor "github.com/gogo/protobuf/protoc-gen-gogo/descriptor"
	"strings"
)

func check(err error) {
	if err != nil {
		panic(err)
	}
}

func getExtendee(path []string, d []*descriptor.DescriptorProto) *descriptor.DescriptorProto {
	for _, val := range d {
		c := 0
		for ; path[c] == ""; c++ {

		}
		fmt.Println(*val.Name)
		fmt.Println(*val.Name + "=?" + path[c])
		if *val.Name == path[c] {
			fmt.Println("match " + path[c])
			if len(path) == 1 {
				return val
			} else {
				return getExtendee(path[c+1:], val.NestedType)
			}
		}
	}
	fmt.Println("fail")
	return nil
}

func getChangesFieldExtension(d *descriptor.DescriptorProto, f *FieldDescriptorProto) {
	var output DifferenceList
	for _, val1 := range newer { //loop through both arrays to see which fields existed in the older version too and which were newly added
			if *val1.Number == *f.Number { //if message exists in both, check label, numeric tag and type for dissimilarities
				exist = true
				if val1.Label.String() != f.Label.String() { //If field label changed add it to differences
					if *val1.Label == descriptor.FieldDescriptorProto_LABEL_REQUIRED {
						output.addError(ChangedLabel, val1.Label.String(), f.Label.String(), prefix, strconv.Itoa(int(*val1.Number)), "")
					} else if *f.Label == descriptor.FieldDescriptorProto_LABEL_REQUIRED {
						output.addError(ChangedLabel, val1.Label.String(), f.Label.String(), prefix, strconv.Itoa(int(*val1.Number)), "")
					} else {
						output.addWarning(ChangedLabel, val1.Label.String(), f.Label.String(), prefix, strconv.Itoa(int(*val1.Number)), "")
					}
				}
				if *val1.Name != *f.Name {
					output.addWarning(ChangedName, *val1.Name, *f.Name, prefix, strconv.Itoa(int(*val1.Number)), "")
				}
				if *val1.Type != *f.Type {
					compatible := false
					if *val1.Type == descriptor.FieldDescriptorProto_TYPE_INT32 || *val1.Type == descriptor.FieldDescriptorProto_TYPE_INT64 || *val1.Type == descriptor.FieldDescriptorProto_TYPE_UINT32 || *val1.Type == descriptor.FieldDescriptorProto_TYPE_UINT64 || *val1.Type == descriptor.FieldDescriptorProto_TYPE_BOOL {
						if *f.Type == descriptor.FieldDescriptorProto_TYPE_INT32 || *f.Type == descriptor.FieldDescriptorProto_TYPE_INT64 || *f.Type == descriptor.FieldDescriptorProto_TYPE_UINT32 || *f.Type == descriptor.FieldDescriptorProto_TYPE_UINT64 || *f.Type == descriptor.FieldDescriptorProto_TYPE_BOOL {
							compatible = true
						}
					}
					if *val1.Type == descriptor.FieldDescriptorProto_TYPE_SINT32 || *val1.Type == descriptor.FieldDescriptorProto_TYPE_SINT64 {
						if *f.Type == descriptor.FieldDescriptorProto_TYPE_SINT32 || *f.Type == descriptor.FieldDescriptorProto_TYPE_SINT64 {
							compatible = true
						}
					}
					if *val1.Type == descriptor.FieldDescriptorProto_TYPE_STRING || *val1.Type == descriptor.FieldDescriptorProto_TYPE_BYTES {
						if *f.Type == descriptor.FieldDescriptorProto_TYPE_STRING || *f.Type == descriptor.FieldDescriptorProto_TYPE_BYTES {
							compatible = true
						}
					}
					if *val1.Type == descriptor.FieldDescriptorProto_TYPE_FIXED32 || *val1.Type == descriptor.FieldDescriptorProto_TYPE_FIXED64 || *val1.Type == descriptor.FieldDescriptorProto_TYPE_SFIXED32 || *val1.Type == descriptor.FieldDescriptorProto_TYPE_SFIXED64 {
						if *f.Type == descriptor.FieldDescriptorProto_TYPE_FIXED32 || *f.Type == descriptor.FieldDescriptorProto_TYPE_FIXED64 || *f.Type == descriptor.FieldDescriptorProto_TYPE_SFIXED32 || *f.Type == descriptor.FieldDescriptorProto_TYPE_SFIXED64 {
							compatible = true
						}
					}
					if compatible {
						output.addWarning(ChangedType, val1.Type.String(), f.Type.String(), prefix, strconv.Itoa(int(*val1.Number)), "")
					} else {
						output.addError(ChangedType, val1.Type.String(), f.Type.String(), prefix, strconv.Itoa(int(*val1.Number)), "")
					}
				}
				if val1.DefaultValue != f.DefaultValue {
					output.addWarning(ChangedDefault, val1.GetDefaultValue(), f.GetDefaultValue(), prefix, strconv.Itoa(int(*val1.Number)), "")
				}
			}
		}
		if !exist {
			fmt.Println(*val1.Number)
			fmt.Println(oldEx)
			if *val1.Label == descriptor.FieldDescriptorProto_LABEL_REQUIRED {
				if isExtension(int(*val1.Number), oldEx) {
					output.addError(AddedField, val1.Label.String(), "", prefix, strconv.Itoa(int(*val1.Number)), " this number was previously used by extensions")
				} else {
					output.addError(AddedField, val1.Label.String(), "", prefix, strconv.Itoa(int(*val1.Number)), "")
				}
			}
		}
	}
	for _, val1 := range older {
		exist := false
		for _, f := range newer {
			if *val1.Number == *f.Number {
				exist = true
			}
		}
		if !exist {
			if *val1.Label == descriptor.FieldDescriptorProto_LABEL_REQUIRED {
				if isExtension(int(*val1.Number), newEx) {
					output.addError(RemovedField, val1.Label.String(), "", prefix, strconv.Itoa(int(*val1.Number)), " this number is now assigned to extensions")
				} else {
					output.addError(RemovedField, val1.Label.String(), "", prefix, strconv.Itoa(int(*val1.Number)), "")
				}
			} else if isExtension(int(*val1.Number), newEx) {
				output.addWarning(RemovedField, val1.Label.String(), "", prefix, strconv.Itoa(int(*val1.Number)), " this number is now usable by extensions")
			} else {
				output.addWarning(RemovedField, val1.Label.String(), "", prefix, strconv.Itoa(int(*val1.Number)), " consider prefixing \"OBSOLETE_\" instead")
			}
		}
	}
	for _, val1 := range newer {
		for _, f := range older {
			if *val1.Name == *f.Name {
				if *val1.Number != *f.Number {
					output.addWarning(ChangedNumber, strconv.Itoa(int(*val1.Number)), strconv.Itoa(int(*f.Number)), prefix, *val1.Name, "")
				}
			}
		}
	}
	return output
}

func main() {
	// d1, err1 := parser.ParseFile("./p.proto", ".")
	// check(err1)
	d2, err2 := parser.ParseFile("./ExtensionProtos/q.proto", "./ExtensionProtos/")
	check(err2)

	fmt.Println(getExtendee(strings.Split(".m.n", "."), d2.File[0].MessageType))
}
