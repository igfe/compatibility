package compatibility

import (
	"fmt"
	descriptor "github.com/gogo/protobuf/protoc-gen-gogo/descriptor"
)

//traverse filedescriptorproto
//traverse descriptorproto
//traverse fields

//get extendee
//get old extendee extensionrange
//get fieldnumbers

//compare extendee range to extensionfieldnumbers

//split extendee name
//traverse messages
//compare new-old
//compare old-new

func extCompareFiles(d1, d2 []descriptor.FileDescriptorProto) {
	for _, val1 := range d1 {

	}
}

func extCompareMessages(m1, m2 []*descriptor.DescriptorProto, prefix string) DifferenceList {
	var output DifferenceList
	for _, val1 := range newer {
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

func main() {

}
