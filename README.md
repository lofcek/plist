# plist
Library for parsing of xml property list in golang.

This is very simple, but useful parser of plist (https://en.wikipedia.org/wiki/Property_list) in xml files. Format is frequently used on Mac OS.

Parser can unmarshal all known types (including arrays, timestamps and base64 encoded data). I'm just working on support of dictionary.

TODOs:
- full support of dicts
- support embeeded types
- data could be encoded not not in io.Writers, but also in something like json.RawMessage
- write examples of using the library into this file
- support also member like pointer to something instead of direct value
- recovery from panic
- add names of plist items into tags
