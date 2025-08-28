package pds

import "time"

type Customer struct {
	Name     string        `json:"name"`
	Address  string        `json:"addr"`
	Delivery time.Duration `json:"delivery"`
}

func (c Customer) Unknown() bool {
	return c.Name == ""
}
