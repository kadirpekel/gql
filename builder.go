package gql

import (
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/graphql-go/graphql"
	"github.com/graphql-go/graphql/language/ast"
)

type RootType string

const (
	Query        RootType = "Query"
	Mutation     RootType = "Mutation"
	Subscription RootType = "Subscription"
)

type SchemaBuilder struct {
	query         interface{}
	mutation      interface{}
	subscription  interface{}
	typeRegistry  map[reflect.Type]graphql.Output
	customTypes   map[reflect.Type]graphql.Output
	processing    map[reflect.Type]bool           // Track types currently being processed to prevent cycles
	fieldsCache   map[reflect.Type]graphql.Fields // Cache fields for types being processed
	rootInstances map[reflect.Type]interface{}    // Registry for root instances (Query, Mutation)
}

func NewSchemaBuilder() *SchemaBuilder {
	sb := &SchemaBuilder{
		typeRegistry:  make(map[reflect.Type]graphql.Output),
		customTypes:   make(map[reflect.Type]graphql.Output),
		processing:    make(map[reflect.Type]bool),
		fieldsCache:   make(map[reflect.Type]graphql.Fields),
		rootInstances: make(map[reflect.Type]interface{}),
	}

	// Register default custom types (standard library types only)
	// Framework-specific types (e.g., gorm.DeletedAt) should be registered
	// by the application using RegisterCustomType()
	sb.RegisterCustomType(reflect.TypeOf(time.Time{}), createDateTimeScalar())
	sb.RegisterCustomType(reflect.TypeOf((*time.Time)(nil)).Elem(), createDateTimeScalar())

	return sb
}

// RegisterCustomType registers a custom type mapping
func (b *SchemaBuilder) RegisterCustomType(goType reflect.Type, graphqlType graphql.Output) {
	b.customTypes[goType] = graphqlType
}

// createDateTimeScalar creates a DateTime scalar for time.Time
func createDateTimeScalar() *graphql.Scalar {
	return graphql.NewScalar(graphql.ScalarConfig{
		Name:        "DateTime",
		Description: "DateTime scalar type (RFC3339 format)",
		Serialize: func(value interface{}) interface{} {
			switch v := value.(type) {
			case time.Time:
				return v.Format(time.RFC3339)
			case *time.Time:
				if v == nil {
					return nil
				}
				return v.Format(time.RFC3339)
			default:
				return nil
			}
		},
		ParseValue: func(value interface{}) interface{} {
			switch v := value.(type) {
			case string:
				t, err := time.Parse(time.RFC3339, v)
				if err != nil {
					return nil
				}
				return t
			default:
				return nil
			}
		},
		ParseLiteral: func(valueAST ast.Value) interface{} {
			if strValue, ok := valueAST.(*ast.StringValue); ok {
				t, err := time.Parse(time.RFC3339, strValue.Value)
				if err != nil {
					return nil
				}
				return t
			}
			return nil
		},
	})
}

func (b *SchemaBuilder) WithQuery(query interface{}) *SchemaBuilder {
	b.query = query
	if query != nil {
		t := reflect.TypeOf(query)
		if t.Kind() == reflect.Ptr {
			t = t.Elem()
		}
		b.rootInstances[t] = query
	}
	return b
}

func (b *SchemaBuilder) WithMutation(mutation interface{}) *SchemaBuilder {
	b.mutation = mutation
	if mutation != nil {
		t := reflect.TypeOf(mutation)
		if t.Kind() == reflect.Ptr {
			t = t.Elem()
		}
		b.rootInstances[t] = mutation
	}
	return b
}

func (b *SchemaBuilder) WithSubscription(subscription interface{}) *SchemaBuilder {
	b.subscription = subscription
	if subscription != nil {
		t := reflect.TypeOf(subscription)
		if t.Kind() == reflect.Ptr {
			t = t.Elem()
		}
		b.rootInstances[t] = subscription
	}
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
	// Check for custom type mappings first
	if customType, ok := b.customTypes[definition]; ok {
		return &graphql.Field{
			Type: customType,
		}, nil
	}

	switch definition.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return &graphql.Field{
			Type: graphql.Int,
		}, nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
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
	case reflect.Map:
		// Maps are not directly supported in GraphQL
		// They should be excluded using gql:"-" tag
		// If we reach here, it means a map type was encountered without exclusion
		return nil, fmt.Errorf("map types are not supported in GraphQL schema. Use gql:\"-\" tag to exclude map fields")
	// struct or pointer to struct including slices
	case reflect.Struct, reflect.Ptr:
		realDefinition := definition

		if definition.Kind() == reflect.Ptr {
			realDefinition = definition.Elem()

			if realDefinition.Kind() != reflect.Struct {
				return b.TypeAsGraphqlField(realDefinition)
			}
		}

		// Check if this type is already registered (prevents infinite recursion)
		if existingType, ok := b.typeRegistry[realDefinition]; ok {
			return &graphql.Field{Type: existingType}, nil
		}

		// Check if this type is currently being processed (circular reference)
		if b.processing[realDefinition] {
			// For circular references, return the placeholder that was already created
			// The thunk will resolve fields from the cache when ready
			if existingType, ok := b.typeRegistry[realDefinition]; ok {
				return &graphql.Field{Type: existingType}, nil
			}
			// Create placeholder with thunk that reads from fields cache
			builderRef := b
			typeRef := realDefinition
			placeholder := graphql.NewObject(graphql.ObjectConfig{
				Name: realDefinition.Name(),
				Fields: graphql.FieldsThunk(func() graphql.Fields {
					// Read fields from cache (populated when processing completes)
					if fields, ok := builderRef.fieldsCache[typeRef]; ok {
						return fields
					}
					return graphql.Fields{}
				}),
			})
			b.typeRegistry[realDefinition] = placeholder
			return &graphql.Field{Type: placeholder}, nil
		}

		// Mark as processing
		b.processing[realDefinition] = true
		defer func() {
			delete(b.processing, realDefinition)
		}()

		fields := graphql.Fields{}
		for _, field := range reflect.VisibleFields(realDefinition) {
			fieldName, isNonNull, err := GetGqlTag(&field)
			if err != nil {
				return nil, err
			}

			// if the tag is empty or "-", skip the field, we're interested in fields with a gql tag
			if fieldName == "" || fieldName == "-" {
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
				// Skip methods that don't match resolver signature (e.g., TableName(), BeforeCreate(), etc.)
				resolveInfo, err := NewResolveInfo(method.Func)
				if err != nil {
					// Not a valid resolver, skip it
					continue
				}

				// UPDATE: Check if we have a bound instance for this type
				if instance, ok := b.rootInstances[realDefinition]; ok {
					val := reflect.ValueOf(instance)
					resolveInfo.BoundReceiver = &val
				}

				fieldName := strings.ToLower(method.Name[0:1]) + method.Name[1:]

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

		// Store fields in cache for thunk-based placeholders
		b.fieldsCache[realDefinition] = fields

		// Check if a placeholder was already created (due to circular reference)
		if existingType, ok := b.typeRegistry[realDefinition]; ok {
			// Placeholder exists - return it (its thunk will read from fieldsCache)
			return &graphql.Field{Type: existingType}, nil
		}

		// Check if type has a custom GraphQL type name method
		typeName := realDefinition.Name()
		if method, ok := realDefinition.MethodByName("GraphQLTypeName"); ok {
			if method.Type.NumIn() == 1 && method.Type.NumOut() == 1 {
				// Call the method on a zero value to get the type name
				zeroValue := reflect.New(realDefinition).Elem()
				result := method.Func.Call([]reflect.Value{zeroValue})
				if len(result) > 0 && result[0].Kind() == reflect.String {
					typeName = result[0].String()
				}
			}
		}

		// Create the object with populated fields
		graphqlType := graphql.NewObject(graphql.ObjectConfig{
			Name:   typeName,
			Fields: fields,
		})

		// Register the fully populated object
		b.typeRegistry[realDefinition] = graphqlType

		return &graphql.Field{Type: graphqlType}, nil
	default:
		return nil, fmt.Errorf("Unsupported type: %s", definition.Kind())
	}
}

func (b *SchemaBuilder) TypeAsGraphqlArgumentConfig(definition reflect.Type) (*graphql.ArgumentConfig, error) {
	// Check for custom type mappings first
	if customType, ok := b.customTypes[definition]; ok {
		return &graphql.ArgumentConfig{
			Type: customType,
		}, nil
	}

	switch definition.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return &graphql.ArgumentConfig{
			Type: graphql.Int,
		}, nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
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

			if fieldName == "" || fieldName == "-" {
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
	// Handle pointer types
	if definition.Kind() == reflect.Ptr {
		definition = definition.Elem()
	}

	if definition.Kind() != reflect.Struct {
		return fmt.Errorf("Arguments type must be a struct, got %s", definition.Kind())
	}

	graphqlField.Args = graphql.FieldConfigArgument{}

	// Iterate over struct fields directly
	// This supports both named and anonymous structs
	for i := 0; i < definition.NumField(); i++ {
		field := definition.Field(i)

		fieldName, isNonNull, err := GetGqlTag(&field)
		if err != nil {
			return err
		}

		// Skip fields without valid tags
		if fieldName == "" || fieldName == "-" {
			continue
		}

		// Create argument config for the field
		fieldArgConfig, err := b.TypeAsGraphqlArgumentConfig(field.Type)
		if err != nil {
			return err
		}

		if isNonNull {
			fieldArgConfig.Type = graphql.NewNonNull(fieldArgConfig.Type)
		}

		graphqlField.Args[fieldName] = fieldArgConfig
	}

	return nil
}
