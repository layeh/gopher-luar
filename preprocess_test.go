package luar

import (
	"reflect"
	"testing"

	lua "github.com/yuin/gopher-lua"
)

type Foo struct{ A Bar } // test field

type Bar struct{ A int } // no-op

func (b Bar) Baz() Baz { return Baz{A: struct{ A FooFoo }{}} } // test method

type Baz struct{ A struct{ A FooFoo } } // test nested

type FooFoo struct {
	a FooBar
	b FooBaz

	T1 int
	T2 string
	T3 map[string]int
	T4 []int
} // b is hidden as field, visible as method

func (f FooFoo) Baz(bar BarBaz) FooBaz { return f.b }

type FooBar struct{ A int }

type FooBaz struct{ A [][][]*[][]*[]*[][]*BarFoo } // ungodly depth achieved!

type BarFoo struct{ A BarBar }

type BarBar int

type BarBaz interface{ Baz() FooBaz }

func Test_preprocess(t *testing.T) {
	t.Run("enabled", testPreprocess([]reflect.Type{
		reflect.TypeOf(Foo{}),
		reflect.TypeOf(Bar{}),
		reflect.TypeOf(Baz{}),
		reflect.TypeOf(struct{ A FooFoo }{}), // nested
		reflect.TypeOf(FooFoo{}),
		// reflect.TypeOf(FooBar{}), // hidden, skipped
		reflect.TypeOf(FooBaz{}), // hidden as field, visible as method
		// reflect.TypeOf(BarBar(0)), // not of interest
		reflect.TypeOf([]int{}),
		reflect.TypeOf(map[string]int{}),

		// oh no...
		reflect.TypeOf([][][]*[][]*[]*[][]*BarFoo{}),
		reflect.TypeOf([][]*[][]*[]*[][]*BarFoo{}),
		reflect.TypeOf([]*[][]*[]*[][]*BarFoo{}),
		reflect.TypeOf(&[][]*[]*[][]*BarFoo{}),
		reflect.TypeOf([][]*[]*[][]*BarFoo{}),
		reflect.TypeOf([]*[]*[][]*BarFoo{}),
		reflect.TypeOf(&[]*[][]*BarFoo{}),
		reflect.TypeOf([]*[][]*BarFoo{}),
		reflect.TypeOf(&[][]*BarFoo{}),
		reflect.TypeOf([][]*BarFoo{}),
		reflect.TypeOf([]*BarFoo{}),
		reflect.TypeOf(&BarFoo{}),
		reflect.TypeOf(BarFoo{}),
	}, true, New))
	t.Run("disabled", testPreprocess([]reflect.Type{
		reflect.TypeOf(Foo{}), // no preprocessing -> only the original type
	}, false, New))

	t.Run("type_enabled", testPreprocess([]reflect.Type{reflect.TypeOf(Foo{})}, true, NewType))
	t.Run("type_disabled", testPreprocess([]reflect.Type{reflect.TypeOf(Foo{})}, false, NewType))
}

type newfn = func(L *lua.LState, value interface{}) lua.LValue

func testPreprocess(expected []reflect.Type, preprocess bool, newfn newfn) func(t *testing.T) {
	return func(t *testing.T) {
		L := lua.NewState()
		defer L.Close()

		got := map[string]bool{}

		GetConfig(L).PreprocessMetatables = preprocess
		GetConfig(L).Metatable = func(L *lua.LState, t reflect.Type, mt *lua.LTable, constructor bool) *lua.LTable {
			got[t.String()] = true
			return mt
		}

		newfn(L, Foo{})

		expectedM := map[string]bool{}
		for _, v := range expected {
			expectedM[v.String()] = true
		}
		for name := range expectedM {
			if _, ok := got[name]; ok {
				t.Logf("processed %s", name)
			} else {
				t.Errorf("expected %s to be processed", name)
			}
		}
		for name := range got {
			if _, ok := expectedM[name]; !ok {
				t.Errorf("did not expect %s to be processed", name)
			}
		}
	}
}
