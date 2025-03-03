package gql

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/graphql-go/graphql"
)

type SchemaBuilder struct {
	query        map[string]interface{}
	mutation     map[string]interface{}
	subscription map[string]interface{}
	typeRegistry map[reflect.Type]graphql.Output
}

func NewSchemaBuilder() *SchemaBuilder {
	return &SchemaBuilder{
		typeRegistry: make(map[reflect.Type]graphql.Output),
	}
}

func (b *SchemaBuilder) WithQuery(query map[string]interface{}) *SchemaBuilder {
	b.query = query
	return b
}

func (b *SchemaBuilder) WithMutation(mutation map[string]interface{}) *SchemaBuilder {
	b.mutation = mutation
	return b
}

func (b *SchemaBuilder) WithSubscription(subscription map[string]interface{}) *SchemaBuilder {
	b.subscription = subscription
	return b
}

func (b *SchemaBuilder) BuildSchemaConfig() (*graphql.SchemaConfig, error) {

	var queryObject *graphql.Object
	var err error
	if b.query != nil {
		queryObject, err = b.MapAsGraphqlObject("Query", b.query)
		if err != nil {
			return nil, fmt.Errorf("failed to build query type: %w", err)
		}
	}

	var mutationObject *graphql.Object
	if b.mutation != nil {
		mutationObject, err = b.MapAsGraphqlObject("Mutation", b.mutation)
		if err != nil {
			return nil, fmt.Errorf("failed to build mutation type: %w", err)
		}
	}

	var subscriptionObject *graphql.Object
	if b.subscription != nil {
		subscriptionObject, err = b.MapAsGraphqlObject("Subscription", b.subscription)
		if err != nil {
			return nil, fmt.Errorf("failed to build subscription type: %w", err)
		}
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

func (b *SchemaBuilder) MapAsGraphqlObject(name string, iface map[string]interface{}) (*graphql.Object, error) {
	fields := graphql.Fields{}
	for fieldName, fn := range iface {
		// get fn as a reflect.Func
		fnValue := reflect.ValueOf(fn)
		if fnValue.Kind() != reflect.Func {
			return nil, fmt.Errorf("field %s is not a method", fieldName)
		}
		resolveInfo, err := NewResolveInfo(fnValue, false)
		if err != nil {
			return nil, err
		}

		if resolveInfo.Output == nil {
			return nil, fmt.Errorf("unbound resolvers should have an output type, %s", fieldName)
		}

		graphqlField, err := b.ReflectTypeAsGraphqlField(resolveInfo.Output.Type)
		if err != nil {
			return nil, err
		}

		graphqlField.Resolve = resolveInfo.Resolve
		if resolveInfo.Input != nil {
			err := b.populateGraphqlFieldArgs(graphqlField, resolveInfo.Input.Type)
			if err != nil {
				return nil, err
			}
		}
		fields[fieldName] = graphqlField
	}
	return graphql.NewObject(graphql.ObjectConfig{Name: name, Fields: fields}), nil
}

func (b *SchemaBuilder) ReflectTypeAsGraphqlField(definition reflect.Type) (*graphql.Field, error) {
	switch definition.Kind() {
	case reflect.Int:
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
	case reflect.Float64:
		return &graphql.Field{
			Type: graphql.Float,
		}, nil
	case reflect.Slice:
		elemField, err := b.ReflectTypeAsGraphqlField(definition.Elem())
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
				return b.ReflectTypeAsGraphqlField(realDefinition)
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

			graphqlField, err := b.ReflectTypeAsGraphqlField(field.Type)
			if err != nil {
				return nil, err
			}

			graphqlField.Name = fieldName

			if isNonNull {
				graphqlField.Type = graphql.NewNonNull(graphqlField.Type)
			}

			resolveMethodName := ResolvePrefix + strings.Title(field.Name)
			method, ok := definition.MethodByName(resolveMethodName)
			if ok {
				resolveInfo, err := NewResolveInfo(method.Func, true)
				if err != nil {
					return nil, err
				}
				graphqlField.Resolve = resolveInfo.Resolve

				if resolveInfo.Output != nil {
					if resolveInfo.Output.Type != field.Type {
						return nil, fmt.Errorf("output type %s does not match field type %s", resolveInfo.Output.Type, field.Type)
					}
				}

				if resolveInfo.Input != nil {
					if resolveInfo.Input != nil {
						err := b.populateGraphqlFieldArgs(graphqlField, resolveInfo.Input.Type)
						if err != nil {
							return nil, err
						}
					}
				}
			}

			fields[fieldName] = graphqlField
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

const (
	ResolvePrefix = "Resolve"
)

func (b *SchemaBuilder) ReflectTypeAsGraphqlArgumentConfig(definition reflect.Type) (*graphql.ArgumentConfig, error) {
	switch definition.Kind() {
	case reflect.Int:
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
	case reflect.Float64:
		return &graphql.ArgumentConfig{
			Type: graphql.Float,
		}, nil
	case reflect.Slice:
		elemConfig, err := b.ReflectTypeAsGraphqlArgumentConfig(definition.Elem())
		if err != nil {
			return nil, err
		}
		return &graphql.ArgumentConfig{
			Type: graphql.NewList(elemConfig.Type),
		}, nil
	case reflect.Ptr:
		return b.ReflectTypeAsGraphqlArgumentConfig(definition.Elem())
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

			fieldConfig, err := b.ReflectTypeAsGraphqlArgumentConfig(field.Type)
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
	argumentConfig, err := b.ReflectTypeAsGraphqlArgumentConfig(definition)
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
