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
	"code.google.com/p/gogoprotobuf/parser"
	"strconv"
	"testing"
)

func TestByteString(t *testing.T) {
	newer, err1 := parser.ParseFile("./TestProtos/BytesStringProtos/Changes/Original.proto", "./TestProtos/BytesStringProtos/Changes")
	check(err1)
	older, err2 := parser.ParseFile("./TestProtos/BytesStringProtos/Original.proto", "./TestProtos/BytesStringProtos")
	check(err2)
	d := getChangesFileDP(newer.File, older.File)
	if !d.isCompatible() {
		t.Error("Changes to BYTES or STRING broke the compatibility")
	}
}

func TestFixed(t *testing.T) {
	newer, err1 := parser.ParseFile("./TestProtos/FixedProtos/Changes/Original.proto", "./TestProtos/FixedProtos/Changes")
	check(err1)
	older, err2 := parser.ParseFile("./TestProtos/FixedProtos/Original.proto", "./TestProtos/FixedProtos")
	check(err2)
	d := getChangesFileDP(newer.File, older.File)
	if !d.isCompatible() {
		t.Error("Changes to fixed integer types broke the compatibility")
	}
}

func TestIncompatibility(t *testing.T) {
	newer, err1 := parser.ParseFile("./TestProtos/Incompatibility/Original.proto", "./TestProtos/Incompatibility/")
	check(err1)
	older, err2 := parser.ParseFile("./TestProtos/Incompatibility/Changes/Original.proto", "./TestProtos/Incompatibility/Changes/")
	check(err2)
	d := getChangesFileDP(newer.File, older.File)
	if len(d.Error) != 87 {
		t.Error(strconv.Itoa(len(d.Error)) + " incompatibilities out of 87")
	}
}

func TestOptionalRepeated(t *testing.T) {
	newer, err1 := parser.ParseFile("./TestProtos/OptionalRepeatedProtos/Changes/Original.proto", "./TestProtos/OptionalRepeatedProtos/Changes")
	check(err1)
	older, err2 := parser.ParseFile("./TestProtos/OptionalRepeatedProtos/Original.proto", "./TestProtos/OptionalRepeatedProtos")
	check(err2)
	d := getChangesFileDP(newer.File, older.File)
	if !d.isCompatible() {
		t.Error("Switching between labels broke the compatibility")
	}
}

func TestInt(t *testing.T) {
	newer, err1 := parser.ParseFile("./TestProtos/IntProtos/Changes/Original.proto", "./TestProtos/IntProtos/Changes")
	check(err1)
	older, err2 := parser.ParseFile("./TestProtos/IntProtos/Original.proto", "./TestProtos/IntProtos")
	check(err2)
	d := getChangesFileDP(newer.File, older.File)
	if !d.isCompatible() {
		t.Error("Changes to integer types broke the compatibility")
	}
}

func TestNestedAdded(t *testing.T) {
	newer, err1 := parser.ParseFile("./TestProtos/NestedProtos/NestedAdded/Changes/Original.proto", "./TestProtos/NestedProtos/NestedAdded/Changes")
	check(err1)
	older, err2 := parser.ParseFile("./TestProtos/NestedProtos/NestedAdded/Original.proto", "./TestProtos/NestedProtos/NestedAdded")
	check(err2)
	d := getChangesFileDP(newer.File, older.File)
	if len(d.Error) != 2 {
		t.Error("Expected 2 errors, found " + strconv.Itoa(len(d.Error)))
	}
	for _, val := range d.Error {
		if val.condition != AddedField {
			t.Error("Incompatible error condition: Not AddedField")
		}
	}
}

func TestNestedLabel(t *testing.T) {
	newer, err1 := parser.ParseFile("./TestProtos/NestedProtos/NestedLabel/Changes/Original.proto", "./TestProtos/NestedProtos/NestedLabel/Changes")
	check(err1)
	older, err2 := parser.ParseFile("./TestProtos/NestedProtos/NestedLabel/Original.proto", "./TestProtos/NestedProtos/NestedLabel")
	check(err2)
	d := getChangesFileDP(newer.File, older.File)
	if len(d.Error) != 4 {
		t.Error("Expected 4 errors, found " + strconv.Itoa(len(d.Error)))
	}
	for _, val := range d.Error {
		if val.condition != ChangedLabel {
			t.Error("Incompatible error condition: Not ChangedLabel")
		}
	}
}

func TestNestedRemoved(t *testing.T) {
	newer, err1 := parser.ParseFile("./TestProtos/NestedProtos/NestedRemoved/Changes/Original.proto", "./TestProtos/NestedProtos/NestedRemoved/Changes")
	check(err1)
	older, err2 := parser.ParseFile("./TestProtos/NestedProtos/NestedRemoved/Original.proto", "./TestProtos/NestedProtos/NestedRemoved")
	check(err2)
	d := getChangesFileDP(newer.File, older.File)
	if len(d.Error) != 2 {
		t.Error("Expected 2 errors, found " + strconv.Itoa(len(d.Error)))
	}
	for _, val := range d.Error {
		if val.condition != RemovedField {
			t.Error("Incompatible error condition: Not RemovedField")
		}
	}
}
