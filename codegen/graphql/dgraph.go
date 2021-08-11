package graphql

/**
This part is taken from the official DGraph Source to provide all functionality for the schema validation.
**/

const (
	SchemaInputs = `
	"""
	The Int64 scalar type represents a signed 64‐bit numeric non‐fractional value.
	Int64 can represent values in range [-(2^63),(2^63 - 1)].
	"""
	scalar Int64
	"""
	The Int64 scalar type represents a signed 64‐bit numeric non‐fractional value.
	Int64 can represent values in range [-(2^63),(2^63 - 1)].
	"""
	scalar Password
"""
The DateTime scalar type represents date and time as a string in RFC3339 format.
For example: "1985-04-12T23:20:50.52Z" represents 20 minutes and 50.52 seconds after the 23rd hour of April 12th, 1985 in UTC.
"""
scalar DateTime
input IntRange{
	min: Int!
	max: Int!
}
input FloatRange{
	min: Float!
	max: Float!
}
input Int64Range{
	min: Int64!
	max: Int64!
}
input DateTimeRange{
	min: DateTime!
	max: DateTime!
}
input StringRange{
	min: String!
	max: String!
}
enum DgraphIndex {
	int
	int64
	float
	bool
	hash
	exact
	term
	fulltext
	trigram
	regexp
	year
	month
	day
	hour
	geo
}
input AuthRule {
	and: [AuthRule]
	or: [AuthRule]
	not: AuthRule
	rule: String
}
enum HTTPMethod {
	GET
	POST
	PUT
	PATCH
	DELETE
}
enum Mode {
	BATCH
	SINGLE
}
input CustomHTTP {
	url: String!
	method: HTTPMethod!
	body: String
	graphql: String
	mode: Mode
	forwardHeaders: [String!]
	secretHeaders: [String!]
	introspectionHeaders: [String!]
	skipIntrospection: Boolean
}
type Point {
	longitude: Float!
	latitude: Float!
}
input PointRef {
	longitude: Float!
	latitude: Float!
}
input NearFilter {
	distance: Float!
	coordinate: PointRef!
}
input PointGeoFilter {
	near: NearFilter
	within: WithinFilter
}
type PointList {
	points: [Point!]!
}
input PointListRef {
	points: [PointRef!]!
}
type Polygon {
	coordinates: [PointList!]!
}
input PolygonRef {
	coordinates: [PointListRef!]!
}
type MultiPolygon {
	polygons: [Polygon!]!
}
input MultiPolygonRef {
	polygons: [PolygonRef!]!
}
input WithinFilter {
	polygon: PolygonRef!
}
input ContainsFilter {
	point: PointRef
	polygon: PolygonRef
}
input IntersectsFilter {
	polygon: PolygonRef
	multiPolygon: MultiPolygonRef
}
input PolygonGeoFilter {
	near: NearFilter
	within: WithinFilter
	contains: ContainsFilter
	intersects: IntersectsFilter
}
input GenerateQueryParams {
	get: Boolean
	query: Boolean
	password: Boolean
	aggregate: Boolean
}
input GenerateMutationParams {
	add: Boolean
	update: Boolean
	delete: Boolean
}
`
	DirectiveDefs = `
directive @hasInverse(field: String!) on FIELD_DEFINITION
directive @search(by: [DgraphIndex!]) on FIELD_DEFINITION
directive @dgraph(type: String, pred: String) on OBJECT | INTERFACE | FIELD_DEFINITION
directive @id(interface: Boolean) on FIELD_DEFINITION
directive @withSubscription on OBJECT | INTERFACE | FIELD_DEFINITION
directive @secret(field: String!, pred: String) on OBJECT | INTERFACE
directive @auth(
	password: AuthRule
	query: AuthRule,
	add: AuthRule,
	update: AuthRule,
	delete: AuthRule) on OBJECT | INTERFACE
directive @custom(http: CustomHTTP, dql: String) on FIELD_DEFINITION
directive @remote on OBJECT | INTERFACE | UNION | INPUT_OBJECT | ENUM
directive @remoteResponse(name: String) on FIELD_DEFINITION
directive @cascade(fields: [String]) on FIELD
directive @lambda on FIELD_DEFINITION
directive @lambdaOnMutate(add: Boolean, update: Boolean, delete: Boolean) on OBJECT | INTERFACE
directive @cacheControl(maxAge: Int!) on QUERY
directive @generate(
	query: GenerateQueryParams,
	mutation: GenerateMutationParams,
	subscription: Boolean) on OBJECT | INTERFACE
`
	// see: https://www.apollographql.com/docs/federation/gateway/#custom-directive-support
	// So, we should only add type system directives here.
	// Even with type system directives, there is a bug in Apollo Federation due to which the
	// directives having non-scalar args cause issues in schema stitching in gateway.
	// See: https://github.com/apollographql/apollo-server/issues/3655
	// So, such directives have to be missed too.
	ApolloSupportedDirectiveDefs = `
directive @hasInverse(field: String!) on FIELD_DEFINITION
directive @search(by: [DgraphIndex!]) on FIELD_DEFINITION
directive @dgraph(type: String, pred: String) on OBJECT | INTERFACE | FIELD_DEFINITION
directive @id(interface: Boolean) on FIELD_DEFINITION
directive @withSubscription on OBJECT | INTERFACE | FIELD_DEFINITION
directive @secret(field: String!, pred: String) on OBJECT | INTERFACE
directive @remote on OBJECT | INTERFACE | UNION | INPUT_OBJECT | ENUM
directive @remoteResponse(name: String) on FIELD_DEFINITION
directive @lambda on FIELD_DEFINITION
directive @lambdaOnMutate(add: Boolean, update: Boolean, delete: Boolean) on OBJECT | INTERFACE
`
	FilterInputs = `
input IntFilter {
	eq: Int
	in: [Int]
	le: Int
	lt: Int
	ge: Int
	gt: Int
	between: IntRange
}
input Int64Filter {
	eq: Int64
	in: [Int64]
	le: Int64
	lt: Int64
	ge: Int64
	gt: Int64
	between: Int64Range
}
input FloatFilter {
	eq: Float
	in: [Float]
	le: Float
	lt: Float
	ge: Float
	gt: Float
	between: FloatRange
}
input DateTimeFilter {
	eq: DateTime
	in: [DateTime]
	le: DateTime
	lt: DateTime
	ge: DateTime
	gt: DateTime
	between: DateTimeRange
}
input StringTermFilter {
	allofterms: String
	anyofterms: String
}
input StringRegExpFilter {
	regexp: String
}
input StringFullTextFilter {
	alloftext: String
	anyoftext: String
}
input StringExactFilter {
	eq: String
	in: [String]
	le: String
	lt: String
	ge: String
	gt: String
	between: StringRange
}
input StringHashFilter {
	eq: String
	in: [String]
}
`

	apolloSchemaExtras = `
scalar _Any
scalar _FieldSet
type _Service {
	sdl: String
}
directive @external on FIELD_DEFINITION
directive @requires(fields: _FieldSet!) on FIELD_DEFINITION
directive @provides(fields: _FieldSet!) on FIELD_DEFINITION
directive @key(fields: _FieldSet!) on OBJECT | INTERFACE
directive @extends on OBJECT | INTERFACE
`
	apolloSchemaQueries = `
type Query {
	_entities(representations: [_Any!]!): [_Entity]!
	_service: _Service!
}
`
)
