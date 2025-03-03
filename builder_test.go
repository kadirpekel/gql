package gql

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/graphql-go/graphql"
)

type NoTagged struct {
	Field string
}

type Tagged struct {
	Field string `gql:"field"`
}

type TaggedNonNull struct {
	Field string `gql:"field,nonNull"`
}

type WithResolver struct {
	Field string `gql:"field"`
}

type WithSubResolver struct {
	Field string `gql:"field"`
	Sub   *WithResolver
}

func (w *WithSubResolver) ResolveField(ctx context.Context, info graphql.ResolveInfo) string {
	return "parent " + w.Field + " " + w.Sub.Field
}

func (w *WithResolver) ResolveField(ctx context.Context, info graphql.ResolveInfo) string {
	return "parent " + w.Field
}

func NoInputNoOutput() {}

func NoInputWithOutput() string {
	return "foobar"
}

func WithContext(ctx context.Context) string {
	return ctx.Value("ctxKey").(string)
}

func WithContextAndResolveInfo(ctx context.Context, info graphql.ResolveInfo) string {
	return ctx.Value("ctxKey").(string) + " " + info.FieldName
}

func WithContextAndResolveInfoAndNonstructInput(ctx context.Context, info graphql.ResolveInfo, input string) string {
	return ctx.Value("ctxKey").(string) + " " + info.FieldName + " " + input
}

func WithContextAndResolveInfoAndNoTaggedInput(ctx context.Context, info graphql.ResolveInfo, input NoTagged) string {
	return ctx.Value("ctxKey").(string) + " " + info.FieldName + " " + input.Field
}

func WithContextAndResolveInfoAndTaggedInput(ctx context.Context, info graphql.ResolveInfo, input Tagged) string {
	return ctx.Value("ctxKey").(string) + " " + info.FieldName + " " + input.Field
}

func WithContextAndResolveInfoAndTaggedNonNullInput(ctx context.Context, info graphql.ResolveInfo, input TaggedNonNull) string {
	return ctx.Value("ctxKey").(string) + " " + info.FieldName + " " + input.Field
}

func WithContextAndResolveInfoAndTaggedNonNullInputSwapped1(info graphql.ResolveInfo, input TaggedNonNull, ctx context.Context) string {
	return ctx.Value("ctxKey").(string) + " " + info.FieldName + " " + input.Field
}

func WithContextAndResolveInfoAndTaggedNonNullInputSwapped2(input TaggedNonNull, ctx context.Context, info graphql.ResolveInfo) string {
	return ctx.Value("ctxKey").(string) + " " + info.FieldName + " " + input.Field
}

func WithContextAndResolveInfoAndTaggedInputWithNoTaggedOutput(ctx context.Context, info graphql.ResolveInfo, input Tagged) NoTagged {
	return NoTagged{
		Field: ctx.Value("ctxKey").(string) + " " + info.FieldName + " " + input.Field,
	}
}

func WithContextAndResolveInfoAndTaggedInputWithTaggedOutput(ctx context.Context, info graphql.ResolveInfo, input Tagged) Tagged {
	return Tagged{
		Field: ctx.Value("ctxKey").(string) + " " + info.FieldName + " " + input.Field,
	}
}

func WithContextAndResolveInfoAndTaggedInputWithTaggedOutputAndError(ctx context.Context, info graphql.ResolveInfo, input Tagged) (Tagged, error) {
	return Tagged{}, errors.New("error")
}

func WithContextAndResolveInfoAndTaggedInputWithTaggedOutputAllPtr(ctx context.Context, info *graphql.ResolveInfo, input *Tagged) *Tagged {
	return &Tagged{
		Field: ctx.Value("ctxKey").(string) + " " + info.FieldName + " " + input.Field,
	}
}

func WithNestedResolvers() *WithSubResolver {
	return &WithSubResolver{
		Field: "foobar",
		Sub:   &WithResolver{Field: "subfoobar"},
	}
}

func TestResolve(t *testing.T) {

	type Case struct {
		field         string
		fn            interface{}
		query         string
		expected      interface{}
		isSchemaError bool
		isError       bool
	}

	cases := []Case{
		{
			field:         "noInputNoOutput",
			fn:            NoInputNoOutput,
			query:         `{ noInputNoOutput }`,
			expected:      ``,
			isSchemaError: true,
			isError:       true,
		},
		{
			field:    "noInputWithOutput",
			fn:       NoInputWithOutput,
			query:    `{ noInputWithOutput }`,
			expected: `foobar`,
		},
		{
			field:    "withContext",
			fn:       WithContext,
			query:    `{ withContext }`,
			expected: `ctxValue`,
		},
		{
			field:    "withContextAndResolveInfo",
			fn:       WithContextAndResolveInfo,
			query:    `{ withContextAndResolveInfo }`,
			expected: `ctxValue withContextAndResolveInfo`,
		},
		{
			field:         "withContextAndResolveInfoAndNonstructInput",
			fn:            WithContextAndResolveInfoAndNonstructInput,
			query:         `{ withContextAndResolveInfoAndNonstructInput(input: "input") }`,
			expected:      `ctxValue withContextAndResolveInfoAndNonstructInput input`,
			isSchemaError: true,
		},
		{
			field:         "withContextAndResolveInfoAndNoTaggedInput",
			fn:            WithContextAndResolveInfoAndNoTaggedInput,
			query:         `{ withContextAndResolveInfoAndNoTaggedInput( field: "foobar" ) }`,
			expected:      `ctxValue withContextAndResolveInfoAndNoTaggedInput foobar`,
			isSchemaError: true,
		},
		{
			field:    "withContextAndResolveInfoAndTaggedInput",
			fn:       WithContextAndResolveInfoAndTaggedInput,
			query:    `{ withContextAndResolveInfoAndTaggedInput( field: "foobar" ) }`,
			expected: `ctxValue withContextAndResolveInfoAndTaggedInput foobar`,
		},
		{
			field:    "withContextAndResolveInfoAndTaggedInput",
			fn:       WithContextAndResolveInfoAndTaggedInput,
			query:    `{ withContextAndResolveInfoAndTaggedInput }`,
			expected: `ctxValue withContextAndResolveInfoAndTaggedInput `,
		},
		{
			field:    "withContextAndResolveInfoAndTaggedNonNullInput",
			fn:       WithContextAndResolveInfoAndTaggedNonNullInput,
			query:    `{ withContextAndResolveInfoAndTaggedNonNullInput }`,
			expected: `ctxValue withContextAndResolveInfoAndTaggedNonNullInput foobar`,
			isError:  true,
		},
		{
			field:    "withContextAndResolveInfoAndTaggedNonNullInput",
			fn:       WithContextAndResolveInfoAndTaggedNonNullInput,
			query:    `{ withContextAndResolveInfoAndTaggedNonNullInput( field: "foobar" ) }`,
			expected: `ctxValue withContextAndResolveInfoAndTaggedNonNullInput foobar`,
		},
		{
			field:    "withContextAndResolveInfoAndTaggedNonNullInputSwapped1",
			fn:       WithContextAndResolveInfoAndTaggedNonNullInputSwapped1,
			query:    `{ withContextAndResolveInfoAndTaggedNonNullInputSwapped1( field: "foobar" ) }`,
			expected: `ctxValue withContextAndResolveInfoAndTaggedNonNullInputSwapped1 foobar`,
		},
		{
			field:    "withContextAndResolveInfoAndTaggedNonNullInputSwapped2",
			fn:       WithContextAndResolveInfoAndTaggedNonNullInputSwapped2,
			query:    `{ withContextAndResolveInfoAndTaggedNonNullInputSwapped2( field: "foobar" ) }`,
			expected: `ctxValue withContextAndResolveInfoAndTaggedNonNullInputSwapped2 foobar`,
		},
		{
			field:         "withContextAndResolveInfoAndTaggedInputWithNoTaggedOutput",
			fn:            WithContextAndResolveInfoAndTaggedInputWithNoTaggedOutput,
			query:         `{ withContextAndResolveInfoAndTaggedInputWithNoTaggedOutput( field: "foobar" ) }`,
			expected:      `ctxValue withContextAndResolveInfoAndTaggedInputWithNoTaggedOutput foobar`,
			isSchemaError: true,
		},
		{
			field:    "withContextAndResolveInfoAndTaggedInputWithTaggedOutput",
			fn:       WithContextAndResolveInfoAndTaggedInputWithTaggedOutput,
			query:    `{ withContextAndResolveInfoAndTaggedInputWithTaggedOutput( field: "foobar" ) { field } }`,
			expected: map[string]interface{}{"field": "ctxValue withContextAndResolveInfoAndTaggedInputWithTaggedOutput foobar"},
		},
		{
			field:    "withContextAndResolveInfoAndTaggedInputWithTaggedOutputAndError",
			fn:       WithContextAndResolveInfoAndTaggedInputWithTaggedOutputAndError,
			query:    `{ withContextAndResolveInfoAndTaggedInputWithTaggedOutputAndError( field: "foobar" ) { field } }`,
			expected: nil,
			isError:  true,
		},
		{
			field:    "withContextAndResolveInfoAndTaggedInputWithTaggedOutputAllPtr",
			fn:       WithContextAndResolveInfoAndTaggedInputWithTaggedOutputAllPtr,
			query:    `{ withContextAndResolveInfoAndTaggedInputWithTaggedOutputAllPtr( field: "foobar" ) { field } }`,
			expected: map[string]interface{}{"field": "ctxValue withContextAndResolveInfoAndTaggedInputWithTaggedOutputAllPtr foobar"},
		},
		{
			field:    "withNestedResolvers",
			fn:       WithNestedResolvers,
			query:    `{ withNestedResolvers { field } }`,
			expected: map[string]interface{}{"field": "parent foobar subfoobar"},
		},
	}

	grandSchemaMap := map[string]interface{}{}

	validateCase := func(t *testing.T, schema *graphql.Schema, c *Case) {

		result := graphql.Do(graphql.Params{
			Schema:        *schema,
			RequestString: c.query,
			Context:       context.WithValue(context.Background(), "ctxKey", "ctxValue"),
		})

		if c.isError {
			if result.Errors == nil {
				t.Errorf("expected errors, got %v", result.Errors)
			}
		} else {
			if result.Errors != nil {
				t.Errorf("expected no errors, got %v", result.Errors)
			}

			data, ok := result.Data.(map[string]interface{})
			if !ok {
				t.Errorf("expected data to be a map[string]interface{}, got %T", result.Data)
			}

			if !reflect.DeepEqual(data[c.field], c.expected) {
				t.Errorf("expected %v, got %v", c.expected, data[c.field])
			}
		}
	}

	for _, c := range cases {

		schema, err := NewSchemaBuilder().WithQuery(map[string]interface{}{
			c.field: c.fn,
		}).BuildSchema()

		if c.isSchemaError {
			if err == nil {
				t.Errorf("expected schema error, got nil")
			}
			continue
		} else {
			if err != nil {
				t.Errorf("expected no error, got %v", err)
			}
			grandSchemaMap[c.field] = c.fn
		}

		validateCase(t, schema, &c)
	}

	grandSchema, err := NewSchemaBuilder().WithQuery(grandSchemaMap).BuildSchema()
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	for _, c := range cases {
		if c.isSchemaError {
			continue
		}

		validateCase(t, grandSchema, &c)
	}
}
