package pds

import "slices"

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

func (m Menu) Has(n int) bool {
	return slices.ContainsFunc(m.Pizzas, func(p Pizza) bool {
		return p.Number == n
	})
}

func (m Menu) Pizza(n int) (Pizza, error) {
	for _, p := range m.Pizzas {
		if p.Number == n {
			return p, nil
		}
	}

	return Pizza{}, ErrUnknownPizza

}

type Pizza struct {
	Number int    `json:"nr"`
	Name   string `json:"name"`
	Price  int    `json:"price"`
}
