package model

import(
	"fmt"
	"github.com/twpayne/go-geom"
	"io"
	"github.com/schartey/dgraph-lambda-go/examples/models"
	"strconv"
	"time"
)
type Color interface {
	IsColor()
}
type Fruit interface {
	IsFruit()
}
type Post interface {
	IsPost()
}
type Shape interface {
	IsShape()
}
type Apple struct {
		Id string `json:"id"`
		Price int64 `json:"price"`
		Color string `json:"color"`
}
type Author struct {
		Id string `json:"id"`
		Name string `json:"name"`
		Posts []*Post `json:"posts"`
		RecentlyLiked []*Post `json:"recentlyLiked"`
		Friends []*Author `json:"friends"`
}
type Figure struct {
		Id string `json:"id"`
		Shape string `json:"shape"`
		Color string `json:"color"`
		Size int64 `json:"size"`
}
type Hotel struct {
		Id string `json:"id"`
		Name string `json:"name"`
		Location *geom.T `json:"location"`
		Area *geom.T `json:"area"`
}
type PointList struct {
		Points []*geom.T `json:"points"`
}
type User struct {
		UserID string `json:"userID"`
		Credentials *models.Credentials `json:"credentials"`
		Name string `json:"name"`
		LastSignIn *time.Time `json:"lastSignIn"`
		RecentScores []float64 `json:"recentScores"`
		Likes int64 `json:"likes"`
		Reputation int64 `json:"reputation"`
		Rank int64 `json:"rank"`
		Active bool `json:"active"`
}

type Tag string
const (
	TagGraphQL Tag = "GraphQL"
	TagDatabase Tag = "Database"
	TagQuestion Tag = "Question"
)

var AllTag = []Tag{
	TagGraphQL,
	TagDatabase,
	TagQuestion,
}

func (e Tag) IsValid() bool {
	switch e {
	case TagGraphQL,TagDatabase,TagQuestion:
		return true
	}
	return false
}

func (e Tag) String() string {
	return string(e)
}

func (e *Tag) UnmarshalGQL(v interface{}) error {
	str, ok := v.(string)
	if !ok {
		return fmt.Errorf("enums must be strings")
	}

	*e = Tag(str)
	if !e.IsValid() {
		return fmt.Errorf("%s is not a valid Tag", str)
	}
	return nil
}

func (e Tag) MarshalGQL(w io.Writer) {
	fmt.Fprint(w, strconv.Quote(e.String()))
}
