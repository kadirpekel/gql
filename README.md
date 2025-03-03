# Gql - The Missing GraphQL Schema Builder for Golang

A flexible and type-safe GraphQL schema builder for Go that automatically generates GraphQL schemas from Go types and functions. It nicely stands on the shoulders of [`github.com/graphql-go/graphql`](https://github.com/graphql-go/graphql), a battle-tested package widely used in production environments.

## Features

- **Type-Safe**: Uses Go types to define your GraphQL schema effortlessly.
- **Automatic Schema Generation**: No need to manually define GraphQL types; just annotate your Go structs.
- **Built on `graphql-go/graphql`**: Leverages an established GraphQL implementation for Go.
- **Flexible and Extensible**: Supports custom resolvers, input types, and mutations.
- **Flexible Resolver Signature**: Resolver methods can accept `context.Context`, `graphql.ResolveInfo`, and input structs in any order and combination.
- **GQL Tags for Struct Fields**: Uses struct tags to define GraphQL fields and arguments easily.

## Installation

```bash
go get github.com/kadirpekel/gql
```

## Quick Start

Here's a simple example of how to define and use a GraphQL schema with `gql`:

```go
package main

import (
	"context"
	"fmt"
	"github.com/graphql-go/graphql"
	"github.com/kadirpekel/gql"
)

// Struct fields are annotated with gql tags to define GraphQL schema
// The tag format is `gql:"fieldName[,modifier]"`
// Example: `gql:"ID,nonNull"` ensures the field is required in GraphQL

type User struct {
	ID        string `gql:"ID"`
	FirstName string `gql:"firstName"`
	LastName  string `gql:"lastName"`
}

func (u *User) FullName() (string, error) {
	return fmt.Sprintf("%s %s", u.FirstName, u.LastName), nil
}

type UserInput struct {
	ID string `gql:"ID,nonNull"`
}

type query struct{}

func (q query) GetUser(ctx context.Context, args UserInput, info graphql.ResolveInfo) (*User, error) {
	return &User{ID: args.ID, FirstName: "John", LastName: "Doe"}, nil
}

func (q query) ListUsers(info graphql.ResolveInfo) ([]*User, error) {
	return []*User{
		{ID: "1", FirstName: "John", LastName: "Doe"},
		{ID: "2", FirstName: "Jane", LastName: "Doe"},
	}, nil
}

func main() {
	schema, err := gql.NewSchemaBuilder().
		WithQuery(query{}).
		BuildSchema()

	if err != nil {
		panic(err)
	}

	// Use the schema with your GraphQL server
}
```

## GQL Tags for Struct Fields

The `gql` struct tags define GraphQL schema properties directly on Go structs:

- **Basic Mapping**: `gql:"fieldName"` maps the Go struct field to a GraphQL field.
- **Modifiers**: Add modifiers such as `nonNull` for required fields.
- **Example Usage**:

```go
type User struct {
	ID   string `gql:"ID,nonNull"` // Required GraphQL ID field
	Name string `gql:"name"`       // Maps to GraphQL "name" field
}
```

## Resolver Method Signature

Resolvers in `gql` are flexible and can accept parameters in any order:

- `context.Context`: Allows passing request-scoped values like authentication data.
- `graphql.ResolveInfo`: Provides details about the query execution.
- Input structs: Used for passing arguments to the resolver.

For example, all of the following resolver signatures are valid:

```go
func (q query) GetUser(ctx context.Context, args UserInput) (*User, error) {}
func (q query) GetUser(info graphql.ResolveInfo, ctx context.Context, args UserInput) (*User, error) {}
func (q query) GetUser(args UserInput) (*User, error) {}
```

## Defining Mutations

You can also define mutations using the same approach:

```go
type mutation struct{}

func (m mutation) CreateUser(ctx context.Context, args User) (*User, error) {
	return &User{ID: "3", FirstName: args.FirstName, LastName: args.LastName}, nil
}
```

Then include it in your schema:

```go
schema, err := gql.NewSchemaBuilder().
	WithQuery(query{}).
	WithMutation(mutation{}).
	BuildSchema()
```

## Running a GraphQL Server

To integrate with a GraphQL server, use `github.com/graphql-go/handler`:

```go
import (
	"net/http"
	"github.com/graphql-go/handler"
)

func main() {
	schema, err := gql.NewSchemaBuilder().WithQuery(query{}).BuildSchema()
	if err != nil {
		panic(err)
	}

	h := handler.New(&handler.Config{
		Schema:   schema,
		GraphiQL: true,
	})

	http.Handle("/graphql", h)
	http.ListenAndServe(":8080", nil)
}
```

## License

This project is licensed under the MIT License.
