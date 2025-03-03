package gql

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/graphql-go/graphql"
)

type RootType string

const (
	Query        RootType = "Query"
	Mutation     RootType = "Mutation"
	Subscription RootType = "Subscription"
)

type SchemaBuilder struct {
	query        interface{}
	mutation     interface{}
	subscription interface{}
	typeRegistry map[reflect.Type]graphql.Output
}

func NewSchemaBuilder() *SchemaBuilder {
	return &SchemaBuilder{
		typeRegistry: make(map[reflect.Type]graphql.Output),
	}
}

func (b *SchemaBuilder) WithQuery(query interface{}) *SchemaBuilder {
	b.query = query
	return b
}

func (b *SchemaBuilder) WithMutation(mutation interface{}) *SchemaBuilder {
	b.mutation = mutation
	return b
}

func (b *SchemaBuilder) WithSubscription(subscription interface{}) *SchemaBuilder {
	b.subscription = subscription
	return b
}

func (b *SchemaBuilder) BuildSchemaConfig() (*graphql.SchemaConfig, error) {

	var queryObject, mutationObject, subscriptionObject *graphql.Object

	if b.query != nil {
		graphqlField, err := b.TypeAsGraphqlField(reflect.TypeOf(b.query))
		if err != nil {
			return nil, fmt.Errorf("failed to build query type: %w", err)
		}
		queryObject = graphqlField.Type.(*graphql.Object)
	}

	if b.mutation != nil {
		graphqlField, err := b.TypeAsGraphqlField(reflect.TypeOf(b.mutation))
		if err != nil {
			return nil, fmt.Errorf("failed to build mutation type: %w", err)
		}
		mutationObject = graphqlField.Type.(*graphql.Object)
	}

	if b.subscription != nil {
		graphqlField, err := b.TypeAsGraphqlField(reflect.TypeOf(b.subscription))
		if err != nil {
			return nil, fmt.Errorf("failed to build subscription type: %w", err)
		}
		subscriptionObject = graphqlField.Type.(*graphql.Object)
	}

	return &graphql.SchemaConfig{
		Query:        queryObject,
		Mutation:     mutationObject,
		Subscription: subscriptionObject,
	}, nil
}

func (b *SchemaBuilder) BuildSchema() (*graphql.Schema, error) {
	schemaConfig, err := b.BuildSchemaConfig()
	if err != nil {
		return nil, err
	}
	schema, err := graphql.NewSchema(*schemaConfig)
	if err != nil {
		return nil, err
	}
	return &schema, nil
}

func (b *SchemaBuilder) TypeAsGraphqlField(definition reflect.Type) (*graphql.Field, error) {
	switch definition.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return &graphql.Field{
			Type: graphql.Int,
		}, nil
	case reflect.String:
		return &graphql.Field{
			Type: graphql.String,
		}, nil
	case reflect.Bool:
		return &graphql.Field{
			Type: graphql.Boolean,
		}, nil
	case reflect.Float64, reflect.Float32:
		return &graphql.Field{
			Type: graphql.Float,
		}, nil
	case reflect.Slice, reflect.Array:
		elemField, err := b.TypeAsGraphqlField(definition.Elem())
		if err != nil {
			return nil, err
		}
		return &graphql.Field{
			Type: graphql.NewList(elemField.Type),
		}, nil
	// struct or pointer to struct including slices
	case reflect.Struct, reflect.Ptr:
		realDefinition := definition

		if definition.Kind() == reflect.Ptr {
			realDefinition = definition.Elem()

			if realDefinition.Kind() != reflect.Struct {
				return b.TypeAsGraphqlField(realDefinition)
			}
		}

		fields := graphql.Fields{}
		for _, field := range reflect.VisibleFields(realDefinition) {
			fieldName, isNonNull, err := GetGqlTag(&field)
			if err != nil {
				return nil, err
			}

			// if the tag is empty, skip the field, we're interested in fields with a gql tag
			if fieldName == "" {
				continue
			}

			graphqlField, err := b.TypeAsGraphqlField(field.Type)
			if err != nil {
				return nil, err
			}

			graphqlField.Name = fieldName

			if isNonNull {
				graphqlField.Type = graphql.NewNonNull(graphqlField.Type)
			}

			fields[fieldName] = graphqlField
		}

		for i := 0; i < definition.NumMethod(); i++ {
			method := definition.Method(i)
			if method.IsExported() {
				fieldName := strings.ToLower(method.Name[0:1]) + method.Name[1:]

				resolveInfo, err := NewResolveInfo(method.Func)
				if err != nil {
					return nil, err
				}

				graphqlField, err := b.TypeAsGraphqlField(resolveInfo.Output.Type)
				if err != nil {
					return nil, err
				}

				graphqlField.Name = fieldName
				graphqlField.Resolve = resolveInfo.Resolve
				if resolveInfo.Input != nil {
					err := b.populateGraphqlFieldArgs(graphqlField, resolveInfo.Input.Type)
					if err != nil {
						return nil, err
					}
				}
				fields[fieldName] = graphqlField
			}
		}

		graphqlType, ok := b.typeRegistry[realDefinition]
		if !ok {
			graphqlType = graphql.NewObject(graphql.ObjectConfig{Name: realDefinition.Name(), Fields: fields})
			b.typeRegistry[realDefinition] = graphqlType
		}

		return &graphql.Field{Type: graphqlType}, nil
	default:
		return nil, fmt.Errorf("Unsupported type: %s", definition.Kind())
	}
}

func (b *SchemaBuilder) TypeAsGraphqlArgumentConfig(definition reflect.Type) (*graphql.ArgumentConfig, error) {
	switch definition.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return &graphql.ArgumentConfig{
			Type: graphql.Int,
		}, nil
	case reflect.String:
		return &graphql.ArgumentConfig{
			Type: graphql.String,
		}, nil
	case reflect.Bool:
		return &graphql.ArgumentConfig{
			Type: graphql.Boolean,
		}, nil
	case reflect.Float64, reflect.Float32:
		return &graphql.ArgumentConfig{
			Type: graphql.Float,
		}, nil
	case reflect.Slice, reflect.Array:
		elemConfig, err := b.TypeAsGraphqlArgumentConfig(definition.Elem())
		if err != nil {
			return nil, err
		}
		return &graphql.ArgumentConfig{
			Type: graphql.NewList(elemConfig.Type),
		}, nil
	case reflect.Ptr:
		return b.TypeAsGraphqlArgumentConfig(definition.Elem())
	case reflect.Struct:
		fields := graphql.InputObjectConfigFieldMap{}
		for i := 0; i < definition.NumField(); i++ {
			field := definition.Field(i)
			fieldName, isNonNull, err := GetGqlTag(&field)
			if err != nil {
				return nil, err
			}

			if fieldName == "" {
				continue
			}

			fieldConfig, err := b.TypeAsGraphqlArgumentConfig(field.Type)
			if err != nil {
				return nil, err
			}

			if isNonNull {
				fieldConfig.Type = graphql.NewNonNull(fieldConfig.Type)
			}

			fields[fieldName] = &graphql.InputObjectFieldConfig{
				Type: fieldConfig.Type,
			}
		}
		return &graphql.ArgumentConfig{
			Type: graphql.NewInputObject(graphql.InputObjectConfig{Name: definition.Name(), Fields: fields}),
		}, nil
	default:
		return nil, fmt.Errorf("Unsupported type: %s", definition.Kind())
	}
}

func (b *SchemaBuilder) populateGraphqlFieldArgs(graphqlField *graphql.Field, definition reflect.Type) error {
	argumentConfig, err := b.TypeAsGraphqlArgumentConfig(definition)
	if err != nil {
		return err
	}
	argFields := argumentConfig.Type.(*graphql.InputObject).Fields()
	graphqlField.Args = graphql.FieldConfigArgument{}
	for fieldName, argField := range argFields {
		graphqlField.Args[fieldName] = &graphql.ArgumentConfig{
			Type: argField.Type,
		}
	}
	return nil
}
