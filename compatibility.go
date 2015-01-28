// Copyright [2015] [Ignazio Ferreira]

// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at

//     http://www.apache.org/licenses/LICENSE-2.0

// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package compatibility

import (
	"fmt"
	"github.com/gogo/protobuf/parser"
	descriptor "github.com/gogo/protobuf/protoc-gen-gogo/descriptor"
	"os"
	"strconv"
	"strings"
)

func check(err error) {
	if err != nil {
		panic(err)
	}
}

func GetDescriptor(path string, f *descriptor.FileDescriptorSet) *descriptor.DescriptorProto {
	pathA := strings.Split(path, ".")
	for _, v1 := range f.File {
		out := getDescriptor(pathA, v1.MessageType)
		if out != nil {
			return out
		}
	}
	fmt.Println("FAILED to find descriptor " + path)
	return nil
}

func getDescriptor(path []string, d []*descriptor.DescriptorProto) *descriptor.DescriptorProto {
	for _, val := range d {
		c := 0
		for ; path[c] == ""; c++ {
		}
		if *val.Name == path[c] {
			if len(path)-c == 1 {
				return val
			} else {
				return getDescriptor(path[c+1:], val.NestedType)
			}
		}
	}
	return nil
}

type Condition int

const (
	ChangedLabel            Condition = 1
	AddedField              Condition = 2
	RemovedField            Condition = 3
	ChangedName             Condition = 4
	ChangedType             Condition = 5
	ChangedNumber           Condition = 6
	ChangedDefault          Condition = 7
	ChangedTypeName         Condition = 8
	NonFieldIncompatibility Condition = 9
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
		return "Added Field nr " + d.qualifier + " in " + d.path + " of label " + d.newValue + d.message
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
	Error     []Difference
	Warning   []Difference
	Extension []Difference
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

func (d1 *DifferenceList) mergeExt(d2 DifferenceList) {
	d1.Extension = append(d1.Extension, d2.Error...)
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

type Comparer struct {
	Newer *descriptor.FileDescriptorSet
	Older *descriptor.FileDescriptorSet
}

func (c *Comparer) AppendExtensions() {
	for _, val := range c.Newer.File {
		for _, ext := range val.Extension {
			d := GetDescriptor(*ext.Extendee, c.Newer)
			d.Field = append(d.Field, ext)
		}
		for _, message := range val.MessageType {
			apndext(message, c.Newer)
		}
	}
	for _, val := range c.Older.File {
		for _, ext := range val.Extension {
			if ext != nil {
				d := GetDescriptor(*ext.Extendee, c.Older)
				d.Field = append(d.Field, ext)
			}
		}
		for _, message := range val.MessageType {
			apndext(message, c.Older)
		}
	}
}

func apndext(d *descriptor.DescriptorProto, c *descriptor.FileDescriptorSet) {
	for _, ext := range d.Extension {
		d := GetDescriptor(*ext.Extendee, c)
		d.Field = append(d.Field, ext)
	}
	for _, msg := range d.NestedType {
		apndext(msg, c)
	}
}

func (c *Comparer) Compare() DifferenceList {
	c.AppendExtensions()
	var output DifferenceList
	for _, val1 := range c.Newer.File { //loop through both arrays to see which fields existed in the older version too and which were newly added
		exist := false
		for _, val2 := range c.Older.File {
			if val1.GetPackage() == val2.GetPackage() {
				exist = true
				output.merge(getChangesDP(val1.MessageType, val2.MessageType, "", *c)) //if proto exists in both files, compare it
			}
		}
		if !exist {
			output.addWarning(NonFieldIncompatibility, "", "", "", "", "Added proto file "+strings.Split(*val1.Name, ".")[0])
		}
	}
	for _, val1 := range c.Older.File {
		exist := false
		for _, val2 := range c.Newer.File {
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

func getChangesDP(newer, older []*descriptor.DescriptorProto, path string, c Comparer) DifferenceList {
	var output DifferenceList
	for _, val1 := range newer {
		exist := false
		for _, val2 := range older {
			if *val1.Name == *val2.Name {
				exist = true
				output.merge(getChangesFieldDP(val1.Field, val2.Field, val1.ExtensionRange, val2.ExtensionRange, path+"."+*val1.Name))
				output.merge(getChangesDP(val1.NestedType, val2.NestedType, path+"."+*val1.Name, c))
				output.merge(getChangesEDP(val1.EnumType, val2.EnumType, path+"."+*val1.Name))
			}
		}
		if !exist {
			output.addWarning(NonFieldIncompatibility, "", "", "", "", "Added message "+*val1.Name+" in "+path)
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
			output.addWarning(NonFieldIncompatibility, "", "", "", "", "Removed message "+*val1.Name+" in "+path)
		}
	}
	return output
}

func getChangesEDP(newer, older []*descriptor.EnumDescriptorProto, path string) DifferenceList {
	var output DifferenceList
	for _, val1 := range newer {
		exist := false
		for _, val2 := range older {
			if *val1.Name == *val2.Name {
				exist = true
				output.merge(getChangesEVDP(val1.Value, val2.Value, path+"."+*val1.Name))
			}
		}
		if !exist {
			output.addWarning(NonFieldIncompatibility, "", "", "", "", "Added enum "+*val1.Name+" in "+path)
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
			output.addWarning(NonFieldIncompatibility, "", "", "", "", "Removed enum "+*val1.Name+" in "+path)
		}
	}
	return output
}

func getChangesFieldDP(newer, older []*descriptor.FieldDescriptorProto, newEx, oldEx []*descriptor.DescriptorProto_ExtensionRange, path string) DifferenceList {
	var output DifferenceList
	for _, val1 := range newer { //loop through both arrays to see which fields existed in the older version too and which were newly added
		exist := false
		for _, val2 := range older {
			if *val1.Number == *val2.Number { //if message exists in both, check label, numeric tag and type for dissimilarities
				exist = true
				output.merge(compareFields(*val1, *val2, path))
			}
		}
		if !exist {
			if *val1.Label == descriptor.FieldDescriptorProto_LABEL_REQUIRED {
				output.addError(AddedField, val1.Label.String(), "", path, strconv.Itoa(int(*val1.Number)), "")
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
				output.addError(RemovedField, val1.Label.String(), "", path, strconv.Itoa(int(*val1.Number)), "")
			} else {
				output.addWarning(RemovedField, val1.Label.String(), "", path, strconv.Itoa(int(*val1.Number)), " consider pathing \"OBSOLETE_\" instead")
			}
		}
	}
	for _, val1 := range newer {
		for _, val2 := range older {
			if *val1.Name == *val2.Name {
				if *val1.Number != *val2.Number {
					output.addWarning(ChangedNumber, strconv.Itoa(int(*val1.Number)), strconv.Itoa(int(*val2.Number)), path, *val1.Name, "")
				}
			}
		}
	}
	return output
}

func compareFields(val1, val2 descriptor.FieldDescriptorProto, path string) DifferenceList {
	var output DifferenceList
	if val1.Label.String() != val2.Label.String() { //If field label changed add it to differences
		if *val1.Label == descriptor.FieldDescriptorProto_LABEL_REQUIRED {
			output.addError(ChangedLabel, val1.Label.String(), val2.Label.String(), path, strconv.Itoa(int(*val1.Number)), "")
		} else if *val2.Label == descriptor.FieldDescriptorProto_LABEL_REQUIRED {
			output.addError(ChangedLabel, val1.Label.String(), val2.Label.String(), path, strconv.Itoa(int(*val1.Number)), "")
		} else {
			output.addWarning(ChangedLabel, val1.Label.String(), val2.Label.String(), path, strconv.Itoa(int(*val1.Number)), "")
		}
	}
	if *val1.Name != *val2.Name {
		output.addWarning(ChangedName, *val1.Name, *val2.Name, path, strconv.Itoa(int(*val1.Number)), "")
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
			output.addWarning(ChangedType, val1.Type.String(), val2.Type.String(), path, strconv.Itoa(int(*val1.Number)), "")
		} else {
			output.addError(ChangedType, val1.Type.String(), val2.Type.String(), path, strconv.Itoa(int(*val1.Number)), "")
		}
	}
	if val1.DefaultValue != val2.DefaultValue {
		output.addWarning(ChangedDefault, val1.GetDefaultValue(), val2.GetDefaultValue(), path, strconv.Itoa(int(*val1.Number)), "")
	}
	return output
}

func getChangesEVDP(newer, older []*descriptor.EnumValueDescriptorProto, path string) DifferenceList {
	var output DifferenceList
	for _, val1 := range newer {
		exist := false
		for _, val2 := range older {
			if *val1.Name == *val2.Name {
				exist = true
			}
		}
		if !exist {
			output.addError(AddedField, *val1.Name, "", path, strconv.Itoa(int(*val1.Number)), "")
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
			output.addError(RemovedField, *val1.Name, "", path, strconv.Itoa(int(*val1.Number)), "")
		}
	}
	return output
}

func isExtension(tag int, ext []*descriptor.DescriptorProto_ExtensionRange) bool {
	for _, val1 := range ext {
		if tag >= int(*val1.Start) && tag <= int(*val1.End) {
			return true
		}
	}
	return false
}

func main() {
	if len(os.Args) == 5 {
		newer, err1 := parser.ParseFile(os.Args[1], strings.Split(os.Args[2], ":")...)
		check(err1)
		older, err2 := parser.ParseFile(os.Args[3], strings.Split(os.Args[4], ":")...)
		check(err2)
		c := Comparer{newer, older}
		d := c.Compare()
		fmt.Print(d)
		if d.Error != nil {
			os.Exit(1)
		}
	} else if len(os.Args) == 6 {
		newer, err1 := parser.ParseFile(os.Args[1], strings.Split(os.Args[2], ":")...)
		check(err1)
		older, err2 := parser.ParseFile(os.Args[3], strings.Split(os.Args[4], ":")...)
		check(err2)
		c := Comparer{newer, older}
		d := c.Compare()
		fmt.Print(d.String(true))
		if d.Error != nil {
			os.Exit(1)
		}
	} else if len(os.Args) == 1 {
		newer, err1 := parser.ParseFile("./ExtensionProtos/Changes/p.proto", "./ExtensionProtos/Changes")
		check(err1)
		older, err2 := parser.ParseFile("./ExtensionProtos/p.proto", "./ExtensionProtos/")
		check(err2)
		c := Comparer{newer, older}
		d := c.Compare()
		fmt.Print(d.String(false))
		if d.Error != nil {
			os.Exit(1)
		}
	} else {
		fmt.Println(len(os.Args))
		fmt.Println("Use either 0 parameters for hard coded imports or 4,5 paramters to pass relative filepath")
		fmt.Println("Use parameters {proto path 1} {proto 1 dependancies} {proto path 2} {proto 2 dependancies} if there is more than 1 dependency for a proto seperate them by \":\"")
		os.Exit(1)
	}
}
