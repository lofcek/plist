# plist
Library for parsing of xml property list in golang.

This is very simple, but useful parser of plist (https://en.wikipedia.org/wiki/Property_list) in xml files. Format is frequently used on Mac OS.

Parser can unmarshal all known types (including arrays, timestamps and base64 encoded data). I'm just working on support of dictionary.

TODOs:
- verify whether dict really has elements <dict></dict>
- support embeeded types
- data could be encoded not not in io.Writers, but also in something like json.RawMessage
- write examples of using the library into this file
- recovery from panic
- add names of plist items into tags
- ignore <?xml version="1.0" encoding="UTF-8"?> or <!DOCTYPE plist SYSTEM "file://localhost/System/Library/DTDs/PropertyList.dtd">
