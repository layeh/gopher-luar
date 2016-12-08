package luar

import (
	"fmt"
	"strconv"
	"strings"
	"testing"

	"github.com/yuin/gopher-lua"
)

type Person struct {
	Name       string
	Age        int
	Friend     *Person
	LastAddSum int
}

func (p Person) Hello() string {
	return "Hello, " + p.Name
}

func (p Person) String() string {
	return p.Name + " (" + strconv.Itoa(p.Age) + ")"
}

func (p *Person) AddNumbers(L *LState) int {
	sum := 0
	for i := L.GetTop(); i >= 1; i-- {
		sum += L.CheckInt(i)
	}
	L.Push(lua.LString(p.Name + " counts: " + strconv.Itoa(sum)))
	p.LastAddSum = sum
	return 1
}

func (p *Person) IncreaseAge() {
	p.Age++
}

type Family struct {
	Mother   Person
	Father   Person
	Children []Person
}

func ExampleStructUsage() {
	const code = `
	print(user1.Name)
	print(user1.Age)
	print(user1:Hello())

	print(user2.Name)
	print(user2.Age)
	hello = user2.Hello
	print(hello(user2))
	`

	L := lua.NewState()
	defer L.Close()

	tim := &Person{
		Name: "Tim",
		Age:  30,
	}

	john := Person{
		Name: "John",
		Age:  40,
	}

	L.SetGlobal("user1", New(L, tim))
	L.SetGlobal("user2", New(L, john))

	if err := L.DoString(code); err != nil {
		panic(err)
	}

	// Output:
	// Tim
	// 30
	// Hello, Tim
	// John
	// 40
	// Hello, John
}

func ExampleMapAndSlice() {
	const code = `
	for i = 1, #things do
		print(things[i])
	end
	things[1] = "cookie"

	print()

	print(thangs.ABC)
	print(thangs.DEF)
	print(thangs.GHI)
	thangs.GHI = 789
	thangs.ABC = nil
	`

	L := lua.NewState()
	defer L.Close()

	things := []string{
		"cake",
		"wallet",
		"calendar",
		"phone",
		"speaker",
	}

	thangs := map[string]int{
		"ABC": 123,
		"DEF": 456,
	}

	L.SetGlobal("things", New(L, things))
	L.SetGlobal("thangs", New(L, thangs))

	if err := L.DoString(code); err != nil {
		panic(err)
	}

	fmt.Println()
	fmt.Println(things[0])
	fmt.Println(thangs["GHI"])
	_, ok := thangs["ABC"]
	fmt.Println(ok)

	// Output:
	// cake
	// wallet
	// calendar
	// phone
	// speaker
	//
	// 123
	// 456
	// nil
	//
	// cookie
	// 789
	// false
}

func ExampleStructConstructorAndMap() {
	const code = `
	user2 = Person()
	user2.Name = "John"
	user2.Friend = user1
	print(user2.Name)
	print(user2.Friend.Name)

	everyone = People()
	everyone["tim"] = user1
	everyone["john"] = user2
	`

	L := lua.NewState()
	defer L.Close()

	tim := &Person{
		Name: "Tim",
	}

	L.SetGlobal("user1", New(L, tim))
	L.SetGlobal("Person", NewType(L, Person{}))
	L.SetGlobal("People", NewType(L, map[string]*Person{}))

	if err := L.DoString(code); err != nil {
		panic(err)
	}

	everyone := L.GetGlobal("everyone").(*lua.LUserData).Value.(*reflectedInterface).Interface.(map[string]*Person)
	fmt.Println(len(everyone))

	// Output:
	// John
	// Tim
	// 2
}

func ExampleGoFunc() {
	const code = `
	print(getHello(person))
	`

	L := lua.NewState()
	defer L.Close()

	tim := &Person{
		Name: "Tim",
	}

	fn := func(p *Person) string {
		return "Hello, " + p.Name
	}

	L.SetGlobal("person", New(L, tim))
	L.SetGlobal("getHello", New(L, fn))

	if err := L.DoString(code); err != nil {
		panic(err)
	}

	// Output:
	// Hello, Tim
}

func ExampleChan() {
	const code = `
	print(ch:receive())
	ch:send("John")
	print(ch:receive())
	`

	L := lua.NewState()
	defer L.Close()

	ch := make(chan string)
	go func() {
		ch <- "Tim"
		name, ok := <-ch
		fmt.Printf("%s\t%v\n", name, ok)
		close(ch)
	}()

	L.SetGlobal("ch", New(L, ch))

	if err := L.DoString(code); err != nil {
		panic(err)
	}

	// Output:
	// Tim	true
	// John	true
	// nil	false
}

func ExampleMap() {
	const code = `
	local sorted = {}
	for k, v in countries() do
		table.insert(sorted, v)
	end
	table.sort(sorted)
	for i = 1, #sorted do
		print(sorted[i])
	end
	`

	L := lua.NewState()
	defer L.Close()

	countries := map[string]string{
		"JP": "Japan",
		"CA": "Canada",
		"FR": "France",
	}

	L.SetGlobal("countries", New(L, countries))

	if err := L.DoString(code); err != nil {
		panic(err)
	}

	// Output:
	// Canada
	// France
	// Japan
}

func ExampleFuncVariadic() {
	const code = `
	fn("a", 1, 2, 3)
	fn("b")
	fn("c", 4)
	`

	L := lua.NewState()
	defer L.Close()

	fn := func(str string, extra ...int) {
		fmt.Printf("%s\n", str)
		for _, x := range extra {
			fmt.Printf("%d\n", x)
		}
	}

	L.SetGlobal("fn", New(L, fn))

	if err := L.DoString(code); err != nil {
		panic(err)
	}

	// Output:
	// a
	// 1
	// 2
	// 3
	// b
	// c
	// 4
}

func ExampleLuaFuncVariadic() {
	const code = `
	for _, x in ipairs(fn(1, 2, 3)) do
		print(x)
	end
	for _, x in ipairs(fn()) do
		print(x)
	end
	for _, x in ipairs(fn(4)) do
		print(x)
	end
	`

	L := lua.NewState()
	defer L.Close()

	fn := func(x ...float64) *lua.LTable {
		tbl := L.NewTable()
		for i := len(x) - 1; i >= 0; i-- {
			tbl.Insert(len(x)-i, lua.LNumber(x[i]))
		}
		return tbl
	}

	L.SetGlobal("fn", New(L, fn))

	if err := L.DoString(code); err != nil {
		panic(err)
	}

	// Output:
	// 3
	// 2
	// 1
	// 4
}

func ExampleSlice() {
	const code = `
	print(#items)
	print(items:capacity())
	items = items:append("hello", "world")
	print(#items)
	print(items:capacity())
	print(items[1])
	print(items[2])
	`

	L := lua.NewState()
	defer L.Close()

	items := make([]string, 0, 10)

	L.SetGlobal("items", New(L, items))

	if err := L.DoString(code); err != nil {
		panic(err)
	}

	// Output:
	// 0
	// 10
	// 2
	// 10
	// hello
	// world
}

func ExampleSliceCapacity() {
	const code = `
	ints = newInts(1)
	print(#ints, ints:capacity())

	ints = newInts(0, 10)
	print(#ints, ints:capacity())
	`

	L := lua.NewState()
	defer L.Close()

	type ints []int

	L.SetGlobal("newInts", NewType(L, ints{}))

	if err := L.DoString(code); err != nil {
		panic(err)
	}

	// Output:
	// 1	1
	// 0	10
}

func ExampleStructPtrEquality() {
	const code = `
	print(-p1 == -p1)
	print(-p1 == -p1_alias)
	print(p1 == p1)
	print(p1 == p1_alias)
	print(p1 == p2)
	`

	L := lua.NewState()
	defer L.Close()

	p1 := Person{
		Name: "Tim",
	}
	p2 := Person{
		Name: "John",
	}

	L.SetGlobal("p1", New(L, &p1))
	L.SetGlobal("p1_alias", New(L, &p1))
	L.SetGlobal("p2", New(L, &p2))

	if err := L.DoString(code); err != nil {
		panic(err)
	}

	// Output:
	// true
	// true
	// true
	// true
	// false
}

func ExampleStructStringer() {
	const code = `
	print(p1)
	print(p2)
	`

	L := lua.NewState()
	defer L.Close()

	p1 := Person{
		Name: "Tim",
		Age:  99,
	}
	p2 := Person{
		Name: "John",
		Age:  2,
	}

	L.SetGlobal("p1", New(L, &p1))
	L.SetGlobal("p2", New(L, &p2))

	if err := L.DoString(code); err != nil {
		panic(err)
	}

	// Output:
	// Tim (99)
	// John (2)
}

func ExamplePtrMethod() {
	const code = `
	print(p:AddNumbers(1, 2, 3, 4, 5))
	`

	L := lua.NewState()
	defer L.Close()

	p := Person{
		Name: "Tim",
	}

	L.SetGlobal("p", New(L, &p))

	if err := L.DoString(code); err != nil {
		panic(err)
	}

	// Output:
	// Tim counts: 15
}

func ExampleStruct() {
	const code = `
	print(p:hello())
	print(p.age)
	`

	L := lua.NewState()
	defer L.Close()

	p := Person{
		Name: "Tim",
		Age:  66,
	}

	L.SetGlobal("p", New(L, &p))

	if err := L.DoString(code); err != nil {
		panic(err)
	}

	// Output:
	// Hello, Tim
	// 66
}

type OneString [1]string

func (o OneString) Log() string {
	return o[0]
}

func ExampleArray() {
	const code = `
	print(#e.V, e.V[1], e.V[2])
	e.V[1] = "World"
	e.V[2] = "Hello"
	print(#e.V, e.V[1], e.V[2])

	print(#arr, arr[1])
	print(arr:log())
	`

	type Elem struct {
		V [2]string
	}

	L := lua.NewState()
	defer L.Close()

	var elem Elem
	elem.V[0] = "Hello"
	elem.V[1] = "World"

	var arr OneString
	arr[0] = "Test"

	L.SetGlobal("e", New(L, &elem))
	L.SetGlobal("arr", New(L, arr))

	if err := L.DoString(code); err != nil {
		panic(err)
	}

	// Output:
	// 2	Hello	World
	// 2	World	Hello
	// 1	Test
	// Test
}

func ExampleLuaFunc() {
	const code = `
	print(fn("tim", 5))
	`

	L := lua.NewState()
	defer L.Close()

	fn := func(name string, count int) []lua.LValue {
		s := make([]lua.LValue, count)
		for i := 0; i < count; i++ {
			s[i] = lua.LString(name)
		}
		return s
	}

	L.SetGlobal("fn", New(L, fn))

	if err := L.DoString(code); err != nil {
		panic(err)
	}

	// Output:
	// tim	tim	tim	tim	tim
}

func ExamplePtrIndirection() {
	const code = `
	print(-ptr)
	`

	L := lua.NewState()
	defer L.Close()

	str := "hello"

	L.SetGlobal("ptr", New(L, &str))

	if err := L.DoString(code); err != nil {
		panic(err)
	}

	// Output:
	// hello
}

func ExamplePtrEquality() {
	const code = `
	print(ptr1 == nil)
	print(ptr2 == nil)
	print(ptr1 == ptr2)
	`

	L := lua.NewState()
	defer L.Close()

	var ptr1 *string
	str := "hello"

	L.SetGlobal("ptr1", New(L, ptr1))
	L.SetGlobal("ptr2", New(L, &str))

	if err := L.DoString(code); err != nil {
		panic(err)
	}

	// Output:
	// true
	// false
	// false
}

func ExamplePtrAssignment() {
	const code = `
	print(-str)
	print(str ^ "world")
	print(-str)
	`

	L := lua.NewState()
	defer L.Close()

	str := "hello"

	L.SetGlobal("str", New(L, &str))

	if err := L.DoString(code); err != nil {
		panic(err)
	}

	// Output:
	// hello
	// world
	// world
}

type AnonymousFieldsA struct {
	*AnonymousFieldsB
}

type AnonymousFieldsB struct {
	Value *string
	Person
}

func ExampleAnonymousFields() {
	const code = `
	print(a.Value == nil)
	a.Value = str_ptr()
	_ = a.Value ^ "hello"
	print(a.Value == nil)
	print(-a.Value)
	print(a.Name)
	`

	L := lua.NewState()
	defer L.Close()

	a := AnonymousFieldsA{
		AnonymousFieldsB: &AnonymousFieldsB{
			Person: Person{
				Name: "Tim",
			},
		},
	}

	L.SetGlobal("a", New(L, a))
	L.SetGlobal("str_ptr", NewType(L, ""))

	if err := L.DoString(code); err != nil {
		panic(err)
	}

	// Output:
	// true
	// false
	// hello
	// Tim
}

func ExampleEmptyFunc() {
	const code = `
	print(fn == nil)
	`

	L := lua.NewState()
	defer L.Close()

	var fn func()

	L.SetGlobal("fn", New(L, fn))

	if err := L.DoString(code); err != nil {
		panic(err)
	}

	// Output:
	// true
}

func ExampleFuncArray() {
	const code = `
	fn(arr)
	`

	L := lua.NewState()
	defer L.Close()

	arr := [3]int{1, 2, 3}
	fn := func(val [3]int) {
		fmt.Printf("%d %d %d\n", val[0], val[1], val[2])
	}

	L.SetGlobal("fn", New(L, fn))
	L.SetGlobal("arr", New(L, arr))

	if err := L.DoString(code); err != nil {
		panic(err)
	}

	// Output:
	// 1 2 3
}

func ExampleComplex() {
	const code = `
	b = a
	`

	L := lua.NewState()
	defer L.Close()

	a := complex(float64(1), float64(2))

	L.SetGlobal("a", New(L, a))

	if err := L.DoString(code); err != nil {
		panic(err)
	}
	b := L.GetGlobal("b").(*lua.LUserData).Value.(*reflectedInterface).Interface.(complex128)
	fmt.Println(a == b)

	// Output:
	// true
}

type ChanAlias chan string

func (ChanAlias) Test() string {
	return `I'm a "chan string" alias`
}

func (ChanAlias) hidden() {
}

type SliceAlias []string

func (s SliceAlias) Len() int {
	return len(s)
}

func (s *SliceAlias) Append(v string) {
	*s = append(*s, v)
}

type MapAlias map[string]int

func (m MapAlias) Y() int {
	return len(m)
}

func ExampleTypeAlias() {
	const code = `
	print(a:Test())
	local len1 = b:Len()
	b:Append("!")
	print(len1, b:len())
	print(c.x, c:y())
	`

	L := lua.NewState()
	defer L.Close()

	a := make(ChanAlias)
	var b SliceAlias = []string{"Hello", "world"}
	c := MapAlias{
		"x": 15,
	}

	L.SetGlobal("a", New(L, a))
	L.SetGlobal("b", New(L, &b))
	L.SetGlobal("c", New(L, c))

	if err := L.DoString(code); err != nil {
		panic(err)
	}

	// Output:
	// I'm a "chan string" alias
	// 2	3
	// 15	1
}

type StructPtrFuncB struct {
}

func (*StructPtrFuncB) Test() string {
	return "Pointer test"
}

type StructPtrFuncA struct {
	B StructPtrFuncB
}

func ExampleStructPtrFunc() {
	const code = `
	print(a.b:Test())
	`

	L := lua.NewState()
	defer L.Close()

	a := StructPtrFuncA{}
	L.SetGlobal("a", New(L, &a))

	if err := L.DoString(code); err != nil {
		panic(err)
	}
	fmt.Println(a.B.Test())

	// Output:
	// Pointer test
	// Pointer test
}

type HiddenFieldNamesA struct {
	Name   string `luar:"name"`
	Name2  string `luar:"Name"`
	Str    string
	Hidden bool `luar:"-"`
}

func ExampleHiddenFieldNames() {
	const code = `
	print(a.name)
	print(a.Name)
	print(a.str)
	print(a.Str)
	print(a.Hidden)
	print(a.hidden)
	`

	L := lua.NewState()
	defer L.Close()

	a := &HiddenFieldNamesA{
		Name:   "tim",
		Name2:  "bob",
		Str:    "asd123",
		Hidden: true,
	}

	L.SetGlobal("a", New(L, a))

	if err := L.DoString(code); err != nil {
		panic(err)
	}

	// Output:
	// tim
	// bob
	// asd123
	// asd123
	// nil
	// nil
}

func ExampleStructPtrAssignment() {
	const code = `
	print(a.Name)
	_ = a ^ -b
	print(a.Name)
	`

	L := lua.NewState()
	defer L.Close()

	a := &Person{
		Name: "tim",
	}
	b := &Person{
		Name: "bob",
	}

	L.SetGlobal("a", New(L, a))
	L.SetGlobal("b", New(L, b))

	if err := L.DoString(code); err != nil {
		panic(err)
	}

	// Output:
	// tim
	// bob
}

type PtrNonPtrChanMethodsA chan string

func (*PtrNonPtrChanMethodsA) Test() string {
	return "Test"
}

func (PtrNonPtrChanMethodsA) Test2() string {
	return "Test2"
}

func ExamplePtrNonPtrChanMethods() {
	const code = `
	print(b:Test())
	print(b:Test2())
	`

	L := lua.NewState()
	defer L.Close()

	a := make(PtrNonPtrChanMethodsA)
	b := &a

	fmt.Println(b.Test())
	fmt.Println(b.Test2())

	L.SetGlobal("b", New(L, b))

	if err := L.DoString(code); err != nil {
		panic(err)
	}

	// Output:
	// Test
	// Test2
	// Test
	// Test2
}

type StructFieldA string

type StructFieldB struct {
	StructFieldA
}

func ExampleStructField() {
	const code = `
	a.StructFieldA = "world"
	`

	L := lua.NewState()
	defer L.Close()

	a := StructFieldB{}
	a.StructFieldA = "hello"
	fmt.Println(a.StructFieldA)

	L.SetGlobal("a", New(L, &a))

	if err := L.DoString(code); err != nil {
		panic(err)
	}
	fmt.Println(a.StructFieldA)

	// Output:
	// hello
	// world
}

type StructBlacklistA struct {
}

func (*StructBlacklistA) Public() string {
	return "You can call me"
}

func (StructBlacklistA) Private() string {
	return "Should not be able to call me"
}

type StructBlacklistB struct {
	*StructBlacklistA
}

func ExampleStructBlacklist() {
	const code = `
	print(b:public())
	print(b.StructBlacklistA:public())
	pcall(function()
		print(b:private())
	end)
	pcall(function()
		print(b.StructBlacklistA:private())
	end)
	pcall(function()
		print(b:Private())
	end)
	pcall(function()
		print(b.StructBlacklistA:Private())
	end)
	pcall(function()
		local a = -b.StructBlacklistA
		print(a:Private())
	end)
	`

	L := lua.NewState()
	defer L.Close()

	b := &StructBlacklistB{
		StructBlacklistA: &StructBlacklistA{},
	}

	mt := MT(L, StructBlacklistB{})
	mt.Blacklist("private", "Private")

	mt = MT(L, StructBlacklistA{})
	mt.Whitelist("public", "Public")

	L.SetGlobal("b", New(L, b))

	if err := L.DoString(code); err != nil {
		panic(err)
	}

	// Output:
	// You can call me
	// You can call me
}

type SliceAssignmentA struct {
	S []string
}

func ExampleSliceAssignment() {
	const code = `
	x.S = {"a", "b", "", 3, true, "c"}
	`

	L := lua.NewState()
	defer L.Close()

	e := &SliceAssignmentA{}
	L.SetGlobal("x", New(L, e))

	if err := L.DoString(code); err != nil {
		panic(err)
	}
	for _, v := range e.S {
		fmt.Println(v)
	}

	// Output:
	// a
	// b
	//
	// 3
	// true
	// c
}

type SliceTableAssignmentA struct {
	S map[string]string
}

func ExampleSliceTableAssignment() {
	const code = `
	x.S = {
		33,
		a = 123,
		b = nil,
		c = "hello",
		d = false
	}
	`

	L := lua.NewState()
	defer L.Close()

	e := &SliceTableAssignmentA{}
	L.SetGlobal("x", New(L, e))

	if err := L.DoString(code); err != nil {
		panic(err)
	}

	fmt.Println(len(e.S))
	fmt.Println(e.S["a"])
	fmt.Println(e.S["b"])
	fmt.Println(e.S["c"])
	fmt.Println(e.S["d"])

	// Output:
	// 3
	// 123
	//
	// hello
	// false
}

type FieldNameResolutionA struct {
	Person
	P  Person
	P2 Person `luar:"other"`
}

func ExampleFieldNameResolution() {
	const code = `
	x.Person = {
		Name = "Bill",
		Age = 33
	}
	x.P = {
		Name = "Tim",
		Age = 94,
		Friend = {
			Name = "Bob",
			Age = 77
		}
	}
	x.other = {
		Name = "Dale",
		Age = 26
	}
	`

	L := lua.NewState()
	defer L.Close()

	e := &FieldNameResolutionA{}
	L.SetGlobal("x", New(L, e))

	if err := L.DoString(code); err != nil {
		panic(err)
	}

	fmt.Println(e.Name)
	fmt.Println(e.Age)
	fmt.Println(e.P.Name)
	fmt.Println(e.P.Age)
	fmt.Println(e.P.Friend.Name)
	fmt.Println(e.P.Friend.Age)
	fmt.Println(e.P2.Name)
	fmt.Println(e.P2.Age)

	// Output:
	// Bill
	// 33
	// Tim
	// 94
	// Bob
	// 77
	// Dale
	// 26
}

type PCallA struct {
	A string `luar:"q"`
	B int    `luar:"other"`
	C int    `luar:"-"`
}

func ExamplePCall() {
	const code = `
	_ = x ^ {
		q = "Cat",
		other = 675
	}
	pcall(function()
		_ = x ^ {
			C = 333
		}
	end)
	`

	L := lua.NewState()
	defer L.Close()

	e := &PCallA{}
	L.SetGlobal("x", New(L, e))

	if err := L.DoString(code); err != nil {
		panic(err)
	}

	fmt.Println(e.A)
	fmt.Println(e.B)
	fmt.Println(e.C)

	// Output:
	// Cat
	// 675
	// 0
}

type LuaFuncDefinitionA struct {
	Fn  func(a string) (string, int)
	Fn2 func(a string, b ...int) string
}

func ExampleLuaFuncDefinition() {
	const code = `
	i = 0
	x.Fn = function(str)
		i = i + 1
		return ">" .. str .. "<", i
	end

	x.Fn2 = function(str, a, b, c)
		if type(a) == "number" and type(b) == "number" and type(c) == "number" then
			return str
		end
		return ""
	end
	`

	L := lua.NewState()
	defer L.Close()

	e := &LuaFuncDefinitionA{}
	L.SetGlobal("x", New(L, e))

	if err := L.DoString(code); err != nil {
		panic(err)
	}

	for ch := 'A'; ch <= 'C'; ch++ {
		str, i := e.Fn(string(ch))
		fmt.Printf("%s %d\n", str, i)
	}

	fmt.Println(e.Fn2("hello", 1, 2))
	fmt.Println(e.Fn2("hello", 1, 2, 3))

	if L.GetTop() != 0 {
		panic("expecting GetTop to return 0, got " + strconv.Itoa(L.GetTop()))
	}

	// Output:
	// >A< 1
	// >B< 2
	// >C< 3
	//
	// hello
}

type LuaFuncPtrA struct {
	F1 *lua.LFunction
}

func ExampleLuaFuncPtr() {
	const code = `
	x.F1 = function(str)
		print("Hello World")
	end
	`

	L := lua.NewState()
	defer L.Close()

	e := &LuaFuncPtrA{}
	L.SetGlobal("x", New(L, e))

	if err := L.DoString(code); err != nil {
		panic(err)
	}

	L.Push(e.F1)
	L.Call(0, 0)

	// Output:
	// Hello World
}

func ExampleSliceAndArrayTypes() {
	const code = `
	for i, x in s() do
		print(i, x)
	end
	for i, x in e() do
		print(i, x)
	end
	for i, x in a() do
		print(i, x)
	end
	for i, x in ap() do
		print(i, x)
	end
	`

	L := lua.NewState()
	defer L.Close()

	s := []string{
		"hello",
		"there",
		"tim",
	}

	e := []string{}

	a := [...]string{"x", "y"}

	L.SetGlobal("s", New(L, s))
	L.SetGlobal("e", New(L, e))
	L.SetGlobal("a", New(L, a))
	L.SetGlobal("ap", New(L, &a))

	if err := L.DoString(code); err != nil {
		panic(err)
	}

	// Output:
	// 1	hello
	// 2	there
	// 3	tim
	// 1	x
	// 2	y
	// 1	x
	// 2	y
}

type StructArrayAndSliceA string

func (s *StructArrayAndSliceA) ToUpper() {
	*s = StructArrayAndSliceA(strings.ToUpper(string(*s)))
}

func ExampleStructArrayAndSlice() {
	const code = `
	print(a[1]:AddNumbers(1, 2, 3, 4, 5))
	print(s[1]:AddNumbers(1, 2, 3, 4))
	print(s[1].LastAddSum)
	print(p:AddNumbers(1, 2, 3, 4, 5))
	print(p.LastAddSum)

	print(p.Age)
	p:IncreaseAge()
	print(p.Age)

	print(-str)
	str:ToUpper()
	print(-str)
	`

	L := lua.NewState()
	defer L.Close()

	a := [...]Person{
		{Name: "Tim"},
	}
	s := []Person{
		{Name: "Tim", Age: 32},
	}

	str := StructArrayAndSliceA("Hello World")

	L.SetGlobal("a", New(L, &a))
	L.SetGlobal("s", New(L, s))
	L.SetGlobal("p", New(L, s[0]))
	L.SetGlobal("str", New(L, &str))

	if err := L.DoString(code); err != nil {
		panic(err)
	}

	// Output:
	// Tim counts: 15
	// Tim counts: 10
	// 10
	// Tim counts: 15
	// 15
	// 32
	// 33
	// Hello World
	// HELLO WORLD
}

func ExampleLStateFunc() {
	const code = `
	print(sum(1, 2, 3, 4, 5))
	`

	L := lua.NewState()
	defer L.Close()

	sum := func(L *LState) int {
		total := 0
		for i := 1; i <= L.GetTop(); i++ {
			total += L.CheckInt(i)
		}
		L.Push(lua.LNumber(total))
		return 1
	}

	L.SetGlobal("sum", New(L, sum))

	if err := L.DoString(code); err != nil {
		panic(err)
	}

	// Output:
	// 15
}

func ExampleNewType() {
	L := lua.NewState()
	defer L.Close()

	type Song struct {
		Title  string
		Artist string
	}

	L.SetGlobal("Song", NewType(L, Song{}))
	L.DoString(`
		s = Song()
		s.Title = "Montana"
		s.Artist = "Tycho"
		print(s.Artist .. " - " .. s.Title)
	`)

	// Output:
	// Tycho - Montana
}

func TestImmutableStructFieldModify(t *testing.T) {
	// Modifying a field on an immutable struct - should error
	const code = `
	p.Name = "Tom"
	`

	L := lua.NewState()
	defer L.Close()

	p := Person{
		Name: "Tim",
		Age:  66,
	}

	L.SetGlobal("p", New(L, p, ReflectOptions{Immutable: true}))

	err := L.DoString(code)
	if err == nil {
		t.Fatal("Expected error, none thrown")
	}
	if !strings.Contains(err.Error(), "invalid operation on immutable struct") {
		t.Fatal("Expected invalid operation error, got:", err)
	}
}

func TestImmutableStructPtrFunc(t *testing.T) {
	// Calling a pointer function on an immutable struct - should error
	const code = `
	p:IncreaseAge()
	`

	L := lua.NewState()
	defer L.Close()

	p := Person{
		Name: "Tim",
		Age:  66,
	}

	L.SetGlobal("p", New(L, &p, ReflectOptions{Immutable: true}))

	err := L.DoString(code)
	if err == nil {
		t.Fatal("Expected error, none thrown")
	}
	if !strings.Contains(err.Error(), "cannot call pointer methods on immutable objects") {
		t.Fatal("Expected call error, got:", err)
	}
}

func ExampleImmutableStructFieldAccess() {
	// Accessing a field and calling a regular function on an immutable
	// struct - should be fine
	const code = `
	print(p:Hello())
	print(p.Name)
	`

	L := lua.NewState()
	defer L.Close()

	p := Person{
		Name: "Tim",
		Age:  66,
	}

	L.SetGlobal("p", New(L, p, ReflectOptions{Immutable: true}))

	if err := L.DoString(code); err != nil {
		panic(err)
	}

	// Output:
	// Hello, Tim
	// Tim
}

func TestImmutableSliceAssignment(t *testing.T) {
	// Attempting to modify an immutable slice - should error
	const code = `
	s[1] = "hi"
	`

	L := lua.NewState()
	defer L.Close()

	s := []string{"first", "second"}
	L.SetGlobal("s", New(L, s, ReflectOptions{Immutable: true}))

	err := L.DoString(code)
	if err == nil {
		t.Fatal("Expected error, none thrown")
	}
	if !strings.Contains(err.Error(), "invalid operation on immutable slice") {
		t.Fatal("Expected invalid operation error, got:", err)
	}
}

func TestImmutableSliceAppend(t *testing.T) {
	// Attempting to append to an immutable slice - should error
	const code = `
	s = s:append("hi")
	`

	L := lua.NewState()

	s := []string{"first", "second"}
	L.SetGlobal("s", New(L, s, ReflectOptions{Immutable: true}))

	err := L.DoString(code)
	if err == nil {
		t.Fatal("Expected error, none thrown")
	}
	if !strings.Contains(err.Error(), "invalid operation on immutable slice") {
		t.Fatal("Expected invalid operation error, got:", err)
	}
}

func ExampleImmutableSliceAccess() {
	// Attempting to access a member of an immutable slice - should be fine
	const code = `
	print(s[1])
	`

	L := lua.NewState()
	defer L.Close()

	s := []string{"first", "second"}
	L.SetGlobal("s", New(L, s, ReflectOptions{Immutable: true}))

	if err := L.DoString(code); err != nil {
		panic(err)
	}

	// Output:
	// first
}

func TestImmutableMapAssignment(t *testing.T) {
	// Attempting to modify an immutable map - should error
	const code = `
	m["newKey"] = "hi"
	`

	L := lua.NewState()
	defer L.Close()

	m := map[string]string{"first": "foo", "second": "bar"}
	L.SetGlobal("m", New(L, m, ReflectOptions{Immutable: true}))

	err := L.DoString(code)
	if err == nil {
		t.Fatal("Expected error, none thrown")
	}
	if !strings.Contains(err.Error(), "invalid operation on immutable map") {
		t.Fatal("Expected invalid operation error, got:", err)
	}
}

func ExampleImmutableMapAccess() {
	// Attempting to access a member of an immutable map - should be fine
	const code = `
	print(m["first"])
	`

	L := lua.NewState()
	defer L.Close()

	m := map[string]string{"first": "foo", "second": "bar"}
	L.SetGlobal("m", New(L, m, ReflectOptions{Immutable: true}))

	if err := L.DoString(code); err != nil {
		panic(err)
	}

	// Output:
	// foo
}

func TestImmutableNestedStructField(t *testing.T) {
	// Attempt to modify a nested field on an immutable struct - should error
	const code = `
	f.Mother.Name = "Laura"
	`

	L := lua.NewState()
	defer L.Close()

	f := Family{
		Mother: Person{
			Name: "Luara",
		},
		Father: Person{
			Name: "Tim",
		},
	}

	L.SetGlobal("f", New(L, f, ReflectOptions{Immutable: true}))

	err := L.DoString(code)
	if err == nil {
		t.Fatal("Expected error, none thrown")
	}
	if !strings.Contains(err.Error(), "invalid operation on immutable struct") {
		t.Fatal("Expected invalid operation error, got:", err)
	}
}

func TestImmutableNestedStructFieldVar(t *testing.T) {
	// Assign a nested struct field to a variable - should inherit
	// parent's immutable setting and cause error
	const code = `
	mother = f.Mother
	mother.Name = "Laura"
	`

	L := lua.NewState()
	defer L.Close()

	f := Family{
		Mother: Person{
			Name: "Luara",
		},
		Father: Person{
			Name: "Tim",
		},
	}

	L.SetGlobal("f", New(L, f, ReflectOptions{Immutable: true}))

	err := L.DoString(code)
	if err == nil {
		t.Fatal("Expected error, none thrown")
	}
	if !strings.Contains(err.Error(), "invalid operation on immutable struct") {
		t.Fatal("Expected invalid operation error, got:", err)
	}
}

func TestImmutableNestedStructSliceField(t *testing.T) {
	// Attempt to modify a nested field in a nested slice, on an immutable
	// struct - should error
	const code = `
	f.Children[1].Name = "Bill"
	`

	L := lua.NewState()
	defer L.Close()

	f := Family{
		Mother: Person{
			Name: "Luara",
		},
		Father: Person{
			Name: "Tim",
		},
		Children: []Person{
			{Name: "Bill"},
		},
	}

	L.SetGlobal("f", New(L, f, ReflectOptions{Immutable: true}))

	err := L.DoString(code)
	if err == nil {
		t.Fatal("Expected error, none thrown")
	}
	if !strings.Contains(err.Error(), "invalid operation on immutable struct") {
		t.Fatal("Expected invalid operation error, got:", err)
	}
}

func TestImmutableNestedStructPtrSliceField(t *testing.T) {
	// Attempt to modify a nested field in a nested slice, on an immutable
	// struct pointer - should error
	const code = `
	f.Children[1].Name = "Bill"
	`

	L := lua.NewState()
	defer L.Close()

	f := Family{
		Mother: Person{
			Name: "Luara",
		},
		Father: Person{
			Name: "Tim",
		},
		Children: []Person{
			{Name: "Bill"},
		},
	}

	L.SetGlobal("f", New(L, &f, ReflectOptions{Immutable: true}))

	err := L.DoString(code)
	if err == nil {
		t.Fatal("Expected error, none thrown")
	}
	if !strings.Contains(err.Error(), "invalid operation on immutable struct") {
		t.Fatal("Expected invalid operation error, got:", err)
	}
}

func TestImmutablePointerAssignment(t *testing.T) {
	// Attempt to modify the value of an immutable pointer - should error
	const code = `
	_ = str ^ "world"
	`

	L := lua.NewState()
	defer L.Close()

	str := "hello"

	L.SetGlobal("str", New(L, &str, ReflectOptions{Immutable: true}))

	err := L.DoString(code)
	if err == nil {
		t.Fatal("Expected error, none thrown")
	}
	if !strings.Contains(err.Error(), "invalid operation for immutable pointer") {
		t.Fatal("Expected invalid operation error, got:", err)
	}
}

func ExampleImmutablePointerAccess() {
	// Attempt to access the value of an immutable pointer - should be fine
	const code = `
	print(-str)
	`

	L := lua.NewState()
	defer L.Close()

	str := "hello"

	L.SetGlobal("str", New(L, &str, ReflectOptions{Immutable: true}))

	if err := L.DoString(code); err != nil {
		panic(err)
	}

	// Output:
	// hello
}

func TestImmutableChanClose(t *testing.T) {
	// Attempt to close an immutable channel - should error
	const code = `
	ch:close()
	`

	L := lua.NewState()
	defer L.Close()

	ch := make(chan string)

	L.SetGlobal("ch", New(L, ch, ReflectOptions{Immutable: true}))

	err := L.DoString(code)
	if err == nil {
		t.Fatal("Expected error, none thrown")
	}
	if !strings.Contains(err.Error(), "cannot close immutable channel") {
		t.Fatal("Expected invalid operation error, got:", err)
	}
}

type TransparentPtrAccessB struct {
	Str *string
}

type TransparentPtrAccessA struct {
	B *TransparentPtrAccessB
}

func ExampleTransparentPtrAccess() {
	// Access an undefined pointer field - should auto populate with zero
	// value as if a non-pointer object
	const code = `
	print(b.Str)
	`

	L := lua.NewState()
	defer L.Close()

	val := "foo"
	b := TransparentPtrAccessB{}
	b.Str = &val

	L.SetGlobal("b", New(L, &b, ReflectOptions{TransparentPointers: true}))

	if err := L.DoString(code); err != nil {
		panic(err)
	}

	// Output:
	// foo
}

func ExampleTransparentPtrAssignment() {
	// Assign one pointer value to another, with the left side
	// transparent - requires indirection of the right side since
	// the left behaves like a non-pointer field. They should
	// also be separate objects at that point - no shared address.
	// This is distinct from regular pointer assignment, where
	// modifying a value would change it for both references.
	const code = `
	a.B = -b
	print(a.B.Str)
	b.Str = "new value"
	print(b.Str)
	print(a.B.Str)
	`

	L := lua.NewState()
	defer L.Close()

	val := "assigned ptr value"
	a := TransparentPtrAccessA{}
	b := TransparentPtrAccessB{
		Str: &val,
	}
	L.SetGlobal("a", New(L, &a, ReflectOptions{TransparentPointers: true}))
	L.SetGlobal("b", New(L, &b, ReflectOptions{TransparentPointers: true}))

	if err := L.DoString(code); err != nil {
		panic(err)
	}

	// Output:
	// assigned ptr value
	// new value
	// assigned ptr value
}

func ExampleTransparentPtrValueAssignment() {
	// Assign a non-pointer struct value to a pointer field -
	// should be fine
	const code = `
	a.B = b
	print(a.B.Str)
	print(b.Str)
	`

	L := lua.NewState()
	defer L.Close()

	val := "assigned ptr value"
	a := TransparentPtrAccessA{}
	b := TransparentPtrAccessB{
		Str: &val,
	}
	L.SetGlobal("a", New(L, &a, ReflectOptions{TransparentPointers: true}))
	// Non-pointer
	L.SetGlobal("b", New(L, b, ReflectOptions{TransparentPointers: true}))

	if err := L.DoString(code); err != nil {
		panic(err)
	}

	// Output:
	// assigned ptr value
	// assigned ptr value
}

func TestTransparentValueAccess(t *testing.T) {
	// Attempt to access a nil pointer field on a transparent pointer
	// struct that was reflected by value. Since we can't actually set
	// values back to a struct that was reflected by value (as
	// opposed to by reference), an error will result.
	const code = `
	print(b.Str)
	`

	L := lua.NewState()
	defer L.Close()

	b := TransparentPtrAccessB{}

	// Non-pointer
	L.SetGlobal("b", New(L, b, ReflectOptions{TransparentPointers: true}))

	err := L.DoString(code)
	if err == nil {
		t.Fatal("Expected error, none thrown")
	}
	if !strings.Contains(err.Error(), "cannot transparently create pointer field Str") {
		t.Fatal("Expected invalid operation error, got:", err)
	}
}

func ExampleTransparentNestedStructPtrAccess() {
	// Access an undefined nested pointer field - should auto populate
	// with zero values as if a non-pointer object
	const code = `
	print(a.B.Str)
	`

	L := lua.NewState()
	defer L.Close()

	a := TransparentPtrAccessA{}

	L.SetGlobal("a", New(L, &a, ReflectOptions{TransparentPointers: true}))

	if err := L.DoString(code); err != nil {
		panic(err)
	}

	// Output:
	//
}

func ExampleTransparentNestedStructPtrAssignment() {
	// Set an undefined nested pointer field - should get assigned like
	// a regular non-pointer field
	const code = `
	a.B.Str = "hello, world!"
	print(a.B.Str)
	`

	L := lua.NewState()
	defer L.Close()

	a := TransparentPtrAccessA{}

	L.SetGlobal("a", New(L, &a, ReflectOptions{TransparentPointers: true}))

	if err := L.DoString(code); err != nil {
		panic(err)
	}

	// Output:
	// hello, world!
}

func ExampleTransparentPtrEquality() {
	// Check equality on a pointer field - should act like a plain field
	const code = `
	print(b.Str == "foo")
	`

	L := lua.NewState()
	defer L.Close()

	b := TransparentPtrAccessB{}
	val := "foo"
	b.Str = &val

	L.SetGlobal("b", New(L, &b, ReflectOptions{TransparentPointers: true}))

	if err := L.DoString(code); err != nil {
		panic(err)
	}

	// Output:
	// true
}

func TestTransparentPtrPowOp(t *testing.T) {
	// Access a pointer field in the normal pointer way - should error
	const code = `
	_ = b.Str ^ "hello"
	`

	L := lua.NewState()
	defer L.Close()

	b := TransparentPtrAccessB{}
	val := "foo"
	b.Str = &val

	L.SetGlobal("b", New(L, &b, ReflectOptions{TransparentPointers: true}))

	err := L.DoString(code)
	if err == nil {
		t.Fatal("Expected error, none thrown")
	}
	if !strings.Contains(err.Error(), "cannot perform pow operation between string and string") {
		t.Fatal("Expected invalid operation error, got:", err)
	}
}

type TransparentStructSliceFieldA struct {
	List []string
}

func ExampleTransparentStructSliceField() {
	// Access an undefined slice field - should be automatically created
	const code = `
	print(#a.List)
	`

	L := lua.NewState()
	defer L.Close()

	a := TransparentStructSliceFieldA{}

	L.SetGlobal("a", New(L, &a, ReflectOptions{TransparentPointers: true}))

	if err := L.DoString(code); err != nil {
		panic(err)
	}

	// Output:
	// 0
}

func ExampleTransparentStructSliceAppend() {
	// Append to an undefined slice field - should be fine
	const code = `
	a.List = a.List:append("hi")
	print(#a.List)
	`

	L := lua.NewState()
	defer L.Close()

	a := TransparentStructSliceFieldA{}

	L.SetGlobal("a", New(L, &a, ReflectOptions{TransparentPointers: true}))

	if err := L.DoString(code); err != nil {
		panic(err)
	}

	// Output:
	// 1
}

func ExampleTransparentNestedStructVar() {
	// Assign the value of a pointer field to a variable - variable should
	// inherit the transparent reflect options
	const code = `
	b = a.B
	print(b.Str)
	`

	L := lua.NewState()
	defer L.Close()

	val := "hello, world!"
	a := TransparentPtrAccessA{
		&TransparentPtrAccessB{&val},
	}

	L.SetGlobal("a", New(L, &a, ReflectOptions{TransparentPointers: true}))

	if err := L.DoString(code); err != nil {
		panic(err)
	}

	// Output:
	// hello, world!
}

func ExampleTransparentSliceElementVar() {
	// Assign a slice value to a variable - variable should inherit the
	// transparent reflect options
	const code = `
	a = list[1]
	print(a.B.Str)
	`

	val := "hello, world!"
	L := lua.NewState()
	defer L.Close()

	list := []TransparentPtrAccessA{
		{&TransparentPtrAccessB{&val}},
	}

	L.SetGlobal("list", New(L, &list, ReflectOptions{TransparentPointers: true}))

	if err := L.DoString(code); err != nil {
		panic(err)
	}

	// Output:
	// hello, world!
}

func ExampleImmutableTransparentPtrFieldAccess() {
	// Access a transparent pointer field on an immutable
	// struct - should be fine
	const code = `
	print(b.Str)
	`

	L := lua.NewState()
	defer L.Close()

	val := "foo"
	b := TransparentPtrAccessB{}
	b.Str = &val

	L.SetGlobal("b", New(L, &b, ReflectOptions{Immutable: true, TransparentPointers: true}))

	if err := L.DoString(code); err != nil {
		panic(err)
	}

	// Output:
	// foo
}

func TestImmutableTransparentPtrFieldAssignment(t *testing.T) {
	// Attempt to modify a transparent pointer field on an
	// immutable struct - should error
	const code = `
	b.Str = "bar"
	`

	L := lua.NewState()
	defer L.Close()

	val := "foo"
	b := TransparentPtrAccessB{}
	b.Str = &val

	L.SetGlobal("b", New(L, &b, ReflectOptions{Immutable: true, TransparentPointers: true}))

	err := L.DoString(code)
	if err == nil {
		t.Fatal("Expected error, none thrown")
	}
	if !strings.Contains(err.Error(), "invalid operation on immutable struct") {
		t.Fatal("Expected invalid operation error, got:", err)
	}
}

func ExampleMultipleReflectedStructsDifferentOptions() {
	// Set up two structs with different ReflectOptions -
	// should retain their own behaviors
	const code = `
	-- A's pointers are not transparent
	print(-a.Str)
	-- B has transparent pointers, and is mutable
	print(b.Str)
	b.Str = "world"
	print(b.Str)
	-- Assigning to B with transparent pointers shouldn't affect
	-- other pointers that originally pointed to the same string
	print(-a.Str)
	`

	L := lua.NewState()
	defer L.Close()

	val := "hello"
	a := TransparentPtrAccessB{Str: &val}
	b := TransparentPtrAccessB{Str: &val}

	L.SetGlobal("a", New(L, &a, ReflectOptions{Immutable: true}))
	L.SetGlobal("b", New(L, &b, ReflectOptions{TransparentPointers: true}))

	if err := L.DoString(code); err != nil {
		panic(err)
	}

	// Output:
	// hello
	// hello
	// world
	// hello
}

func ExampleTransparentPtrSliceCall() {
	// Iterate over a slice via the call method. Access an undefined pointer
	// field on the returned object. Returned object should inherit the transparent
	// pointer behavior, and the field should be accessible without indirection.
	const code = `
	for i, b in slice() do
		print(i)
		print(b.Str)
	end
	`

	L := lua.NewState()
	defer L.Close()

	val := "foo"
	b := TransparentPtrAccessB{}
	b.Str = &val

	slice := []*TransparentPtrAccessB{&b}

	L.SetGlobal("slice", New(L, slice, ReflectOptions{TransparentPointers: true}))

	if err := L.DoString(code); err != nil {
		panic(err)
	}

	// Output:
	// 1
	// foo
}
