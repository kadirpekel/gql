package main

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/graphql-go/graphql"
	"github.com/kadirpekel/gql"
)

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

func (q query) GetUser(args UserInput) (*User, error) {
	return &User{ID: args.ID, FirstName: "John", LastName: "Doe"}, nil
}

func (q query) ListUsers() ([]*User, error) {
	return []*User{
		{ID: "1", FirstName: "John", LastName: "Doe"},
		{ID: "2", FirstName: "Jane", LastName: "Doe"},
	}, nil
}

func TestSimple(t *testing.T) {
	schema, err := gql.NewSchemaBuilder().
		WithQuery(query{}).
		BuildSchema()

	if err != nil {
		panic(err)
	}

	params := graphql.Params{
		Schema: *schema,
		RequestString: `
			query {
				getUser(ID: "1") {
					ID
					firstName
					lastName
					fullName
				}
			}
		`,
	}

	result := graphql.Do(params)
	if len(result.Errors) > 0 {
		t.Fatalf("expected no errors, got %v", result.Errors)
	}

	if result.Data == nil {
		t.Fatalf("expected data, got nil")
	}

	expected := map[string]interface{}{
		"getUser": map[string]interface{}{
			"firstName": "John",
			"lastName":  "Doe",
			"fullName":  "John Doe",
			"ID":        "1",
		},
	}

	if !reflect.DeepEqual(result.Data, expected) {
		t.Fatalf("expected data, got %v", result.Data)
	}

	params = graphql.Params{
		Schema: *schema,
		RequestString: `
			query {
				listUsers {
					ID
					firstName
					lastName
					fullName
				}
			}
		`,
	}

	result = graphql.Do(params)
	if len(result.Errors) > 0 {
		t.Fatalf("expected no errors, got %v", result.Errors)
	}

	expected = map[string]interface{}{
		"listUsers": []interface{}{
			map[string]interface{}{"ID": "1", "firstName": "John", "lastName": "Doe", "fullName": "John Doe"},
			map[string]interface{}{"ID": "2", "firstName": "Jane", "lastName": "Doe", "fullName": "Jane Doe"},
		},
	}

	if !reflect.DeepEqual(result.Data, expected) {
		t.Fatalf("expected data, got %v", result.Data)
	}
}
