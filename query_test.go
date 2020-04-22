package query

import (
	"context"
	"runtime"
	"testing"
)

func newQuery() Expression {
	return NOT{
		OR{
			Assertion(func(_ context.Context) bool {
				// time.Sleep(1 * time.Millisecond)
				return false
			}),
			Assertion(func(_ context.Context) bool {
				// time.Sleep(1 * time.Millisecond)
				return false
			}),
			AND{
				NOT{
					Assertion(func(_ context.Context) bool {
						// time.Sleep(1 * time.Millisecond)
						return false
					}),
				},
				Assertion(func(_ context.Context) bool {
					// time.Sleep(1 * time.Millisecond)
					return true
				}),
				IF(
					AND{
						Assertion(func(_ context.Context) bool {
							// time.Sleep(1 * time.Millisecond)
							return false
						}),
						Assertion(func(_ context.Context) bool {
							// time.Sleep(1 * time.Millisecond)
							return true
						}),
					}.Race(),
					Assertion(func(_ context.Context) bool {
						// time.Sleep(1 * time.Millisecond)
						return false
					}),
					Assertion(func(_ context.Context) bool {
						// time.Sleep(1 * time.Millisecond)
						return true
					}),
				),
			}.Race(),
		}.Race(),
	}
}

func TestQuery(t *testing.T) {
	s := newQuery()
	if s.Eval(context.TODO()) != false {
		t.Fatal("expected false")
	}
}

func BenchmarkQuery(b *testing.B) {
	b.ReportAllocs()
	s := newQuery()
	result := false
	for i := 0; i < b.N; i++ {
		result = s.Eval(context.TODO())
	}
	runtime.KeepAlive(result)
}
