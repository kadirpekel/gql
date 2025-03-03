package gql

import (
	"context"
	"errors"
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

func (f FixtureType) NonStructInput(a int) {}

func (f FixtureType) MoreThanThreeInputs(a ValidFixtureInput, b context.Context, c graphql.ResolveInfo, d ValidFixtureInput) error {
	return nil
}

func (f FixtureType) MoreThanOneInputType(a ValidFixtureInput, b context.Context, d ValidFixtureInput) error {
	return nil
}

func (f FixtureType) NoOutput(a ValidFixtureInput, b context.Context, c graphql.ResolveInfo) {}

func (f FixtureType) MoreThanTwoReturns(a ValidFixtureInput, b context.Context, c graphql.ResolveInfo) (int, string, error) {
	return 1, "foo", nil
}

func (f FixtureType) OneInputOneOutput(a ValidFixtureInput) int {
	return 1
}

func (f FixtureType) OneInputOneOutputWithInvalidInput(a InvalidFixtureInput) int {
	return 1
}

func (f FixtureType) OneInputOneOutputWithInvalidOutput(a ValidFixtureInput) InvalidFixtureOutput {
	return InvalidFixtureOutput{A: "foo", B: 1}
}

func (f FixtureType) OneInputTwoOutputsWithoutError(a ValidFixtureInput) (int, string) {
	return 1, "foo"
}

func (f FixtureType) TwoInputsOneOutput(a ValidFixtureInput, b context.Context) int {
	return 1
}

func (f FixtureType) TwoInputsTwoOutputs(a ValidFixtureInput, b context.Context) (int, error) {
	return 1, nil
}

func (f FixtureType) ThreeInputsOneOutput(a ValidFixtureInput, b context.Context, c graphql.ResolveInfo) int {
	return 1
}

func (f FixtureType) ThreeInputsTwoOutputs(a ValidFixtureInput, b context.Context, c graphql.ResolveInfo) (int, error) {
	return 1, nil
}

func (f FixtureType) ThreeInputsTwoOutputsWithStruct(a ValidFixtureInput, b context.Context, c graphql.ResolveInfo) (ValidFixtureOutput, error) {
	return ValidFixtureOutput{}, nil
}

func (f FixtureType) ThreeInputsThreeOutputs(a ValidFixtureInput, b context.Context, c graphql.ResolveInfo) (int, string, error) {
	return 1, "foo", nil
}

func UnboundNoInputNoOutput() {}

func UnboundNoInputNoOutputWithError() error {
	return errors.New("error")
}

func UnboundNoInputWithOutput() int {
	return 1
}

func UnboundOneInputNoOutput(a ValidFixtureInput) {}

func UnboundOneInputWithOutput(a ValidFixtureInput) int {
	return 1
}

func UnboundTwoInputsWithOutput(a ValidFixtureInput, b context.Context) int {
	return 1
}

func UnboundThreeInputsWithOutput(a ValidFixtureInput, b context.Context, c graphql.ResolveInfo) int {
	return 1
}

func UnboundFourInputsWithOutput(a ValidFixtureInput, b context.Context, c graphql.ResolveInfo, d ValidFixtureInput) int {
	return 1
}

func UnboundFourInputsWithOutputAndError(a ValidFixtureInput, b context.Context, c graphql.ResolveInfo, d ValidFixtureInput) (int, error) {
	return 1, errors.New("error")
}

func TestNewResolveInfo(t *testing.T) {
	fixtureType := reflect.TypeOf(FixtureType{})
	fnMap := make(map[string]reflect.Value)
	for i := 0; i < fixtureType.NumMethod(); i++ {
		method := fixtureType.Method(i)
		fnMap[method.Name] = method.Func
	}

	cases := []struct {
		fn        reflect.Value
		isError   bool
		isUnbound bool
	}{
		{
			fn:        reflect.ValueOf(UnboundNoInputNoOutput),
			isUnbound: true,
			isError:   true,
		},
		{
			fn:        reflect.ValueOf(UnboundOneInputNoOutput),
			isUnbound: true,
			isError:   true,
		},
		{
			fn:        reflect.ValueOf(UnboundNoInputNoOutputWithError),
			isUnbound: true,
			isError:   true,
		},
		{
			fn:        reflect.ValueOf(UnboundNoInputWithOutput),
			isUnbound: true,
		},
		{
			fn:        reflect.ValueOf(UnboundOneInputWithOutput),
			isUnbound: true,
		},
		{
			fn:        reflect.ValueOf(UnboundTwoInputsWithOutput),
			isUnbound: true,
		},
		{
			fn:        reflect.ValueOf(UnboundThreeInputsWithOutput),
			isUnbound: true,
		},
		{
			fn: fnMap["NoInputNoOutput"],
		},
		{
			fn: fnMap["OneInputOneOutput"],
		},
		{
			fn: fnMap["TwoInputsOneOutput"],
		},
		{
			fn: fnMap["TwoInputsTwoOutputs"],
		},
		{
			fn: fnMap["ThreeInputsOneOutput"],
		},
		{
			fn: fnMap["ThreeInputsTwoOutputsWithStruct"],
		},
		{
			fn:        reflect.ValueOf(UnboundFourInputsWithOutput),
			isUnbound: true,
			isError:   true,
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
			fn:      fnMap["OneInputOneOutputWithInvalidInput"],
			isError: true,
		},
		{
			fn:      fnMap["OneInputOneOutputWithInvalidOutput"],
			isError: true,
		},
		{
			fn:      fnMap["OneInputTwoOutputsWithoutError"],
			isError: true,
		},
		{
			fn:      fnMap["ThreeInputsThreeOutputs"],
			isError: true,
		},
	}

	for _, c := range cases {
		_, err := NewResolveInfo(c.fn, !c.isUnbound)
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
