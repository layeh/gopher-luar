package luar

import (
	"testing"

	"github.com/yuin/gopher-lua"
)

func Test_struct(t *testing.T) {
	L := lua.NewState()
	defer L.Close()

	tim := &StructTestPerson{
		Name: "Tim",
		Age:  30,
	}

	john := StructTestPerson{
		Name: "John",
		Age:  40,
	}

	L.SetGlobal("user1", New(L, tim))
	L.SetGlobal("user2", New(L, john))

	testReturn(t, L, `return user1.Name`, "Tim")
	testReturn(t, L, `return user1.Age`, "30")
	testReturn(t, L, `return user1:Hello()`, "Hello, Tim")

	testReturn(t, L, `return user2.Name`, "John")
	testReturn(t, L, `return user2.Age`, "40")
	testReturn(t, L, `local hello = user2.Hello; return hello(user2)`, "Hello, John")
}

func Test_struct_tostring(t *testing.T) {
	L := lua.NewState()
	defer L.Close()

	p1 := StructTestPerson{
		Name: "Tim",
		Age:  99,
	}
	p2 := StructTestPerson{
		Name: "John",
		Age:  2,
	}

	L.SetGlobal("p1", New(L, &p1))
	L.SetGlobal("p2", New(L, &p2))

	testReturn(t, L, `return tostring(p1)`, `Tim (99)`)
	testReturn(t, L, `return tostring(p2)`, `John (2)`)
}

func Test_struct_pointers(t *testing.T) {
	L := lua.NewState()
	defer L.Close()

	p1 := StructTestPerson{
		Name: "Tim",
	}
	p2 := StructTestPerson{
		Name: "John",
	}

	L.SetGlobal("p1", New(L, &p1))
	L.SetGlobal("p1_alias", New(L, &p1))
	L.SetGlobal("p2", New(L, &p2))

	testReturn(t, L, `return -p1 == -p1`, "true")
	testReturn(t, L, `return -p1 == -p1_alias`, "true")
	testReturn(t, L, `return p1 == p1`, "true")
	testReturn(t, L, `return p1 == p1_alias`, "true")
	testReturn(t, L, `return p1 == p2`, "false")
}

func Test_struct_lstate(t *testing.T) {
	L := lua.NewState()
	defer L.Close()

	p := StructTestPerson{
		Name: "Tim",
	}

	L.SetGlobal("p", New(L, &p))

	testReturn(t, L, `return p:AddNumbers(1, 2, 3, 4, 5)`, "Tim counts: 15")
}

type StructTestHidden struct {
	Name   string `luar:"name"`
	Name2  string `luar:"Name"`
	Str    string
	Hidden bool `luar:"-"`
}

func Test_struct_hiddenfields(t *testing.T) {
	L := lua.NewState()
	defer L.Close()

	a := &StructTestHidden{
		Name:   "tim",
		Name2:  "bob",
		Str:    "asd123",
		Hidden: true,
	}

	L.SetGlobal("a", New(L, a))

	testReturn(t, L, `return a.name`, "tim")
	testReturn(t, L, `return a.Name`, "bob")
	testReturn(t, L, `return a.str`, "asd123")
	testReturn(t, L, `return a.Str`, "asd123")
	testReturn(t, L, `return a.Hidden`, "nil")
	testReturn(t, L, `return a.hidden`, "nil")
}

func Test_struct_method(t *testing.T) {
	L := lua.NewState()
	defer L.Close()

	p := StructTestPerson{
		Name: "Tim",
		Age:  66,
	}

	L.SetGlobal("p", New(L, &p))

	testReturn(t, L, `return p:hello()`, "Hello, Tim")
	testReturn(t, L, `return p.age`, "66")
}

type NestedPointer struct {
	B NestedPointerChild
}

type NestedPointerChild struct {
}

func (*NestedPointerChild) Test() string {
	return "Pointer test"
}

func Test_struct_nestedptrmethod(t *testing.T) {
	L := lua.NewState()
	defer L.Close()

	a := NestedPointer{}
	L.SetGlobal("a", New(L, &a))

	testReturn(t, L, `return a.b:Test()`, "Pointer test")
}

type TestStructEmbeddedType struct {
	TestStructEmbeddedTypeString
}

type TestStructEmbeddedTypeString string

func Test_struct_embeddedtype(t *testing.T) {
	L := lua.NewState()
	defer L.Close()

	a := TestStructEmbeddedType{
		TestStructEmbeddedTypeString: "hello",
	}

	L.SetGlobal("a", New(L, &a))

	testReturn(t, L, `a.TestStructEmbeddedTypeString = "world"`)

	if val := a.TestStructEmbeddedTypeString; val != "world" {
		t.Fatalf("expecting %s, got %s", "world", val)
	}
}

type TestStructEmbedded struct {
	StructTestPerson
	P  StructTestPerson
	P2 StructTestPerson `luar:"other"`
}

func Test_struct_embedded(t *testing.T) {
	L := lua.NewState()
	defer L.Close()

	e := &TestStructEmbedded{}
	L.SetGlobal("e", New(L, e))

	testReturn(
		t,
		L,
		`
		e.StructTestPerson = {
			Name = "Bill",
			Age = 33
		}
		e.P = {
			Name = "Tim",
			Age = 94,
			Friend = {
				Name = "Bob",
				Age = 77
			}
		}
		e.other = {
			Name = "Dale",
			Age = 26
		}
		`,
	)

	{
		expected := StructTestPerson{
			Name: "Bill",
			Age:  33,
		}
		if e.StructTestPerson != expected {
			t.Fatalf("expected %#v, got %#v", expected, e.StructTestPerson)
		}
	}

	{
		expected := StructTestPerson{
			Name: "Bob",
			Age:  77,
		}
		if *(e.P.Friend) != expected {
			t.Fatalf("expected %#v, got %#v", expected, *e.P.Friend)
		}
	}

	{
		expected := StructTestPerson{
			Name: "Dale",
			Age:  26,
		}
		if e.P2 != expected {
			t.Fatalf("expected %#v, got %#v", expected, e.P2)
		}
	}
}

type TestPointerReplaceHidden struct {
	A string `luar:"q"`
	B int    `luar:"other"`
	C int    `luar:"-"`
}

func Test_struct_pointerreplacehidden(t *testing.T) {
	L := lua.NewState()
	defer L.Close()

	e := &TestPointerReplaceHidden{}
	L.SetGlobal("e", New(L, e))

	testReturn(
		t,
		L,
		`
		_ = e ^ {
			q = "Cat",
			other = 675
		}
		`,
	)

	expected := TestPointerReplaceHidden{
		A: "Cat",
		B: 675,
	}

	if *e != expected {
		t.Fatalf("expected %v, got %v", expected, *e)
	}

	testError(
		t,
		L,
		`
		_ = e ^ {
			C = 333
		}
		`,
		`type luar.TestPointerReplaceHidden has no field C`,
	)
}

func Test_struct_ptreq(t *testing.T) {
	L := lua.NewState()
	defer L.Close()

	p1 := StructTestPerson{
		Name: "Tim",
	}

	p2 := StructTestPerson{
		Name: "Tim",
	}

	L.SetGlobal("p1", New(L, &p1))
	L.SetGlobal("p2", New(L, &p2))

	if &p1 == &p2 {
		t.Fatal("expected structs to be unequal")
	}
	testReturn(t, L, `return p1 == p2`, "false")
}

type Test_nested_child1 struct {
	Name string
}

type Test_nested_child2 struct {
	Name string
}

type Test_nested_parent struct {
	Test_nested_child1
	Test_nested_child2
}

func Test_ambiguous_field(t *testing.T) {
	L := lua.NewState()
	defer L.Close()

	n := &Test_nested_parent{}

	L.SetGlobal("c", New(L, n))
	testError(t, L, `
		c.Name = "Tim"
	`, "unknown field Name")

	if n.Test_nested_child1.Name != "" {
		t.Errorf("expected Test_nested_child1.Name to be empty")
	}
	if n.Test_nested_child2.Name != "" {
		t.Errorf("expected Test_nested_child2.Name to be empty")
	}
}

type Test_nested_parent2 struct {
	Test_nested_child1
	Test_nested_child2
	Name string
}

func Test_ambiguous_field2(t *testing.T) {
	L := lua.NewState()
	defer L.Close()

	n := &Test_nested_parent2{}

	L.SetGlobal("c", New(L, n))
	testReturn(t, L, `
		c.Name = "Tim"
		return c.Name
	`, "Tim")

	if n.Name != "Tim" {
		t.Errorf("expected Name to be set to `Tim`")
	}
	if n.Test_nested_child1.Name != "" {
		t.Errorf("expected Test_nested_child1.Name to be empty")
	}
	if n.Test_nested_child2.Name != "" {
		t.Errorf("expected Test_nested_child2.Name to be empty")
	}
}

type test_unexport_anonymous_child struct {
	Name string
}

type Test_struct_unexport_anonymous_parent struct {
	test_unexport_anonymous_child
}

func Test_struct_unexport_anonymous_field(t *testing.T) {
	L := lua.NewState()
	defer L.Close()

	n := &Test_struct_unexport_anonymous_parent{}

	L.SetGlobal("c", New(L, n))
	testReturn(t, L, `
		c.Name = "Tim"
		return c.Name
	`, "Tim")

	if n.Name != "Tim" {
		t.Errorf("expected Name to be set to `Tim`")
	}
}
