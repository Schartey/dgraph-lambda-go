package model

import (
	"fmt"
	"io"
	"strconv"
	"time"

	"github.com/schartey/dgraph-lambda-go/examples/models"
	"github.com/twpayne/go-geom"
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
	Id    string `json:"id" dql:"uid"`
	Price int64  `json:"price" dql:"Apple.price"`
	Color string `json:"color" dql:"Apple.color"`
}
type Author struct {
	Id            string    `json:"id" dql:"uid"`
	Name          string    `json:"name" dql:"Author.name"`
	Posts         []*Post   `json:"posts" dql:"Author.posts"`
	RecentlyLiked []*Post   `json:"recentlyLiked" dql:"Author.recentlyLiked"`
	Friends       []*Author `json:"friends" dql:"Author.friends"`
}
type Figure struct {
	Id    string `json:"id" dql:"uid"`
	Shape string `json:"shape" dql:"Figure.shape"`
	Color string `json:"color" dql:"Figure.color"`
	Size  int64  `json:"size" dql:"Figure.size"`
}
type Hotel struct {
	Id       string  `json:"id" dql:"uid"`
	Name     string  `json:"name" dql:"Hotel.name"`
	Location *geom.T `json:"location" dql:"Hotel.location"`
	Area     *geom.T `json:"area" dql:"Hotel.area"`
}
type PointList struct {
	Points []*geom.T `json:"points" dql:"PointList.points"`
}
type User struct {
	UserID       string              `json:"userID" dql:"User.userID"`
	Credentials  *models.Credentials `json:"credentials" dql:"User.credentials"`
	Name         string              `json:"name" dql:"User.name"`
	LastSignIn   *time.Time          `json:"lastSignIn" dql:"User.lastSignIn"`
	RecentScores []float64           `json:"recentScores" dql:"User.recentScores"`
	Likes        int64               `json:"likes" dql:"User.likes"`
	Reputation   int64               `json:"reputation" dql:"User.reputation"`
	Rank         int64               `json:"rank" dql:"User.rank"`
	Active       bool                `json:"active" dql:"User.active"`
}

type Tag string

const (
	TagGraphQL  Tag = "GraphQL"
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
	case TagGraphQL, TagDatabase, TagQuestion:
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
