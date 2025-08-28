package pds

import (
	"context"
	"slices"

	"go.temporal.io/sdk/workflow"
)

func GetMenu(ctx workflow.Context) (Menu, error) {
	a := &Activities{}

	return a.RetrieveMenu(context.Background())
}

func NewMenu() Menu {
	return Menu{
		Pizzas: []Pizza{
			{Number: 1, Name: "Kebab Pizza", Price: 75},
			{Number: 2, Name: "Vesuvio", Price: 60},
			{Number: 3, Name: "Hawaiian", Price: 65},
			{Number: 4, Name: "Margherita", Price: 55},
			{Number: 5, Name: "Capricciosa", Price: 60},
		},
	}
}

type Menu struct {
	Pizzas []Pizza
}

func (m Menu) HasPizza(n int) bool {
	return slices.ContainsFunc(m.Pizzas, func(p Pizza) bool {
		return p.Number == n
	})
}

type Pizza struct {
	Number int    `json:"nr"`
	Name   string `json:"name"`
	Price  int    `json:"price"`
}
