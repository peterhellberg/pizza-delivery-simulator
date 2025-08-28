package pds

import (
	"context"
	"errors"
	"math/rand"
	"time"
)

type Activities struct{}

func (a *Activities) RetrieveMenu(ctx context.Context) (Menu, error) {
	// Pretend we call an external API
	time.Sleep(10 * time.Millisecond)

	// Simulate outcomes
	switch rand.Intn(5) {
	case 0:
		return Menu{}, ErrUnavailableMenu
	default:
		return NewMenu(), nil
	}
}

func (a *Activities) LookupCustomer(ctx context.Context, name string) (Customer, error) {
	// Pretend we call an external API
	time.Sleep(10 * time.Millisecond)

	// Simulate outcomes
	switch rand.Intn(5) {
	case 0:
		return Customer{}, errors.New("network timeout") // retryable
	default:
		return newCustomer(name), nil
	}
}

func newCustomer(name string) Customer {
	switch name {
	case "Peter", "Peter Hellberg":
		return Customer{
			Name:     "Peter Hellberg",
			Address:  "Mosstenabacken 1, 12432 Bandhagen",
			Delivery: 5 * time.Minute,
		}
	case "John", "John Doe":
		return Customer{
			Name:     "John Doe",
			Address:  "Fakestreet 0, 12345 Nowhere",
			Delivery: 1 * time.Hour,
		}
	default:
		return Customer{}
	}
}
