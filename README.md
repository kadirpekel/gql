# Go GraphQL Schema Builder

A flexible and type-safe GraphQL schema builder for Go that automatically generates GraphQL schemas from Go types and functions.

## Features

- 🔄 Automatic schema generation from Go structs and functions
- 🎯 Type-safe resolver functions
- 🏗️ Builder pattern for easy schema construction
- 💪 Support for Queries, Mutations, and Subscriptions
- 🔍 Automatic field resolution based on struct tags
- 📦 Built on top of graphql-go/graphql

## Installation

```bash
go get github.com/kadirpekel/gql
```

## Quick Start

Here's a simple example of how to use the schema builder:

```go
package main

import (
    "github.com/kadirpekel/gql"
)

type User struct {
    ID   string `gql:"id!"`
    Name string `gql:"name"`
}

func main() {
    query := map[string]interface{}{
        "getUser": func(id string) *User {
            return &User{ID: id, Name: "John Doe"}
        },
    }

    builder := gql.NewSchemaBuilder().
        WithQuery(query)

    schema, err := builder.BuildSchema()
    if err != nil {
        panic(err)
    }

    // Use the schema with your GraphQL server
}
```

## Usage

### Creating a Schema

The schema builder supports three main operation types:
- Queries (read operations)
- Mutations (write operations)
- Subscriptions (real-time updates)

```go
builder := gql.NewSchemaBuilder().
    WithQuery(queryMap).
    WithMutation(mutationMap).
    WithSubscription(subscriptionMap)
```

### Defining Types

Use struct tags to define GraphQL fields:

```go
type Product struct {
    ID          string  `gql:"id!"`      // Non-null field
    Name        string  `gql:"name"`     // Nullable field
    Price       float64 `gql:"price"`
    Description string                   // Ignored field
}
```

### Custom Resolvers

You can define custom resolvers for fields using the `Resolve` prefix:

```go
func (p *Product) ResolvePrice(ctx context.Context) float64 {
    // Custom price calculation logic
    return p.Price * 1.2
}
```

### Field Arguments

Support for field arguments in resolvers:

```go
type PriceInput struct {
    Currency string `gql:"currency!"`
}

func (p *Product) ResolvePrice(ctx context.Context, input PriceInput) float64 {
    // Convert price based on currency
    return convertPrice(p.Price, input.Currency)
}
```

## Supported Types

- Basic Types: `int`, `string`, `bool`, `float64`
- Arrays/Slices
- Structs
- Pointers to any supported type
- Custom types via struct tags

## Error Handling

The builder provides detailed error messages for:
- Invalid type definitions
- Resolver function signature mismatches
- Schema construction errors
- Type conversion errors

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

[Your chosen license]

## Credits

Built with [graphql-go/graphql](https://github.com/graphql-go/graphql)