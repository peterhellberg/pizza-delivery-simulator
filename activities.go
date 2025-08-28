package pds

import "context"

type Activities struct{}

func (a *Activities) RetrieveMenu(ctx context.Context) (Menu, error) {
	return NewMenu(), nil
}
