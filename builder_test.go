package gql

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/graphql-go/graphql"
)

type Tagged struct {
	Field string `gql:"field"`
}

type TaggedNonNull struct {
	Field string `gql:"field,nonNull"`
}

type Host struct{}

func (h *Host) ResolveField(ctx context.Context, info graphql.ResolveInfo) (string, error) {
	return "host", nil
}

func (h *Host) WithContext(ctx context.Context) (string, error) {
	return ctx.Value("ctxKey").(string), nil
}

func (h *Host) WithContextAndResolveInfo(ctx context.Context, info graphql.ResolveInfo) (string, error) {
	return ctx.Value("ctxKey").(string) + " " + info.FieldName, nil
}

func (h *Host) WithContextAndResolveInfoAndTaggedInput(ctx context.Context, info graphql.ResolveInfo, input Tagged) (string, error) {
	return ctx.Value("ctxKey").(string) + " " + info.FieldName + " " + input.Field, nil
}

func (h *Host) WithContextAndResolveInfoAndTaggedNonNullInput(ctx context.Context, info graphql.ResolveInfo, input TaggedNonNull) (string, error) {
	return ctx.Value("ctxKey").(string) + " " + info.FieldName + " " + input.Field, nil
}

func (h *Host) WithContextAndResolveInfoAndTaggedNonNullInputSwapped1(info graphql.ResolveInfo, input TaggedNonNull, ctx context.Context) (string, error) {
	return ctx.Value("ctxKey").(string) + " " + info.FieldName + " " + input.Field, nil
}

func (h *Host) WithContextAndResolveInfoAndTaggedNonNullInputSwapped2(input TaggedNonNull, ctx context.Context, info graphql.ResolveInfo) (string, error) {
	return ctx.Value("ctxKey").(string) + " " + info.FieldName + " " + input.Field, nil
}

func (h *Host) WithContextAndResolveInfoAndTaggedInputWithTaggedOutput(ctx context.Context, info graphql.ResolveInfo, input Tagged) (Tagged, error) {
	return Tagged{
		Field: ctx.Value("ctxKey").(string) + " " + info.FieldName + " " + input.Field,
	}, nil
}

func (h *Host) WithContextAndResolveInfoAndTaggedInputWithTaggedOutputAndError(ctx context.Context, info graphql.ResolveInfo, input Tagged) (Tagged, error) {
	return Tagged{}, errors.New("error")
}

func (h *Host) WithContextAndResolveInfoAndTaggedInputWithTaggedOutputAllPtr(ctx context.Context, info *graphql.ResolveInfo, input *Tagged) (*Tagged, error) {
	return &Tagged{
		Field: ctx.Value("ctxKey").(string) + " " + info.FieldName + " " + input.Field,
	}, nil
}

func (h *Host) WithNestedResolvers() (*WithResolver, error) {
	return &WithResolver{
		Field: "foobar",
	}, nil
}

type WithResolver struct {
	Field string `gql:"field"`
}

func (w *WithResolver) ResolvedField(ctx context.Context, info graphql.ResolveInfo) (string, error) {
	return "resolved " + w.Field, nil
}

func TestResolve(t *testing.T) {

	type Case struct {
		query    string
		expected interface{}
		isError  bool
	}

	cases := []Case{
		{
			query:    `{ withContext }`,
			expected: map[string]interface{}{"withContext": "ctxValue"},
		},
		{
			query:    `{ withContextAndResolveInfo }`,
			expected: map[string]interface{}{"withContextAndResolveInfo": "ctxValue withContextAndResolveInfo"},
		},
		{
			query:    `{ withContextAndResolveInfoAndTaggedInput( field: "foobar" ) }`,
			expected: map[string]interface{}{"withContextAndResolveInfoAndTaggedInput": "ctxValue withContextAndResolveInfoAndTaggedInput foobar"},
		},
		{
			query:    `{ withContextAndResolveInfoAndTaggedInput( field: "foobar" ) }`,
			expected: map[string]interface{}{"withContextAndResolveInfoAndTaggedInput": "ctxValue withContextAndResolveInfoAndTaggedInput foobar"},
		},
		{
			query:    `{ withContextAndResolveInfoAndTaggedNonNullInput( field: "foobar" ) }`,
			expected: map[string]interface{}{"withContextAndResolveInfoAndTaggedNonNullInput": "ctxValue withContextAndResolveInfoAndTaggedNonNullInput foobar"},
		},
		{
			query:    `{ withContextAndResolveInfoAndTaggedNonNullInput( field: "foobar" ) }`,
			expected: map[string]interface{}{"withContextAndResolveInfoAndTaggedNonNullInput": "ctxValue withContextAndResolveInfoAndTaggedNonNullInput foobar"},
		},
		{
			query:    `{ withContextAndResolveInfoAndTaggedNonNullInputSwapped1( field: "foobar" ) }`,
			expected: map[string]interface{}{"withContextAndResolveInfoAndTaggedNonNullInputSwapped1": "ctxValue withContextAndResolveInfoAndTaggedNonNullInputSwapped1 foobar"},
		},
		{
			query:    `{ withContextAndResolveInfoAndTaggedNonNullInputSwapped2( field: "foobar" ) }`,
			expected: map[string]interface{}{"withContextAndResolveInfoAndTaggedNonNullInputSwapped2": "ctxValue withContextAndResolveInfoAndTaggedNonNullInputSwapped2 foobar"},
		},
		{
			query:    `{ withContextAndResolveInfoAndTaggedInputWithTaggedOutput( field: "foobar" ) { field } }`,
			expected: map[string]interface{}{"withContextAndResolveInfoAndTaggedInputWithTaggedOutput": map[string]interface{}{"field": "ctxValue withContextAndResolveInfoAndTaggedInputWithTaggedOutput foobar"}},
		},
		{
			query:    `{ withContextAndResolveInfoAndTaggedInputWithTaggedOutputAndError( field: "foobar" ) { field } }`,
			expected: nil,
			isError:  true,
		},
		{
			query:    `{ withContextAndResolveInfoAndTaggedInputWithTaggedOutputAllPtr( field: "foobar" ) { field } }`,
			expected: map[string]interface{}{"withContextAndResolveInfoAndTaggedInputWithTaggedOutputAllPtr": map[string]interface{}{"field": "ctxValue withContextAndResolveInfoAndTaggedInputWithTaggedOutputAllPtr foobar"}},
		},
		{
			query:    `{ withNestedResolvers { resolvedField } }`,
			expected: map[string]interface{}{"withNestedResolvers": map[string]interface{}{"resolvedField": "resolved foobar"}},
		},
	}

	schema, err := NewSchemaBuilder().WithQuery(&Host{}).BuildSchema()
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	for _, c := range cases {
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

			if !reflect.DeepEqual(data, c.expected) {
				t.Errorf("expected %v, got %v", c.expected, data)
			}
		}
	}
}
