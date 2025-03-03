package gql

import (
	"reflect"
	"testing"
)

func TestParseGqlTag(t *testing.T) {

	cases := []struct {
		tag               string
		expectedFieldName string
		expectedNonNull   bool
		expectedError     bool
	}{
		{"", "", false, false},
		{"name", "name", false, false},
		{"name,nonNull", "name", true, false},
		{"name,foo", "name", true, true},
		{"name,nonNull,foo", "name", true, true},
	}
	for _, c := range cases {
		t.Run(c.tag, func(t *testing.T) {
			gqlTag, err := ParseGqlTag(c.tag)
			if err != nil != c.expectedError {
				t.Fatalf("expected error to be %t, got %t", c.expectedError, err != nil)
			}

			if err != nil {
				return
			}

			if gqlTag.GetFieldName() != c.expectedFieldName {
				t.Fatalf("expected field name to be %s, got %s", c.expectedFieldName, gqlTag.GetFieldName())
			}

			if gqlTag.IsNonNull() != c.expectedNonNull {
				t.Fatalf("expected nonNull to be %t, got %t", c.expectedNonNull, gqlTag.IsNonNull())
			}
		})
	}
}

func TestParseGqlTagFromField(t *testing.T) {
	cases := []struct {
		field             *reflect.StructField
		expectedFieldName string
		expectedNonNull   bool
		expectedError     bool
	}{
		{
			field: &reflect.StructField{
				Name: "Name",
				Tag:  reflect.StructTag(`gql:""`),
			},
			expectedFieldName: "",
			expectedNonNull:   false,
			expectedError:     false,
		},
		{
			field: &reflect.StructField{
				Name: "Name",
				Tag:  reflect.StructTag(`gql:"name"`),
			},
			expectedFieldName: "name",
			expectedNonNull:   false,
			expectedError:     false,
		},
		{
			field: &reflect.StructField{
				Name: "Name",
				Tag:  reflect.StructTag(`gql:"name,"`),
			},
			expectedFieldName: "name",
			expectedNonNull:   false,
			expectedError:     true,
		},
		{
			field: &reflect.StructField{
				Name: "Name",
				Tag:  reflect.StructTag(`gql:"name,nonNull"`),
			},
			expectedFieldName: "name",
			expectedNonNull:   true,
			expectedError:     false,
		},
		{
			field: &reflect.StructField{
				Name: "Name",
				Tag:  reflect.StructTag(`json:"name,nonNull"`),
			},
			expectedFieldName: "",
			expectedNonNull:   false,
			expectedError:     false,
		},
		{
			field: &reflect.StructField{
				Name: "Name",
				Tag:  reflect.StructTag(`gql:"name,foo"`),
			},
			expectedFieldName: "",
			expectedNonNull:   false,
			expectedError:     true,
		},
	}

	for _, c := range cases {
		t.Run(c.field.Name, func(t *testing.T) {
			gqlTag, err := ParseGqlTagFromField(c.field)
			if err != nil != c.expectedError {
				t.Fatalf("expected error to be %t, got %t", c.expectedError, err != nil)
			}

			if err != nil {
				return
			}

			if gqlTag.GetFieldName() != c.expectedFieldName {
				t.Fatalf("expected field name to be %s, got %s", c.expectedFieldName, gqlTag.GetFieldName())
			}

			if gqlTag.IsNonNull() != c.expectedNonNull {
				t.Fatalf("expected nonNull to be %t, got %t", c.expectedNonNull, gqlTag.IsNonNull())
			}
		})
	}
}

func TestGetGqlTag(t *testing.T) {
	cases := []struct {
		field             *reflect.StructField
		expectedFieldName string
		expectedNonNull   bool
		expectedError     bool
	}{
		{
			field: &reflect.StructField{
				Name: "Name",
				Tag:  reflect.StructTag(`gql:"name,nonNull"`),
			},
			expectedFieldName: "name",
			expectedNonNull:   true,
			expectedError:     false,
		},
		{
			field: &reflect.StructField{
				Name: "Name",
				Tag:  reflect.StructTag(`gql:"name"`),
			},
			expectedFieldName: "name",
			expectedNonNull:   false,
			expectedError:     false,
		},
		{
			field: &reflect.StructField{
				Name: "Name",
				Tag:  reflect.StructTag(`gql:"name,foo"`),
			},
			expectedFieldName: "",
			expectedNonNull:   false,
			expectedError:     true,
		},
	}

	for _, c := range cases {
		t.Run(c.field.Name, func(t *testing.T) {
			fieldName, nonNull, err := GetGqlTag(c.field)
			if err != nil != c.expectedError {
				t.Fatalf("expected error to be %t, got %t", c.expectedError, err != nil)
			}

			if err != nil {
				return
			}

			if fieldName != c.expectedFieldName {
				t.Fatalf("expected field name to be %s, got %s", c.expectedFieldName, fieldName)
			}

			if nonNull != c.expectedNonNull {
				t.Fatalf("expected nonNull to be %t, got %t", c.expectedNonNull, nonNull)
			}
		})
	}
}
