# dgraph-lambda-go

*Note: this is a fork, for the original project by Schartey go to https://github.com/Schartey/dgraph-lambda-go*

Go Library written to build Dgraph Lambda servers as an alternative to the [Dgraph JS Lambda Server](https://github.com/dgraph-io/dgraph-lambda)

It is currently in **development**! Please create an issue if something is not working correctly.

If you would like to support me please visit my [:coffee:](https://ko-fi.com/schartey)

## Getting started

- Create project ```go mod init```
- To install dgraph-lambda-go run the command ```go get -d github.com/miko/dgraph-lambda-go``` in your project directory.
- Then initialize the project by running ```go run github.com/miko/dgraph-lambda-go init```.
- Set path to your graphql schema in lambda.yaml
- Generate types and resolvers ```go run github.com/miko/dgraph-lambda-go generate```
- Implement your lambda resolvers
- Run your server ```go run server.go```


## Configuration

When first initializing the lambda server it will generate a basic lambda.yaml file with the following configuration:

    schema:
      - ../trendgraph/dgraph/*.graphql

    exec:
      filename: lambda/generated/generated.go
      package: generated

    model:
      filename: lambda/model/models_gen.go
      package: model

    force:
      - "Home"

    autobind:
      - "github.com/miko/dgraph-lambda-go/examples/models"

    resolver:
      dir: lambda/resolvers
      package: resolvers
      filename_template: "{resolver}.resolver.go" # also allow "{name}.resolvers.go"

    server:
      standalone: true

### Schema

A list of graphql schema files using glob. This is probably only one file when using DGraph.

### Exec

This option allows you to select a file path where generated code should go that should NOT be edited.

### Model

This option allows you to define a file path where the generated models should be placed.

### Force

Force generation of specific models

### Autobind

You might have some predefined models already. Here you can define a list of packages in which models can be found that should be used instead of generating them. Models you add within the generated model folder will be autobound automatically.

### Resolver

Define a folder and package name for the generated resolvers. The filename_template can currently only be {resolver}.resolver.go, but I want to allow resolver generation based on name as well in the future. Using {resolver}.resolver.go will generate a fieldResolver.go, queryResolver.go, mutationResolver.go, webhookResolver.go and middlewareResolver.go file where each type of resolver will reside in.

### Server

On initialization a server.go file is generated from which you can start the server. With standalone set to false you can add custom routes to the http server.


## Generating resolvers

This framework is able to generate field, query, mutation and webhook resolvers. These will automatically be detected in the graphql schema file.
To generate middleware you have to use comments within the schema. For example:

### Type Fields:
```graphql
type User @lambdaOnMutate(add: true, update: true, delete: true) {
    id: ID!
    username: String!
    """
    @middleware(["auth"])
    """
    secret: string @lambda
}
```
### Queries
```graphql
type Query {
    """
    @middleware(["auth"])
    """
    randomUser(seed: String): User @lambda
}
```

### Mutations
```graphql
type Mutation {
    """
    @middleware(["auth"])
    """
    createUser(input: CreateUserInput!): User @lambda
}
```


## Implementing resolvers

Here are implementations from the above mentioned schema examples.
### Field Resolver

```golang
func (f *FieldResolver) User_secret(ctx context.Context, parents []string, authHeader api.AuthHeader) ([]string, error) { 
	var secrets []string
    for _, userParent := range userParents {
        secrets = append(secrets, fmt.Sprintf("Secret - %s", userParent.Id))
    }
    return secrets, nil
}
```

### Query Resolver

```golang
func (q *QueryResolver) Query_randomUser(ctx context.Context, seed string, authHeader api.AuthHeader) (*model.User, error) { 
	nameGenerator := namegenerator.NewNameGenerator(seed)
    name := nameGenerator.Generate()

    user := &model.User{
        Id:       "0x1",
        Username: name,
    }
    return user, nil
}
```

### Mutation Resolver

```golang
func (q *MutationResolver) Mutation_createUser(ctx context.Context, input *model.CreateUserInput, authHeader api.AuthHeader) (*model.User, error) {
    user := &User{
        Id:       "0x1",
        Username: createUserInput.Username,
    }
	return user, nil
}
```

### Webhook Resolver

```golang
func (w *WebhookResolver) Webhook_User(ctx context.Context, event api.Event) error {
    // Send Email
	return nil
}
```

### Middleware Resolver

```golang
func (m *MiddlewareResolver) Middleware_auth(md *api.MiddlewareData) error {
    // Check Token
    valid := true //false
    if valid {
    	md.Ctx = context.WithValue(md.Ctx, "logged_in", "true")
        return nil
    } else {
        return errors.New("Token invalid!")
    }
}
```

## Inject custom dependencies

Typically you want to at least inject a graphql/dql client into your resolvers. To do so just add your client to the Resolver struct
```golang
// Add objects to your desire
type Resolver struct {
    Dql *dgo.Dgraph
}
```
and pass the client to the executor in your generated server.go file
```golang
dql := NewDqlClient()
resolver := &resolvers.Resolver{ Dql: dql}
executer := generated.NewExecuter(resolver)
```
Then you can access the client in your resolvers like this
```golang
func (q *QueryResolver) Query_randomUser(ctx context.Context, seed string, authHeader api.AuthHeader) (*model.User, error) {
    // Oversimplified
    vars := map[string]string{"$uid": uid}
	query := `
		query findUser($uid: string) {
			findUser(func: uid($uid)) {
				id: uid
			}
		}`

	res, err := s.dql.NewReadOnlyTxn().QueryWithVars(ctx, query, vars)
	if err != nil {
		return nil, err
	}

    // You can use the provided dql json unmarshaller
	var findUserResult struct {
        FindUser []model.User `dql:"findUser"`
    }
	dson.Unmarshal(res.GetJson(), &findUserResult)

	return &findUserResult.FindUser[0], nil
}
```

## Notes
- Working with Cyclic types
If you have types in your schema with @hasInverse and one of them should not be generated, but you provide your own (see Autobind), then you should put that type declaration into the generation folder with the models_gen.go file. Otherwise the imports will be cyclic and Golang does not like that. See example under examples/lambda/model/cyclic_model.go

- When using graphql to generate the dgraph schema, type fields are prefixed with the type (Type.field). dgraph-lambda-go provides a json parser that uses the tag "dql" and is able to convert dql query results into the generated models using jsoniter. Example:

Schema
```graphql
type User {
    id: ID!
    username: String!
}
```

Model
```golang
type User struct {
	Id       string `json:"id" dql:"uid"`
	Username string `json:"username" dql:"User.username"`
}
```

Code
```golang
    vars := map[string]string{"$uid": uid}
	query := `
		query findUser($uid: string) {
			findUser(func: uid($uid)) {
				id: uid
			}
		}`

	res, err := s.dql.NewReadOnlyTxn().QueryWithVars(ctx, query, vars)
	if err != nil {
		return nil, err
	}

// ======== Here we use the dson unmarshaller ======
	var findUserResult struct {
        FindUser []model.User `dql:"findUser"`
    }
	dson.Unmarshal(res.GetJson(), &findUserResult)
// =================================================

	return &findUserResult.FindUser[0], nil
```
## Known Issues

- In DGraph it is allowed to skip fields in types that are already implemented in the interface. The GraphQl parser used for this project is very strict on the GraphQl specs and does not allow this, so you have to copy all fields you are using in the interface to your type.

## Examples

Additional examples will be provided in the examples module
