# plist
Library for parsing of xml property list in golang.

This is very simple, but useful parser of plist (https://en.wikipedia.org/wiki/Property_list) in xml files. Format is frequently used on Mac OS.

Parser can unmarshal all known types (including arrays, timestamps and base64 encoded data). I'm just working on support of dictionary.

TODOs:
- support of dicts
- don't have a problem with comments and whitespaces in values
- data could be encoded not not in io.Writers, but also in something like json.RawMessage
- use function xml.DecodeElement
- better names in tests
- write examples of using the library into this file
