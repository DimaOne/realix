package repo

import (
	"math/big"
	"sync"
)

// Repo represents simple repository.
type Repo struct {
	ints sync.Map
}

// New returns new repository.
func New() *Repo {
	return &Repo{}
}

// CheckOrStore return true if represents, false if saved.
func (r *Repo) CheckOrStore(i *big.Int) bool {
	// take hash?
	_, ok := r.ints.LoadOrStore(i.String(), struct{}{})

	return ok
}
