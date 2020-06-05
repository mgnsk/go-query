package query

import (
	"context"
	"sync"
	"sync/atomic"
)

// Expr evaluates to a boolean.
type Expr interface {
	Eval(context.Context) bool
}

// Assertion is a function expression.
type Assertion func(context.Context) bool

// Eval the assertion.
func (a Assertion) Eval(ctx context.Context) bool {
	return a(ctx)
}

// Statement is a list of expressions.
type Statement []Expr

// Eval the raw statement.
func (s Statement) Eval(ctx context.Context) bool {
	if len(s) != 1 {
		panic("invalid number of statements")
	}
	return s[0].Eval(ctx)
}

// AND statement evaluates expressions using logical conjunction.
type AND Statement

// Eval the AND statement.
func (s AND) Eval(ctx context.Context) bool {
	for _, exp := range s {
		if !exp.Eval(ctx) {
			return false
		}
	}
	return true
}

// Race the AND statement expressions.
func (s AND) Race() Expr {
	return Assertion(func(ctx context.Context) bool {
		ctx, cancel := context.WithCancel(ctx)
		defer cancel()
		return !race(ctx, Statement(s), func(ctx context.Context, exp Expr) bool {
			if !exp.Eval(ctx) {
				cancel()
				return true
			}
			return false
		})
	})
}

// OR statement evaluates expressions using logical disjunction.
type OR Statement

// Eval the OR statement.
func (s OR) Eval(ctx context.Context) bool {
	for _, exp := range s {
		if exp.Eval(ctx) {
			return true
		}
	}
	return false
}

// Race the OR statement.
func (s OR) Race() Expr {
	return Assertion(func(ctx context.Context) bool {
		ctx, cancel := context.WithCancel(ctx)
		defer cancel()
		return race(ctx, Statement(s), func(ctx context.Context, exp Expr) bool {
			if exp.Eval(ctx) {
				cancel()
				return true
			}
			return false
		})
	})
}

// NOT statement inverts the evaluated expression.
type NOT Statement

// Eval the NOT statement.
func (s NOT) Eval(ctx context.Context) bool {
	return !Statement(s).Eval(ctx)
}

// IF statement builds a condition.
type IF struct {
	test Expr
	then Expr
	els  Expr
}

func (s IF) Eval(ctx context.Context) bool {
	if s.test.Eval(ctx) {
		return s.then.Eval(ctx)
	}
	return s.els.Eval(ctx)
}

// race until any expression in q returns true, then return true, trying to cancel other evaluators,
// otherwise return false.
func race(ctx context.Context, q Statement, eval func(context.Context, Expr) bool) bool {
	var wg sync.WaitGroup
	isTrue := uint64(0)

	for _, exp := range q {
		exp := exp
		wg.Add(1)
		go func() {
			defer wg.Done()
			select {
			case <-ctx.Done():
				// Context canceled by either race winner or parent.
			default:
				// Checks not synced, multiple checks may run after
				// the Statement has Evaluated to a stop condition.
				// but it's good enough for cancellation.
				if atomic.LoadUint64(&isTrue) == 0 && eval(ctx, exp) {
					atomic.StoreUint64(&isTrue, 1)
				}
			}
		}()
	}

	wg.Wait()
	if isTrue == 1 {
		return true
	}

	return false
}
