package model

type InverseType struct {
	Id         string      `json:"id" dql:"uid"`
	Name       string      `json:"name" dql:"InverseType.name"`
	CyclicType *CyclicType `json:"cyclicType" dql:"InverseType.cyclicType"`
}
