package compatibility

import (
	"code.google.com/p/gogoprotobuf/parser"
	descriptor "code.google.com/p/gogoprotobuf/protoc-gen-gogo/descriptor"
	"fmt"
	"os"
	"strconv"
	"strings"
)

type Condition int

const (
	ChangedLabel            Condition = 1
	AddedField              Condition = 2
	RemovedField            Condition = 3
	ChangedName             Condition = 4
	ChangedType             Condition = 5
	ChangedNumber           Condition = 6
	ChangedDefault          Condition = 7
	NonFieldIncompatibility Condition = 8
)

type Difference struct {
	condition Condition
	newValue  string
	oldValue  string
	path      string
	qualifier string
	message   string
}

func (d *Difference) String() string {
	if d.condition == ChangedLabel {
		return "Changed label of field nr " + d.qualifier + " in " + d.path + " from " + d.oldValue + " to " + d.newValue
	} else if d.condition == AddedField {
		return "Added Field nr " + d.qualifier + " in " + d.path + " of label " + d.newValue
	} else if d.condition == RemovedField {
		return "Removed Field nr " + d.qualifier + " in " + d.path + " of label " + d.newValue + d.message
	} else if d.condition == ChangedName {
		return "Changed name of field nr " + d.qualifier + " in " + d.path + " from " + d.oldValue + " to " + d.newValue
	} else if d.condition == ChangedType {
		return "Changed type of field nr " + d.qualifier + " in " + d.path + " from " + d.oldValue + " to " + d.newValue
	} else if d.condition == ChangedNumber {
		return "Changed numeric tag of field named \"" + d.qualifier + "\" in " + d.path + " from " + d.oldValue + " to " + d.newValue
	} else if d.condition == ChangedDefault {
		return "Changed default value of field nr " + d.qualifier + " in " + d.path + " from " + d.oldValue + " to " + d.newValue + " this is generally OK"
	} else if d.condition == NonFieldIncompatibility {
		return d.message
	}
	return ""
}

type DifferenceList struct {
	Error   []Difference
	Warning []Difference
}

func (d *DifferenceList) addWarning(c Condition, newValue, oldValue, path, qualifier, message string) {
	d1 := Difference{c, newValue, oldValue, path, qualifier, message}
	d.Warning = append(d.Warning, d1)
}

func (d *DifferenceList) addError(c Condition, newValue, oldValue, path, qualifier, message string) {
	d1 := Difference{c, newValue, oldValue, path, qualifier, message}
	d.Error = append(d.Error, d1)
}

func (d1 *DifferenceList) merge(d2 DifferenceList) {
	d1.Error = append(d1.Error, d2.Error...)
	d1.Warning = append(d1.Warning, d2.Warning...)
}

func (d *DifferenceList) String(suppressWarning bool) string {
	var output string = ""
	if !suppressWarning && d.Warning != nil {
		output = output + "WARNING\n"
		for _, val := range d.Warning {
			output = output + val.String() + "\n"
		}
	}
	if d.Error != nil {
		output = output + "ERROR\n"
		for _, val := range d.Error {
			output = output + val.String() + "\n"
		}
	}
	return output
}

func (d *DifferenceList) isCompatible() bool {
	if d.Error == nil {
		return true
	}
	return false
}

func check(err error) {
	if err != nil {
		panic(err)
	}
}

func getChangesFileDP(newer, older []*descriptor.FileDescriptorProto) DifferenceList {
	var output DifferenceList
	for _, val1 := range newer { //loop through both arrays to see which fields existed in the older version too and which were newly added
		exist := false
		for _, val2 := range older {
			if val1.GetPackage() == val2.GetPackage() {
				exist = true
				output.merge(getChangesDP(val1.MessageType, val2.MessageType, strings.Split(*val1.Name, ".")[0])) //if proto exists in both files, compare it
			}
		}
		if !exist {
			output.addWarning(NonFieldIncompatibility, "", "", "", "", "Added proto file "+strings.Split(*val1.Name, ".")[0])
		}
	}
	for _, val1 := range older {
		exist := false
		for _, val2 := range newer {
			if val1.GetPackage() == val2.GetPackage() {
				exist = true
			}
		}
		if !exist {
			output.addWarning(NonFieldIncompatibility, "", "", "", "", "Removed proto file "+strings.Split(*val1.Name, ".")[0]) //if it exists only in the old proto, it has been removed
		}
	}
	return output
}

func getChangesDP(newer, older []*descriptor.DescriptorProto, prefix string) DifferenceList {
	var output DifferenceList
	for _, val1 := range newer {
		exist := false
		for _, val2 := range older {
			if *val1.Name == *val2.Name {
				exist = true
				output.merge(getChangesFieldDP(val1.Field, val2.Field, prefix+"/"+*val1.Name))
				output.merge(getChangesDP(val1.NestedType, val2.NestedType, prefix+"/"+*val1.Name))
				output.merge(getChangesEDP(val1.EnumType, val2.EnumType, prefix+"/"+*val1.Name))
			}
		}
		if !exist {
			output.addWarning(NonFieldIncompatibility, "", "", "", "", "Added message "+*val1.Name+" in "+prefix)
		}
	}
	for _, val1 := range older {
		exist := false
		for _, val2 := range newer {
			if *val1.Name == *val2.Name {
				exist = true
			}
		}
		if !exist {
			output.addWarning(NonFieldIncompatibility, "", "", "", "", "Removed message "+*val1.Name+" in "+prefix)
		}
	}
	return output
}

func getChangesEDP(newer, older []*descriptor.EnumDescriptorProto, prefix string) DifferenceList {
	var output DifferenceList
	for _, val1 := range newer {
		exist := false
		for _, val2 := range older {
			if *val1.Name == *val2.Name {
				exist = true
				output.merge(getChangesEVDP(val1.Value, val2.Value, prefix+"/"+*val1.Name))
			}
		}
		if !exist {
			output.addWarning(NonFieldIncompatibility, "", "", "", "", "Added enum "+*val1.Name+" in "+prefix)
		}
	}
	for _, val1 := range older {
		exist := false
		for _, val2 := range newer {
			if *val1.Name == *val2.Name {
				exist = true
			}
		}
		if !exist {
			output.addWarning(NonFieldIncompatibility, "", "", "", "", "Removed enum "+*val1.Name+" in "+prefix)
		}
	}
	return output
}

func getChangesFieldDP(newer, older []*descriptor.FieldDescriptorProto, prefix string) DifferenceList {
	var output DifferenceList
	for _, val1 := range newer { //loop through both arrays to see which fields existed in the older version too and which were newly added
		exist := false
		for _, val2 := range older {
			if *val1.Number == *val2.Number { //if message exists in both, check label, numeric tag and type for dissimilarities
				exist = true
				if val1.Label.String() != val2.Label.String() { //If field label changed add it to differences
					if *val1.Label == descriptor.FieldDescriptorProto_LABEL_REQUIRED {
						output.addError(ChangedLabel, val1.Label.String(), val2.Label.String(), prefix, strconv.Itoa(int(*val1.Number)), "")
					} else if *val2.Label == descriptor.FieldDescriptorProto_LABEL_REQUIRED {
						output.addError(ChangedLabel, val1.Label.String(), val2.Label.String(), prefix, strconv.Itoa(int(*val1.Number)), "")
					} else {
						output.addWarning(ChangedLabel, val1.Label.String(), val2.Label.String(), prefix, strconv.Itoa(int(*val1.Number)), "")
					}
				}
				if *val1.Name != *val2.Name {
					output.addWarning(ChangedName, *val1.Name, *val2.Name, prefix, strconv.Itoa(int(*val1.Number)), "")
				}
				if *val1.Type != *val2.Type {
					compatible := false
					if *val1.Type == descriptor.FieldDescriptorProto_TYPE_INT32 || *val1.Type == descriptor.FieldDescriptorProto_TYPE_INT64 || *val1.Type == descriptor.FieldDescriptorProto_TYPE_UINT32 || *val1.Type == descriptor.FieldDescriptorProto_TYPE_UINT64 || *val1.Type == descriptor.FieldDescriptorProto_TYPE_BOOL {
						if *val2.Type == descriptor.FieldDescriptorProto_TYPE_INT32 || *val2.Type == descriptor.FieldDescriptorProto_TYPE_INT64 || *val2.Type == descriptor.FieldDescriptorProto_TYPE_UINT32 || *val2.Type == descriptor.FieldDescriptorProto_TYPE_UINT64 || *val2.Type == descriptor.FieldDescriptorProto_TYPE_BOOL {
							compatible = true
						}
					}
					if *val1.Type == descriptor.FieldDescriptorProto_TYPE_SINT32 || *val1.Type == descriptor.FieldDescriptorProto_TYPE_SINT64 {
						if *val2.Type == descriptor.FieldDescriptorProto_TYPE_SINT32 || *val2.Type == descriptor.FieldDescriptorProto_TYPE_SINT64 {
							compatible = true
						}
					}
					if *val1.Type == descriptor.FieldDescriptorProto_TYPE_STRING || *val1.Type == descriptor.FieldDescriptorProto_TYPE_BYTES {
						if *val2.Type == descriptor.FieldDescriptorProto_TYPE_STRING || *val2.Type == descriptor.FieldDescriptorProto_TYPE_BYTES {
							compatible = true
						}
					}
					if *val1.Type == descriptor.FieldDescriptorProto_TYPE_FIXED32 || *val1.Type == descriptor.FieldDescriptorProto_TYPE_FIXED64 || *val1.Type == descriptor.FieldDescriptorProto_TYPE_SFIXED32 || *val1.Type == descriptor.FieldDescriptorProto_TYPE_SFIXED64 {
						if *val2.Type == descriptor.FieldDescriptorProto_TYPE_FIXED32 || *val2.Type == descriptor.FieldDescriptorProto_TYPE_FIXED64 || *val2.Type == descriptor.FieldDescriptorProto_TYPE_SFIXED32 || *val2.Type == descriptor.FieldDescriptorProto_TYPE_SFIXED64 {
							compatible = true
						}
					}
					if compatible {
						output.addWarning(ChangedType, val1.Type.String(), val2.Type.String(), prefix, strconv.Itoa(int(*val1.Number)), "")
					} else {
						output.addError(ChangedType, val1.Type.String(), val2.Type.String(), prefix, strconv.Itoa(int(*val1.Number)), "")
					}
				}
				if val1.DefaultValue != val2.DefaultValue {
					output.addWarning(ChangedDefault, val1.GetDefaultValue(), val2.GetDefaultValue(), prefix, strconv.Itoa(int(*val1.Number)), "")
				}
			}
		}
		if !exist {
			if *val1.Label == descriptor.FieldDescriptorProto_LABEL_REQUIRED {
				output.addError(AddedField, val1.Label.String(), "", prefix, strconv.Itoa(int(*val1.Number)), "")
			}
		}
	}
	for _, val1 := range older {
		exist := false
		for _, val2 := range newer {
			if *val1.Number == *val2.Number {
				exist = true
			}
		}
		if !exist {
			if *val1.Label == descriptor.FieldDescriptorProto_LABEL_REQUIRED {
				output.addError(RemovedField, val1.Label.String(), "", prefix, strconv.Itoa(int(*val1.Number)), "")
			} else {
				output.addWarning(RemovedField, val1.Label.String(), "", prefix, strconv.Itoa(int(*val1.Number)), " consider prefixing \"OBSOLETE_\" instead")
			}
		}
	}
	for _, val1 := range newer {
		for _, val2 := range older {
			if *val1.Name == *val2.Name {
				if *val1.Number != *val2.Number {
					output.addWarning(ChangedNumber, strconv.Itoa(int(*val1.Number)), strconv.Itoa(int(*val2.Number)), prefix, *val1.Name, "")
				}
			}
		}
	}
	return output
}

func getChangesEVDP(newer, older []*descriptor.EnumValueDescriptorProto, prefix string) DifferenceList {
	var output DifferenceList
	for _, val1 := range newer {
		exist := false
		for _, val2 := range older {
			if *val1.Name == *val2.Name {
				exist = true
			}
		}
		if !exist {
			output.addError(AddedField, *val1.Name, "", prefix, strconv.Itoa(int(*val1.Number)), "")
		}
	}
	for _, val1 := range older {
		exist := false
		for _, val2 := range newer {
			if *val1.Name == *val2.Name {
				exist = true
			}
		}
		if !exist {
			output.addError(RemovedField, *val1.Name, "", prefix, strconv.Itoa(int(*val1.Number)), "")
		}
	}
	return output
}

func main() {
	if len(os.Args) == 5 {
		newer, err1 := parser.ParseFile(os.Args[1], strings.Split(os.Args[2], ":")...)
		check(err1)
		older, err2 := parser.ParseFile(os.Args[3], strings.Split(os.Args[2], ":")...)
		check(err2)
		d := getChangesFileDP(newer.File, older.File)
		fmt.Print(d.String(false))
		if d.Error != nil {
			os.Exit(1)
		}
	} else if len(os.Args) == 6 {
		newer, err1 := parser.ParseFile(os.Args[1], strings.Split(os.Args[2], ":")...)
		check(err1)
		older, err2 := parser.ParseFile(os.Args[3], strings.Split(os.Args[2], ":")...)
		check(err2)
		d := getChangesFileDP(newer.File, older.File)
		fmt.Print(d.String(true))
		if d.Error != nil {
			os.Exit(1)
		}
	} else {
		fmt.Println(len(os.Args))
		fmt.Println("Use parameters {proto path 1} {proto 1 dependancies} {proto path 2} {proto 2 dependancies} if there is more than 1 dependency for a proto seperate them by \":\"")
		os.Exit(1)
	}
}
