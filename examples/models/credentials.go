package models

import (
	"fmt"
	"io"
	"strconv"
)

type Credentials struct {
	Id          uint64 `json:"id"`
	FirebaseId  string `json:"firebaseId"`
	Email       string `json:"email"`
	PhoneNumber string `json:"phoneNumber"`
}

type EnumTest string

const (
	EnumTestASDF EnumTest = "ASDF"
	EnumTestFDAS EnumTest = "FDAS"
)

var AllEnumTest = []EnumTest{
	EnumTestASDF,
	EnumTestFDAS,
}

func (e EnumTest) IsValid() bool {
	switch e {
	case EnumTestASDF, EnumTestFDAS:
		return true
	}
	return false
}

func (e EnumTest) String() string {
	return string(e)
}

func (e *EnumTest) UnmarshalGQL(v interface{}) error {
	str, ok := v.(string)
	if !ok {
		return fmt.Errorf("enums must be strings")
	}

	*e = EnumTest(str)
	if !e.IsValid() {
		return fmt.Errorf("%s is not a valid EnumTest", str)
	}
	return nil
}

func (e EnumTest) MarshalGQL(w io.Writer) {
	fmt.Fprint(w, strconv.Quote(e.String()))
}

type T interface {
	IsT()
}
