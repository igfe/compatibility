## comments, suggestions
it has been years since I last worked on this code, but it slowly gained traction. If you have any feature requests or suggestions, you can email them to igfe@protonmail.com


## compatibility
Check the backwards compatiblity of google protocol buffers

According to the official language guide, protocol buffers can be updated and still remain compatible so long as certain rules are followed. This program tests two versions of a .proto file and displays an error if it is not compatible. 

The following rules (see https://developers.google.com/protocol-buffers/docs/proto#updating) tested for currently are 

>Don't change the numeric tags for any existing fields.

If this happens, the program might detect a name change or removed/added field related to this problem, however, a warning will be given if the new and old protos hava similarly named fields with different tags.

>Any new fields that you add should be optional or repeated. This means that any messages serialized by code using your "old" message format can be parsed by your new generated code, as they won't be missing any required elements. You should set up sensible default values for these elements so that new code can properly interact with messages generated by old code. Similarly, messages created by your new code can be parsed by your old code: old binaries simply ignore the new field when parsing. However, the unknown fields are not discarded, and if the message is later serialized, the unknown fields are serialized along with it – so if the message is passed on to new code, the new fields are still available.

If a numeric tag is added with a required label an error will be displayed

>Non-required fields can be removed, as long as the tag number is not used again in your updated message type (it may be better to rename the field instead, perhaps adding the prefix "OBSOLETE_", so that future users of your .proto can't accidentally reuse the number).

If a numeric tag with a required label is removed an error will be displayed. If a numeric tag with a non-required label is removed a warning will be given to rather rename it.

>A non-required field can be converted to an extension and vice versa, as long as the type and number stay the same.
    int32, uint32, int64, uint64, and bool are all compatible – this means you can change a field from one of these types to another without breaking forwards- or backwards-compatibility. If a number is parsed from the wire which doesn't fit in the corresponding type, you will get the same effect as if you had cast the number to that type in C++ (e.g. if a 64-bit number is read as an int32, it will be truncated to 32 bits).
    sint32 and sint64 are compatible with each other but are not compatible with the other integer types.
    string and bytes are compatible as long as the bytes are valid UTF-8.
    Embedded messages are compatible with bytes if the bytes contain an encoded version of the message.
    fixed32 is compatible with sfixed32, and fixed64 with sfixed64.
    
If any of these rules are broken an error is displayed.

>optional is compatible with repeated. Given serialized data of a repeated field as input, clients that expect this field to be optional will take the last input value if it's a primitive type field or merge all input elements if it's a message type field.

If either optional or repeated is changed to required, an error is displayed.

>Changing a default value is generally OK, as long as you remember that default values are never sent over the wire. Thus, if a program receives a message in which a particular field isn't set, the program will see the default value as it was defined in that program's version of the protocol. It will NOT see the default value that was defined in the sender's code.

If a default value differs a warning is displayed. Keep in mind that default values are deprecated in proto 3.
