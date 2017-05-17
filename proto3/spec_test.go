package proto3

import (
	"fmt"
	"testing"
)

func TestScalarField_Validate(t *testing.T) {
	type fields struct {
		Name    NameType
		Tag     TagType
		Rule    FieldRule
		Comment string
		Typing  FieldType
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name:    "Valid Scalar field",
			fields:  fields{Name: "MyMap", Tag: 1, Typing: STRING_TYPE},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		s := &ScalarField{
			Name:    tt.fields.Name,
			Tag:     tt.fields.Tag,
			Rule:    tt.fields.Rule,
			Comment: tt.fields.Comment,
			Typing:  tt.fields.Typing,
		}
		if err := s.Validate(); (err != nil) != tt.wantErr {
			t.Errorf("%q. ScalarField.Validate() error = %v, wantErr %v", tt.name, err, tt.wantErr)
		}
	}
}

func TestSpec_Write(t *testing.T) {
	type fields struct {
		Package  string
		Imports  []ImportType
		Messages []Message
	}
	tests := []struct {
		name    string
		fields  fields
		want    string
		wantErr bool
	}{
		{
			name:    "Invalid spec with zero messages",
			fields:  fields{},
			wantErr: true,
		},
		{
			name: "Spec with message using reserved tags",
			fields: fields{
				Package: "foo",
				Messages: []Message{
					Message{
						Name: "Beacon",
						Messages: []Message{
							Message{
								Name: "Event",
								ReservedValues: []Reserved{
									ReservedTagValue{Tag: 1},
									ReservedTagValue{Tag: 2},
									ReservedTagValue{Tag: 3},
									ReservedTagRange{LowerTag: 6, UpperTag: 9},
								},
								Fields: []Field{
									CustomField{Name: "Habitat", Typing: "string", Tag: 10, Rule: REPEATED, Comment: "What am I?"},
									ScalarField{Name: "Continent", Typing: STRING_TYPE, Tag: 11, Comment: "Where am I?"},
									MapField{Name: "LanguageMap", KeyTyping: STRING_TYPE, ValueTyping: STRING_TYPE, Tag: 12, Comment: "Super essential"},
								},
							},
						},
						ReservedValues: []Reserved{
							ReservedTagValue{Tag: 1},
							ReservedTagValue{Tag: 2},
							ReservedTagValue{Tag: 3},
							ReservedTagRange{LowerTag: 6, UpperTag: 9},
						},
						Fields: []Field{
							CustomField{Name: "Habitat", Typing: "string", Tag: 20, Comment: "What am I?"},
							ScalarField{Name: "Continent", Typing: STRING_TYPE, Tag: 21, Rule: REPEATED, Comment: "Where am I?"},
							MapField{Name: "LanguageMap", KeyTyping: STRING_TYPE, ValueTyping: STRING_TYPE, Tag: 22, Comment: "Super essential"},
							CustomMapField{Name: "CustomMap", KeyTyping: STRING_TYPE, ValueTyping: "Event", Tag: 23},
						},
						OneOfs: []OneOf{
							OneOf{
								Name: "test_oneof",
								Fields: []Field{
									ScalarField{Name: "name", Typing: STRING_TYPE, Tag: 24, Comment: "Name"},
									CustomField{Name: "sub_message", Typing: "Event", Tag: 25, Comment: "Sub-Message"},
								},
							},
						},
						Enums: []Enum{
							Enum{
								Name: "Country",
								Values: []EnumValue{
									EnumValue{Name: "US", Tag: 0},
									EnumValue{Name: "CA", Tag: 1, Comment: "Canada"},
									EnumValue{Name: "GB", Tag: 2, Comment: "Great Britain"},
									EnumValue{Name: "MX", Tag: 3, Comment: "Mexico"},
								},
							},
							Enum{
								Name:       "PlaybackState",
								AllowAlias: true,
								Values: []EnumValue{
									EnumValue{Name: "Waiting", Tag: 0},
									EnumValue{Name: "Playing", Tag: 1},
									EnumValue{Name: "Started", Tag: 1},
									EnumValue{Name: "Stopped", Tag: 2},
								},
							},
						},
					},
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		s := &Spec{
			Package:  tt.fields.Package,
			Imports:  tt.fields.Imports,
			Messages: tt.fields.Messages,
		}
		got, err := s.Write()
		if (err != nil) != tt.wantErr {
			t.Errorf("%q. Spec.Write() error = %v, wantErr %v", tt.name, err, tt.wantErr)
			continue
		}
		if tt.wantErr == false && got == "" {
			t.Error("String representing Protobuf spec was unexpectedly empty")
			continue
		}
	}
}

func ExampleSpec_Write() {
	spec := &Spec{
		Package: "foo",
		Messages: []Message{
			Message{
				Name: "Beacon",
				Messages: []Message{
					Message{
						Name: "Event",
						ReservedValues: []Reserved{
							ReservedTagValue{Tag: 1},
							ReservedTagValue{Tag: 2},
							ReservedTagValue{Tag: 3},
							ReservedTagRange{LowerTag: 6, UpperTag: 9},
						},
						Fields: []Field{
							CustomField{Name: "Habitat", Typing: "string", Tag: 10, Rule: REPEATED, Comment: "What am I?"},
							ScalarField{Name: "Continent", Typing: STRING_TYPE, Tag: 11, Comment: "Where am I?"},
							MapField{Name: "LanguageMap", KeyTyping: STRING_TYPE, ValueTyping: STRING_TYPE, Tag: 12, Comment: "Super essential"},
						},
					},
				},
				ReservedValues: []Reserved{
					ReservedTagValue{Tag: 1},
					ReservedTagValue{Tag: 2},
					ReservedTagValue{Tag: 3},
					ReservedTagRange{LowerTag: 6, UpperTag: 9},
				},
				Fields: []Field{
					CustomField{Name: "Habitat", Typing: "string", Tag: 20, Comment: "What am I?"},
					ScalarField{Name: "Continent", Typing: STRING_TYPE, Tag: 21, Rule: REPEATED, Comment: "Where am I?"},
					MapField{Name: "LanguageMap", KeyTyping: STRING_TYPE, ValueTyping: STRING_TYPE, Tag: 22, Comment: "Super essential"},
					CustomMapField{Name: "CustomMap", KeyTyping: STRING_TYPE, ValueTyping: "Event", Tag: 23},
				},
				OneOfs: []OneOf{
					OneOf{
						Name: "test_oneof",
						Fields: []Field{
							ScalarField{Name: "name", Typing: STRING_TYPE, Tag: 24, Comment: "Name"},
							CustomField{Name: "sub_message", Typing: "Event", Tag: 25, Comment: "Sub-Message"},
						},
					},
				},
				Enums: []Enum{
					Enum{
						Name: "Country",
						Values: []EnumValue{
							EnumValue{Name: "US", Tag: 0},
							EnumValue{Name: "CA", Tag: 1, Comment: "Canada"},
							EnumValue{Name: "GB", Tag: 2, Comment: "Great Britain"},
							EnumValue{Name: "MX", Tag: 3, Comment: "Mexico"},
						},
					},
					Enum{
						Name:       "PlaybackState",
						AllowAlias: true,
						Values: []EnumValue{
							EnumValue{Name: "Waiting", Tag: 0},
							EnumValue{Name: "Playing", Tag: 1},
							EnumValue{Name: "Started", Tag: 1},
							EnumValue{Name: "Stopped", Tag: 2},
						},
					},
				},
			},
		},
	}

	s, err := spec.Write()
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println(s)
	}

	// Output:
	// syntax = "proto3";
	// package foo;
	//
	// message Beacon {
	//   message Event {
	//     reserved 1;
	//     reserved 2;
	//     reserved 3;
	//     reserved 6 to 9;
	//     repeated string Habitat = 10;   // What am I?
	//     string Continent = 11;   // Where am I?
	//     map<string, string> LanguageMap = 12;   // Super essential
	//   }
	//
	//   enum Country {
	//     US = 0;
	//     CA = 1;   // Canada
	//     GB = 2;   // Great Britain
	//     MX = 3;   // Mexico
	//   }
	//   enum PlaybackState {
	//     option allow_alias = true;
	//     Waiting = 0;
	//     Playing = 1;
	//     Started = 1;
	//     Stopped = 2;
	//   }
	//   reserved 1;
	//   reserved 2;
	//   reserved 3;
	//   reserved 6 to 9;
	//   string Habitat = 20;   // What am I?
	//   repeated string Continent = 21;   // Where am I?
	//   map<string, string> LanguageMap = 22;   // Super essential
	//   map<string, Event> CustomMap = 23;
	//   oneof test_oneof {
	//     string name = 24;   // Name
	//     Event sub_message = 25;   // Sub-Message
	//   }
	// }
}
