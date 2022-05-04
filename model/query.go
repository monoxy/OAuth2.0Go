package model

import (
	"net/url"
)

type ClientQuery struct {
	Vals url.Values
}

func NewClientQuery(vals url.Values) *ClientQuery {
	return &ClientQuery{Vals: vals}
}

func (c *ClientQuery) Get(key string) string {
	return c.Vals.Get(key)
}
