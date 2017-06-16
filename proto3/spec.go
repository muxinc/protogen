package proto3

import (
	"bytes"
	"errors"
	"fmt"
	"sort"
)

// ImportType applies to a package import statement
type ImportType string

// NameType applies to a field name
type NameType string

// TagType applies to a field tag value
type TagType uint8

// FieldType applies to the data-type of a field value
type FieldType uint8

// FieldRule specifies additional rules (e.g. repeated) that can be set on a field
type FieldRule uint8

// Rules that can be applied to message fields
// https://developers.google.com/protocol-buffers/docs/proto3#specifying-field-rules
const (
	None FieldRule = iota
	Repeated
)

// Built-in field types
// https://developers.google.com/protocol-buffers/docs/proto3#scalar
const (
	DoubleType FieldType = iota
	FloatType
	Int32Type
	Int64Type
	UInt32Type
	UInt64Type
	SInt32Type
	SInt64Type
	Fixed32Type
	Fixed64Type
	SFixed32Type
	SFixed64Type
	BoolType
	StringType
	BytesType
)

// Reserved describes a tag that can be written to a Protobuf.
type Reserved interface {
	Validate() error
	Write() (string, error)
}

// Field describes a Protobuf message field.
type Field interface {
	Validate() error
	Write() (string, error)
}

// Spec represents a top-level Protobuf specification.
type Spec struct {
	Package     string       // https://developers.google.com/protocol-buffers/docs/proto3#packages
	JavaPackage string       // https://developers.google.com/protocol-buffers/docs/reference/java-generated#package
	Imports     []ImportType // https://developers.google.com/protocol-buffers/docs/proto3#importing-definitions
	Messages    []Message
	Enums       []Enum
}

// Message is a single Protobuf message definition.
type Message struct {
	Name           string
	Comment        string
	Messages       []Message
	ReservedValues []Reserved
	Fields         []Field
	OneOfs         []OneOf
	Enums          []Enum
}

// ReservedName is a field name that is reserved within a message type and cannot be reused.
// https://developers.google.com/protocol-buffers/docs/proto3#reserved
type ReservedName struct {
	Name NameType
}

// ReservedTagValue is a single field tag value that is reserved within a message type and cannot be reused.
// https://developers.google.com/protocol-buffers/docs/proto3#reserved
type ReservedTagValue struct {
	Tag TagType
}

// ReservedTagRange is a range of numeric tag values that are reserved within a message type and cannot be reused.
// https://developers.google.com/protocol-buffers/docs/proto3#reserved
type ReservedTagRange struct {
	LowerTag TagType
	UpperTag TagType
}

// CustomField is a message field with an unchecked, custom type. This can be used to define fields that
// use imported types.
type CustomField struct {
	Name    NameType
	Tag     TagType
	Rule    FieldRule
	Comment string
	Typing  string
}

// ScalarField is a message field that uses a built-in protobuf type.
type ScalarField struct {
	Name    NameType
	Tag     TagType
	Rule    FieldRule
	Comment string
	Typing  FieldType
}

// MapField is a message field that maps built-in protobuf type as key-value pairs
// https://developers.google.com/protocol-buffers/docs/proto3#maps
type MapField struct {
	Name        NameType
	Tag         TagType
	Rule        FieldRule
	Comment     string
	KeyTyping   FieldType
	ValueTyping FieldType
}

// CustomMapField is a message field that maps between a built-in protobuf type as
// the key and a custom type as the value.
// https://developers.google.com/protocol-buffers/docs/proto3#maps
type CustomMapField struct {
	Name        NameType
	Tag         TagType
	Rule        FieldRule
	Comment     string
	KeyTyping   FieldType
	ValueTyping string
}

// OneOf defines a set of fields for which only the most-recently-set field will be used.
// https://developers.google.com/protocol-buffers/docs/proto3#oneof
type OneOf struct {
	Name    NameType
	Fields  []Field
	Comment string
}

// Enum defines an enumeration type of a set of values.
// https://developers.google.com/protocol-buffers/docs/proto3#enum
type Enum struct {
	Name       NameType
	Values     []EnumValue
	AllowAlias bool
	Comment    string
}

// EnumValue describes a single enumerated value within an enumeration.
// https://developers.google.com/protocol-buffers/docs/proto3#enum
type EnumValue struct {
	Name    NameType
	Tag     TagType
	Comment string
}

// WRITERS

// Write turns the specification into a string.
func (s *Spec) Write() (string, error) {
	if err := s.Validate(); err != nil {
		return "", err
	}

	var buffer bytes.Buffer
	buffer.WriteString("syntax = \"proto3\";\n")
	if len(s.Package) > 0 {
		buffer.WriteString(fmt.Sprintf("package %s;\n", s.Package))
	}
	if len(s.JavaPackage) > 0 {
		buffer.WriteString(fmt.Sprintf("option java_package = \"%s\";\n", s.JavaPackage))
	}
	for _, importPackage := range s.Imports {
		buffer.WriteString(fmt.Sprintf("import \"%s\";\n", importPackage))
	}

	for _, v := range s.Enums {
		v, err := v.Write(0)
		if err != nil {
			return "", err
		}
		buffer.WriteString(fmt.Sprintf("\n%s\n", v))
	}

	for _, msg := range s.Messages {
		msgSpec, err := msg.Write(0) // write message at level zero (0)
		if err != nil {
			return "", err
		}
		buffer.WriteString(fmt.Sprintf("\n%s\n", msgSpec))
	}
	return buffer.String(), nil
}

// Write the message specification as a string at a given indentation level.
func (m *Message) Write(level int) (string, error) {
	if err := m.Validate(); err != nil {
		return "", err
	}

	var buffer bytes.Buffer
	if m.Comment != "" {
		buffer.WriteString(fmt.Sprintf("%s// %s\n", indentLevel(level), m.Comment))
	}
	buffer.WriteString(fmt.Sprintf("%smessage %s {\n", indentLevel(level), m.Name))

	// NESTED MESSAGE TYPES
	for _, msg := range m.Messages {
		msgSpec, err := msg.Write(level + 1)
		if err != nil {
			return "", err
		}
		buffer.WriteString(fmt.Sprintf("%s\n\n", msgSpec))
	}

	// ENUMS
	for _, v := range m.Enums {
		v, err := v.Write(level + 1)
		if err != nil {
			return "", err
		}
		buffer.WriteString(fmt.Sprintf("%s\n\n", v))
	}

	// RESERVED TAGS
	if len(m.ReservedValues) > 0 {
		for _, reservedValue := range m.ReservedValues {
			v, err := reservedValue.Write()
			if err != nil {
				return "", err
			}
			buffer.WriteString(fmt.Sprintf("%sreserved %s;\n", indentLevel(level+1), v))
		}
		buffer.WriteString("\n")
	}

	// FIELDS
	if len(m.Fields) > 0 {
		for _, v := range m.Fields {
			v, err := v.Write()
			if err != nil {
				return "", err
			}
			buffer.WriteString(fmt.Sprintf("%s%s\n", indentLevel(level+1), v))
		}
		buffer.WriteString("\n")
	}

	// ONE-OF FIELDS
	for _, v := range m.OneOfs {
		v, err := v.Write(level + 1)
		if err != nil {
			return "", err
		}
		buffer.WriteString(fmt.Sprintf("%s\n", v))
	}

	buffer.WriteString(fmt.Sprintf("%s}", indentLevel(level)))
	return buffer.String(), nil
}

// Write a ReservedName as a string
func (r ReservedName) Write() (string, error) {
	return fmt.Sprintf("\"%s\"", r.Name), nil
}

// Write a ReservedTagValue as a string
func (r ReservedTagValue) Write() (string, error) {
	return fmt.Sprintf("%d", r.Tag), nil
}

// Write a ReservedTagRange as a string
func (r ReservedTagRange) Write() (string, error) {
	return fmt.Sprintf("%d to %d", r.LowerTag, r.UpperTag), nil
}

// Write a CustomField as a string
func (c CustomField) Write() (string, error) {
	v := fmt.Sprintf("%s%s %s = %d;", c.Rule.Write(), c.Typing, c.Name, c.Tag)
	if c.Comment != "" {
		v = fmt.Sprintf("%s   // %s", v, c.Comment)
	}
	return v, nil
}

// Write a ScalarField as a string
func (s ScalarField) Write() (string, error) {
	v := fmt.Sprintf("%s%s %s = %d;", s.Rule.Write(), s.Typing.Write(), s.Name, s.Tag)
	if s.Comment != "" {
		v = fmt.Sprintf("%s   // %s", v, s.Comment)
	}
	return v, nil
}

// Write a MapField as a string
func (m MapField) Write() (string, error) {
	v := fmt.Sprintf("%smap<%s, %s> %s = %d;", m.Rule.Write(), m.KeyTyping.Write(), m.ValueTyping.Write(), m.Name, m.Tag)
	if m.Comment != "" {
		v = fmt.Sprintf("%s   // %s", v, m.Comment)
	}
	return v, nil
}

// Write a CustomMapField as a string
func (c CustomMapField) Write() (string, error) {
	v := fmt.Sprintf("%smap<%s, %s> %s = %d;", c.Rule.Write(), c.KeyTyping.Write(), c.ValueTyping, c.Name, c.Tag)
	if c.Comment != "" {
		v = fmt.Sprintf("%s   // %s", v, c.Comment)
	}
	return v, nil
}

// Write a CustomMapField as a string
func (e Enum) Write(level int) (string, error) {
	sort.Sort(e)

	var v string
	if e.Comment != "" {
		v = fmt.Sprintf("%s// %s\n", indentLevel(level), e.Comment)
	}
	v = fmt.Sprintf("%s%senum %s {\n", v, indentLevel(level), e.Name)
	if e.AllowAlias {
		v = fmt.Sprintf("%s%soption allow_alias = true;\n", v, indentLevel(level+1))
	}
	for _, enumValue := range e.Values {
		v = fmt.Sprintf("%s%s%s = %d;", v, indentLevel(level+1), enumValue.Name, enumValue.Tag)
		if enumValue.Comment != "" {
			v = fmt.Sprintf("%s   // %s", v, enumValue.Comment)
		}
		v = fmt.Sprintf("%s\n", v)
	}
	v = fmt.Sprintf("%s%s}", v, indentLevel(level))
	return v, nil
}

func (o OneOf) Write(level int) (string, error) {
	var v string
	if o.Comment != "" {
		v = fmt.Sprintf("%s// %s\n", indentLevel(level), o.Comment)
	}
	v = fmt.Sprintf("%s%soneof %s {\n", v, indentLevel(level), o.Name)

	for _, f := range o.Fields {
		s, err := f.Write()
		if err != nil {
			return "", err
		}
		v = fmt.Sprintf("%s%s%s\n", v, indentLevel(level+1), s)
	}

	v = fmt.Sprintf("%s%s}", v, indentLevel(level))
	return v, nil
}

// Write a FieldRule as a string
func (f *FieldRule) Write() string {
	switch *f {
	case None:
		return ""
	case Repeated:
		return "repeated "
	default:
		return ""
	}
}

// Write a FieldType as a string
func (f *FieldType) Write() string {
	switch *f {
	case DoubleType:
		return "double"
	case FloatType:
		return "float"
	case Int32Type:
		return "int32"
	case Int64Type:
		return "int64"
	case UInt32Type:
		return "uint32"
	case UInt64Type:
		return "uint64"
	case SInt32Type:
		return "sint32"
	case SInt64Type:
		return "sint64"
	case Fixed32Type:
		return "fixed32"
	case Fixed64Type:
		return "fixed64"
	case SFixed32Type:
		return "sfixed32"
	case SFixed64Type:
		return "sfixed64"
	case BoolType:
		return "bool"
	case StringType:
		return "string"
	case BytesType:
		return "bytes"
	default:
		return ""
	}
}

// VALIDATORS

// Validate spec
func (s *Spec) Validate() error {
	if len(s.Messages) == 0 {
		return errors.New("Spec must contain at least one message")
	}
	for _, msg := range s.Messages {
		if err := msg.Validate(); err != nil {
			return err
		}
	}
	for _, v := range s.Enums {
		if err := v.Validate(); err != nil {
			return err
		}
	}
	return nil
}

// Validate the attributes of a message, including all children that can be validated individually.
func (m Message) Validate() error {
	if m.Name == "" {
		return errors.New("Message name cannot be empty")
	}
	for _, v := range m.Fields {
		if err := v.Validate(); err != nil {
			return err
		}
	}
	for _, v := range m.Messages {
		if err := v.Validate(); err != nil {
			return err
		}
	}
	for _, v := range m.ReservedValues {
		if err := v.Validate(); err != nil {
			return err
		}
	}
	for _, v := range m.Enums {
		if err := v.Validate(); err != nil {
			return err
		}
	}
	return nil
}

// Validate field attributes
func (s ScalarField) Validate() error {
	if s.Name == "" {
		return errors.New("Scalar field must have a non-empty name")
	}
	return nil
}

// Validate field attributes
func (r ReservedName) Validate() error {
	if r.Name == "" {
		return errors.New("ReservedName field must have a non-empty name")
	}
	return nil
}

// Validate field attributes
func (r ReservedTagValue) Validate() error {
	return nil
}

// Validate field attributes
func (r ReservedTagRange) Validate() error {
	if r.LowerTag < 0 {
		return errors.New("ReservedTagRange lower-tag must be greater-than-or-equal to zero")
	}
	if r.LowerTag >= r.UpperTag {
		return errors.New("ReservedTagRange upper-tag must be greater-than lower-tag")
	}
	return nil
}

// Validate field attributes
func (c CustomField) Validate() error {
	if c.Name == "" {
		return errors.New("CustomField name must have non-empty name")
	}
	if c.Tag < 0 {
		return errors.New("CustomField must have positive integer for tag")
	}
	return nil
}

// Validate field attributes
func (c CustomMapField) Validate() error {
	if c.Name == "" {
		return errors.New("CustomMapField name must have non-empty name")
	}
	if c.KeyTyping < 0 || c.KeyTyping == DoubleType || c.KeyTyping == FloatType || c.KeyTyping == BytesType {
		return fmt.Errorf("Map field %s must use a scalar integral or string type for the map key", c.Name)
	}
	if c.Rule == Repeated {
		return errors.New("CustomMapField cannot use repeated rule")
	}
	return nil
}

// Validate map attributes
func (m MapField) Validate() error {
	if m.Name == "" {
		return errors.New("MapField must have a non-empty name")
	}
	if m.KeyTyping < 0 || m.KeyTyping == DoubleType || m.KeyTyping == FloatType || m.KeyTyping == BytesType {
		return fmt.Errorf("Map field %s must use a scalar integral or string type for the map key", m.Name)
	}
	if m.ValueTyping < 0 {
		return fmt.Errorf("Map field %s must have a type specified for the map value", m.Name)
	}
	if m.Rule == Repeated {
		return errors.New("MapField cannot use repeated rule")
	}
	return nil
}

// Validate enum attributes
func (e *Enum) Validate() error {
	if e.Name == "" {
		return errors.New("Enum must have a non-empty name")
	}
	if len(e.Values) == 0 {
		return errors.New("Enum must have non-empty set of values")
	}
	if e.AllowAlias == false {
		tags := make(map[TagType]NameType)
		for _, v := range e.Values {
			if _, exists := tags[v.Tag]; exists {
				return fmt.Errorf("Enum value has tag that is already in use while aliasing is not allowed: %s", v.Name)
			}
			tags[v.Tag] = v.Name
		}
	}
	return nil
}

// Validate oneof attributes
func (o OneOf) Validate() error {
	if o.Name == "" {
		return errors.New("OneOf must have a non-empty name")
	}
	if len(o.Fields) == 0 {
		return errors.New("OneOf must have non-empty set of values")
	}
	return nil
}

// FORMATTING

func indentLevel(level int) string {
	var buffer bytes.Buffer
	for i := 0; i < level; i++ {
		buffer.WriteString("  ")
	}
	return buffer.String()
}

// Len reports the number of enum values.
func (e Enum) Len() int { return len(e.Values) }

// Swap entries in the enum at the given positions.
func (e Enum) Swap(i, j int) { e.Values[i], e.Values[j] = e.Values[j], e.Values[i] }

// Less returns true iff the tag at the first position is less than the tag at the second position.
func (e Enum) Less(i, j int) bool { return e.Values[i].Tag < e.Values[j].Tag }
