package gql

import (
	"fmt"
	"reflect"

	"github.com/graphql-go/graphql"
)

/*
ResolveInfo is a struct that contains information about the function that is being resolved.

It contains the function itself, whether it is bound or not, the source, context, info, input, output and error.

The source, context, info, input and output are all ArgInfo structs.

Example Signature Mapping:

	func (source *SomeStruct) SomeMethod(context context.Context, info *graphql.ResolveInfo, input *SomeInput) (OutputType, error) {
		...
	}
*/
type ResolveInfo struct {
	Func    reflect.Value
	IsBound bool
	Source  *ArgInfo
	Context *ArgInfo
	Info    *ArgInfo
	Input   *ArgInfo
	Output  *ArgInfo
	Error   *ArgInfo
}

func hasStructValidGqlTag(t reflect.Type) bool {
	for _, field := range reflect.VisibleFields(t) {
		tag, err := ParseGqlTagFromField(&field)
		if err == nil && tag.FieldName != "" {
			return true
		}
	}
	return false
}

func (r *ResolveInfo) Validate() error {
	if r.Input != nil {
		if r.Input.RealType.Kind() != reflect.Struct || r.Input.IsSlice {
			return fmt.Errorf("Input type should be a struct, got %s", r.Input.Type)
		}

		if !hasStructValidGqlTag(r.Input.RealType) {
			return fmt.Errorf("Input type should have at least one field with a gql tag")
		}
	}

	if r.Output == nil {
		if !r.IsBound {
			return fmt.Errorf("unbound resolvers should have an output type")
		}
	} else {
		if r.Output.RealType.Kind() == reflect.Struct && !hasStructValidGqlTag(r.Output.RealType) {
			return fmt.Errorf("Output type should have at least one field with a gql tag")
		}
	}

	return nil
}

func NewResolveInfo(fn reflect.Value, IsBound bool) (*ResolveInfo, error) {
	r := &ResolveInfo{
		Func:    fn,
		IsBound: IsBound, // This can be maybe auto-detected later on
	}

	maxNumberOfArgs := 3
	baseIndex := 0

	// If the method is bound, the first argument is the source
	if IsBound {
		maxNumberOfArgs = 4
		baseIndex = 1

		if fn.Type().NumIn() == 0 {
			return nil, fmt.Errorf("Resolve method should have a receiver")
		}

		r.Source = NewArgInfo(fn.Type().In(0), 0)

		if r.Source.RealType.Kind() != reflect.Struct {
			return nil, fmt.Errorf("Resolve method should be hosted on a struct, got %s", r.Source.Type)
		}
	}

	// Other validations on the function signature
	if fn.Type().NumIn() > maxNumberOfArgs {
		return nil, fmt.Errorf("Resolve method should have at most %d arguments", maxNumberOfArgs)
	}

	if fn.Type().NumOut() > 2 {
		return nil, fmt.Errorf("Resolve method should have at most 2 return values")
	}

	// Iterate over the input types and determine the context, info, input and error types
	// along with the index
	for i := baseIndex; i < fn.Type().NumIn(); i++ {
		argInfo := NewArgInfo(fn.Type().In(i), i)
		if argInfo.RealType == ContextType {
			r.Context = argInfo
		} else if argInfo.RealType == InfoType {
			r.Info = argInfo
		} else {
			if r.Input == nil {
				r.Input = argInfo
			} else {
				return nil, fmt.Errorf("Expected at most one input type, got %s", argInfo.Type)
			}
		}
	}

	// Iterate over the output types and determine the output and error types along with the index
	for i := 0; i < fn.Type().NumOut(); i++ {
		argInfo := NewArgInfo(fn.Type().Out(i), i)
		if argInfo.RealType == ErrorType {
			r.Error = argInfo
		} else {
			if r.Output == nil {
				r.Output = argInfo
			} else {
				return nil, fmt.Errorf("Expected at most one output type, got %s", argInfo.Type)
			}
		}
	}

	if err := r.Validate(); err != nil {
		return nil, err
	}

	return r, nil
}

func (r *ResolveInfo) Resolve(p graphql.ResolveParams) (interface{}, error) {
	args := make([]reflect.Value, r.Func.Type().NumIn())
	var err error

	// If the method is bound, the first argument is the source
	if r.IsBound {
		args[0], err = r.Source.ValueFrom(p.Source)
		if err != nil {
			return nil, err
		}
	}

	// If there is an input, place it in the input index
	if r.Input != nil {
		args[r.Input.Index], err = r.Input.ValueFrom(p.Args)
		if err != nil {
			return nil, err
		}
	}

	// If there is a context, place it in the context index
	if r.Context != nil {
		args[r.Context.Index] = reflect.ValueOf(p.Context)
	}

	// If there is an info, place it in the info index
	if r.Info != nil {
		args[r.Info.Index], err = r.Info.ValueFrom(p.Info)
		if err != nil {
			return nil, err
		}
	}

	// Call the function with the arguments in the correct order
	values := r.Func.Call(args)

	// If there is an output, place it in the output index
	var output interface{}
	if r.Output != nil {
		output = values[r.Output.Index].Interface()
	}

	if r.Error != nil {
		err, ok := values[r.Error.Index].Interface().(error)
		if ok && err != nil {
			return nil, err
		}
	}
	return output, nil
}
