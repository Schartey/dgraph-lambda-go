# dgraph-lambda-go

Go Library written to build Dgraph Lambda servers as an alternative to the [Dgraph JS Lambda Server](https://github.com/dgraph-io/dgraph-lambda)

## Getting started

- To install dgraph-lambda-go run the command ```go get github.com/schartey/dgraph-lambda-go``` in your project directory.
- Then initialize the project by running ```go run github.com/schartey/dgraph-lambda-go init```.

## Implement resolver functions and middleware

On startup this library provides a resolver. 
```go
err := api.RunServer(func(r *resolver.Resolver, gql *graphql.Client, dql *dgo.Dgraph) {

})
```
Within this startup function you can provide resolver functions and middleware. It's best to first define the input structs for the resolver. For example CreateUserInput struct
```go
type CreateUserInput struct {
	Username string `json:"username"`
}
```
Then you can provide a resolver for fields, queries and mutations like this
```go
// Field Resolver
r.ResolveFunc("User.complexProperty", func(ctx context.Context, input []byte, parents []byte, ah resolver.AuthHeader) ([]byte, error) {
    fmt.Println(string(parents))

    resp := `
    [ "complexPropertyValue" ]`
    return ([]byte)(resp), nil
})

// Query/Mutation Resolver
r.ResolveFunc("Mutation.createUser", func(ctx context.Context, input []byte, ah resolver.AuthHeader) ([]byte, error) {
    var createUserInput CreateUserInput
    json.Unmarshal(input, &createUserInput)

    // Do Something

    resp := `
    {
        "id": "0x1"	
    }`
    return ([]byte)(resp), nil
})
```
You can also provide middleware
```go
r.Use(func(hf resolver.HandlerFunc) resolver.HandlerFunc {
    return func(c context.Context, b []byte, ah resolver.AuthHeader) ([]byte, error) {
        // For example authentication.
        // Add user to context
        return b, nil
    }
})

r.UseOnResolver("Mutation.createUser", func(hf resolver.HandlerFunc) resolver.HandlerFunc {
    return func(c context.Context, b []byte, ah resolver.AuthHeader) ([]byte, error) {
        // For example authentication.
        // Add user to context
        return b, nil
    }
})
```
Additionally a graphql and dql client connected to the dgraph server are provided, so you can query and make changes to the databases.

## Examples

Additional examples will be provided in the examples module
