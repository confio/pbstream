# ProtoBuf Streamer

Protobuf is a widespread binary codec and there are many
libraries for different languages to marshal/unmarshal data.
However, this can be a bit heavy-weight for some applications
that just want to quickly verify a field or two.

This package will extract arbitrary data from a protobuf encoded
message, using only the field indentifiers as arguments (not
even the original .proto file). It should use minimal memory
and resources and return just the desired field without
parsing the whole structure.

One application is a router that just wants to inspect one
field of a message before sending it along.

Another application is an HSM, such as ledger, that wants to
display information from one or two fields, before signing
the raw bytes.

The public API attempts to be minimal, and [can be viewed on godoc.](https://godoc.org/github.com/confio/pbstream)

Basic parsing for protobuf objects:

- [x] Extract field by number
- [x] Extract embedded fields by following path of field numbers
- [x] Parse varints
- [x] Parse fixed ints
- [x] Parse floats
- [x] Parse strings and bytes
- [x] Parse enums / bools (as int32)
- [x] Unpack sint32/64
- [x] Parse packed repeated fields (series of numbers)
- [ ] Parse repeated structs
- [ ] Parse fields embedded inside repeated structs
- [ ] Produce iterator-like parser for repeated

Handle ugly data:
- [ ] Properly handle repeated copies of non-repeated fields (last write wins)
- [ ] Validate with multiple protoc encoders
- [ ] Fuzz results alongside real proto.Unmarshal

Minimize memory usage:
- [x] Only store pointer to original buffer
- [x] Allocate only on parsing numeric types
- [ ] Handle repeated types with iterator
- [ ] Allow parsing input stream (not even have original structure in memory)
- [ ] Port to minimal ANSI C for embedded systems
