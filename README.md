# plist
Library for parsing of xml property list in golang.

This is very simple, but useful parser of plist (https://en.wikipedia.org/wiki/Property_list) in xml files. Format is frequently used on Mac OS.

Parser can unmarshal all known types (including arrays, timestamps and base64 encoded data). I'm just working on support of dictionary.

TODOs:
- support embeeded types
- data could be encoded not not in io.Writers, but also in something like json.RawMessage
- allow <data> to be any interface, not only structs (maybe *struct or channel)
- write examples of using the library into this file
- recovery from panic
