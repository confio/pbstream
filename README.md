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
