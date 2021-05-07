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
Within this startup function you can provide resolver functions and middleware. It's best to first define the input and output structs for the resolver. For example CreateUserInput and UserData struct
```go

type CreateUserInput struct {
	Username string `json:"username"`
}

type UserData struct {
	Id              string `json:"id"`
	Username        string `json:"username"`
	ComplexProperty string `json:"complexProperty"`
}
```
Then you can provide a resolver for fields, queries and mutations like this
```go
// Field Resolver
r.ResolveFunc("UserData.complexProperty", func(ctx context.Context, input []byte, parents []byte, ah resolver.AuthHeader) (interface{}, error) {
    var userParents []UserData
    json.Unmarshal(parents, &userParents)

    var complexProperties []string
    for _, userParent := range userParents {
        complexProperties = append(complexProperties, fmt.Sprintf("VeryComplex - %s", userParent.Id))
    }

    return complexProperties, nil
})

// Query/Mutation Resolver
r.ResolveFunc("Mutation.createUser", func(ctx context.Context, input []byte, parents []byte, ah resolver.AuthHeader) (interface{}, error) {
    var createUserInput CreateUserInput
    json.Unmarshal(input, &createUserInput)

    // Do Something
    user := UserData{
        Id:       "0x1",
        Username: createUserInput.Username,
    }
    return user, nil
})
```
You can also provide global middleware, as well as middleware on specific resolvers
```go
r.Use(func(hf resolver.HandlerFunc) resolver.HandlerFunc {
    return func(c context.Context, b []byte, parents []byte, ah resolver.AuthHeader) (interface{}, error) {
        // Authorization.
        // Add user to context
        return hf(c, b, parents, ah)
    }
})

r.UseOnResolver("Mutation.createUser", func(hf resolver.HandlerFunc) resolver.HandlerFunc {
    return func(c context.Context, b []byte, parents []byte, ah resolver.AuthHeader) (interface{}, error) {
        // Validation .
        return b, nil
    }
})
```
Finally a webhook resolver is also provided on Types
```go
r.WebHookFunc("UserData", func(ctx context.Context, event resolver.Event) error {
    // Send E-Mail
    return nil
})
```

Additionally a graphql and dql client connected to the dgraph server are provided, so you can query and make changes to the databases.

## Examples

Additional examples will be provided in the examples module
