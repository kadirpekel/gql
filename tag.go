package gql

import (
	"fmt"
	"reflect"
	"strings"
)

const (
	GqlTagKey = "gql"
)

type GqlTag struct {
	FieldName string
	NonNull   bool
}

func (t *GqlTag) IsNonNull() bool {
	return t.NonNull
}

func (t *GqlTag) GetFieldName() string {
	return t.FieldName
}

func ParseGqlTag(tag string) (*GqlTag, error) {
	t := &GqlTag{}

	parts := strings.Split(tag, ",")
	if len(parts) > 2 {
		return nil, fmt.Errorf("Invalid gql tag expected fieldName, got: %s", tag)
	}

	t.FieldName = parts[0]
	if len(parts) == 2 {
		if parts[1] == "nonNull" {
			t.NonNull = true
		} else {
			return nil, fmt.Errorf("Invalid gql tag expected nonNull, got: %s", parts[1])
		}
	}

	return t, nil
}

func ParseGqlTagFromField(field *reflect.StructField) (*GqlTag, error) {
	tag := field.Tag.Get(GqlTagKey)
	return ParseGqlTag(tag)
}

func GetGqlTag(field *reflect.StructField) (string, bool, error) {
	gqlTag, err := ParseGqlTagFromField(field)
	if err != nil {
		return "", false, err
	}

	return gqlTag.GetFieldName(), gqlTag.IsNonNull(), nil
}
