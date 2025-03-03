package gql

import (
	"context"
	"reflect"
	"testing"

	"github.com/graphql-go/graphql"
)

type ValidFixtureInput struct {
	A string `gql:"a"`
	B int
}

type InvalidFixtureInput struct {
	A string
	B int
}

type ValidFixtureOutput struct {
	A string `gql:"a"`
	B int
}

type InvalidFixtureOutput struct {
	A string
	B int
}

type FixtureType struct{}

func (f FixtureType) NoInputNoOutput() {}

func (f FixtureType) NoInputWithoutOutput() error {
	return nil
}

func (f FixtureType) NoInputWithoutError() string {
	return "foo"
}

func (f FixtureType) NoInputInvalidOutput() InvalidFixtureOutput {
	return InvalidFixtureOutput{A: "foo", B: 1}
}

func (f FixtureType) NoInput() (string, error) {
	return "foo", nil
}

func (f FixtureType) NonStructInput(a int) (string, error) {
	return "foo", nil
}

func (f FixtureType) MoreThanThreeInputs(
	a ValidFixtureInput,
	b context.Context,
	c graphql.ResolveInfo,
	d ValidFixtureInput,
) (string, error) {
	return "foo", nil
}

func (f FixtureType) MoreThanOneInputType(
	a ValidFixtureInput,
	b context.Context,
	d ValidFixtureInput,
) (string, error) {
	return "foo", nil
}

func (f FixtureType) MoreThanTwoReturns(a ValidFixtureInput, b context.Context, c graphql.ResolveInfo) (int, string, error) {
	return 1, "foo", nil
}

func (f FixtureType) OneInput(a ValidFixtureInput) (string, error) {
	return "foo", nil
}

func (f FixtureType) InvalidInput(a InvalidFixtureInput) (string, error) {
	return "foo", nil
}

func (f FixtureType) TwoInputs(a ValidFixtureInput, b context.Context) (string, error) {
	return "foo", nil
}

func (f FixtureType) ThreeInputs(a ValidFixtureInput, b context.Context, c graphql.ResolveInfo) (string, error) {
	return "foo", nil
}

func (f FixtureType) ThreeInputsWithStructOutput(a ValidFixtureInput, b context.Context, c graphql.ResolveInfo) (ValidFixtureOutput, error) {
	return ValidFixtureOutput{}, nil
}

func (f FixtureType) ThreeInputsThreeOutputs(a ValidFixtureInput, b context.Context, c graphql.ResolveInfo) (int, string, error) {
	return 1, "foo", nil
}

func TestNewResolveInfo(t *testing.T) {
	fixtureType := reflect.TypeOf(FixtureType{})
	fnMap := make(map[string]reflect.Value)
	for i := 0; i < fixtureType.NumMethod(); i++ {
		method := fixtureType.Method(i)
		fnMap[method.Name] = method.Func
	}

	cases := []struct {
		fn      reflect.Value
		isError bool
	}{
		{
			fn:      fnMap["NoInputNoOutput"],
			isError: true,
		},
		{
			fn:      fnMap["NoInputWithoutOutput"],
			isError: true,
		},
		{
			fn:      fnMap["NoInputWithoutError"],
			isError: true,
		},
		{
			fn: fnMap["NoInput"],
		},
		{
			fn: fnMap["OneInput"],
		},
		{
			fn: fnMap["TwoInputs"],
		},
		{
			fn: fnMap["ThreeInputs"],
		},
		{
			fn: fnMap["ThreeInputsWithStructOutput"],
		},
		{
			fn:      fnMap["NonStructInput"],
			isError: true,
		},
		{
			fn:      fnMap["MoreThanThreeInputs"],
			isError: true,
		},
		{
			fn:      fnMap["MoreThanOneInputType"],
			isError: true,
		},
		{
			fn:      fnMap["MoreThanTwoReturns"],
			isError: true,
		},
		{
			fn:      fnMap["InvalidInput"],
			isError: true,
		},
		{
			fn:      fnMap["NoInputInvalidOutput"],
			isError: true,
		},
		{
			fn:      fnMap["ThreeInputsThreeOutputs"],
			isError: true,
		},
	}

	for _, c := range cases {
		_, err := NewResolveInfo(c.fn)
		if c.isError {
			if err == nil {
				t.Errorf("expected error, got %v", err)
			}
		} else {
			if err != nil {
				t.Errorf("expected no error, got %v", err)
			}
		}
	}
}
