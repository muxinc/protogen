# protogen [![GoDoc](https://godoc.org/github.com/go-kit/kit?status.svg)](https://godoc.org/github.com/muxinc/protogen/proto3) [![Go Report Card](https://goreportcard.com/badge/github.com/muxinc/protogen)](https://goreportcard.com/report/github.com/muxinc/protogen) [![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)

[Protobuf](https://developers.google.com/protocol-buffers/docs/proto3) Specification Generator written in Go

This library was developed by [Mux](https://www.mux.com/) to programmatically generate Protobuf specifications.

Mux has a large number of message fields that are used in Protobuf-encoded message-types exchanged throughout our system. Historically the Protobuf specifications had been written by hand, which led to subtle differences in the naming of fields across message specs, not to mention the tedium and error-prone nature of manually editing specs.

The objective of this library is to generate Protobuf specifications automatically from a central repository that keeps track of the many fields we work with and the messages that use them.
