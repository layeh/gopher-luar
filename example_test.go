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

func TestExample__1(t *testing.T) {
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
		t.Fatal(err)
	}
	// Output:
	// Tim
	// 30
	// Hello, Tim
	// John
	// 40
	// Hello, John
}

func TestExample__2(t *testing.T) {
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
		t.Fatal(err)
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

func TestExample__3(t *testing.T) {
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
		t.Fatal(err)
	}

	everyone := L.GetGlobal("everyone").(*lua.LUserData).Value.(map[string]*Person)
	fmt.Println(len(everyone))
	// Output:
	// John
	// Tim
	// 2
}

func TestExample__4(t *testing.T) {
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
		t.Fatal(err)
	}
	// Output:
	// Hello, Tim
}

func TestExample__5(t *testing.T) {
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
		t.Fatal(err)
	}
	// Output:
	// Tim	true
	// John	true
	// nil	false
}

func TestExample__6(t *testing.T) {
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
		t.Fatal(err)
	}
	// Output:
	// Canada
	// France
	// Japan
}

func TestExample__7(t *testing.T) {
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
		t.Fatal(err)
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

func TestExample__8(t *testing.T) {
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
		t.Fatal(err)
	}
	// Output:
	// 3
	// 2
	// 1
	// 4
}

func TestExample__9(t *testing.T) {
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
		t.Fatal(err)
	}
	// Output:
	// 0
	// 10
	// 2
	// 10
	// hello
	// world
}

func TestExample__10(t *testing.T) {
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
		t.Fatal(err)
	}
	// Output:
	// 1	1
	// 0	10
}

func TestExample__11(t *testing.T) {
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
		t.Fatal(err)
	}
	// Output:
	// true
	// true
	// true
	// true
	// false
}

func TestExample__12(t *testing.T) {
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
		t.Fatal(err)
	}
	// Output:
	// Tim (99)
	// John (2)
}

func TestExample__13(t *testing.T) {
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
		t.Fatal(err)
	}
	// Output:
	// Tim counts: 15
}

func TestExample__14(t *testing.T) {
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
		t.Fatal(err)
	}
	// Output:
	// Hello, Tim
	// 66
}

type OneString [1]string

func (o OneString) Print() {
	fmt.Println(o[0])
}

func TestExample__15(t *testing.T) {
	const code = `
	print(#e.V, e.V[1], e.V[2])
	e.V[1] = "World"
	e.V[2] = "Hello"
	print(#e.V, e.V[1], e.V[2])

	print(#arr, arr[1])
	arr:Print()
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
		t.Fatal(err)
	}
	// Output:
	// 2	Hello	World
	// 2	World	Hello
	// 1	Test
	// Test
}

func TestExample__16(t *testing.T) {
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
		t.Fatal(err)
	}
	// Output:
	// tim	tim	tim	tim	tim
}

func TestExample__17(t *testing.T) {
	const code = `
	print(-ptr)
	`

	L := lua.NewState()
	defer L.Close()

	str := "hello"

	L.SetGlobal("ptr", New(L, &str))

	if err := L.DoString(code); err != nil {
		t.Fatal(err)
	}
	// Output:
	// hello
}

func TestExample__18(t *testing.T) {
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
		t.Fatal(err)
	}
	// Output:
	// true
	// false
	// false
}

func TestExample__19(t *testing.T) {
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
		t.Fatal(err)
	}
	// Output:
	// hello
	// world
	// world
}

type Example__20_A struct {
	*Example__20_B
}

type Example__20_B struct {
	Value *string
	Person
}

func TestExample__20(t *testing.T) {
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

	a := Example__20_A{
		Example__20_B: &Example__20_B{
			Person: Person{
				Name: "Tim",
			},
		},
	}

	L.SetGlobal("a", New(L, a))
	L.SetGlobal("str_ptr", NewType(L, ""))

	if err := L.DoString(code); err != nil {
		t.Fatal(err)
	}
	// Output:
	// true
	// false
	// hello
	// Tim
}

func TestExample__21(t *testing.T) {
	const code = `
	print(fn == nil)
	`

	L := lua.NewState()
	defer L.Close()

	var fn func()

	L.SetGlobal("fn", New(L, fn))

	if err := L.DoString(code); err != nil {
		t.Fatal(err)
	}
	// Output:
	// true
}

func TestExample__22(t *testing.T) {
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
		t.Fatal(err)
	}
	// Output:
	// 1 2 3
}

func TestExample__23(t *testing.T) {
	const code = `
	b = a
	`

	L := lua.NewState()
	defer L.Close()

	a := complex(float64(1), float64(2))

	L.SetGlobal("a", New(L, a))

	if err := L.DoString(code); err != nil {
		t.Fatal(err)
	}
	b := L.GetGlobal("b").(*lua.LUserData).Value.(complex128)
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

func TestExample__24(t *testing.T) {
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
		t.Fatal(err)
	}
	// Output:
	// I'm a "chan string" alias
	// 2	3
	// 15	1
}

type E25B struct {
}

func (*E25B) Test() {
	fmt.Println("Pointer test")
}

type E25A struct {
	B E25B
}

func TestExample__25(t *testing.T) {
	const code = `
	a.b:Test()
	`

	L := lua.NewState()
	defer L.Close()

	a := E25A{}
	L.SetGlobal("a", New(L, &a))

	if err := L.DoString(code); err != nil {
		t.Fatal(err)
	}
	a.B.Test()
	// Output:
	// Pointer test
	// Pointer test
}

type E26A struct {
	Name   string `luar:"name"`
	Name2  string `luar:"Name"`
	Str    string
	Hidden bool `luar:"-"`
}

func TestExample__26(t *testing.T) {
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

	a := &E26A{
		Name:   "tim",
		Name2:  "bob",
		Str:    "asd123",
		Hidden: true,
	}

	L.SetGlobal("a", New(L, a))

	if err := L.DoString(code); err != nil {
		t.Fatal(err)
	}
	// Output:
	// tim
	// bob
	// asd123
	// asd123
	// nil
	// nil
}

func TestExample__27(t *testing.T) {
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
		t.Fatal(err)
	}
	// Output:
	// tim
	// bob
}

type E28_Chan chan string

func (*E28_Chan) Test() {
	fmt.Println("E28_Chan.Test")
}

func (E28_Chan) Test2() {
	fmt.Println("E28_Chan.Test2")
}

func TestExample__28(t *testing.T) {
	const code = `
	b:Test()
	b:Test2()
	`

	L := lua.NewState()
	defer L.Close()

	a := make(E28_Chan)
	b := &a

	b.Test()
	b.Test2()

	L.SetGlobal("b", New(L, b))

	if err := L.DoString(code); err != nil {
		t.Fatal(err)
	}
	// Output:
	// E28_Chan.Test
	// E28_Chan.Test2
	// E28_Chan.Test
	// E28_Chan.Test2
}

type E29_String string

type E29_A struct {
	E29_String
}

func TestExample__29(t *testing.T) {
	const code = `
	a.E29_String = "world"
	`

	L := lua.NewState()
	defer L.Close()

	a := E29_A{}
	a.E29_String = "hello"
	fmt.Println(a.E29_String)

	L.SetGlobal("a", New(L, &a))

	if err := L.DoString(code); err != nil {
		t.Fatal(err)
	}
	fmt.Println(a.E29_String)
	// Output:
	// hello
	// world
}

type E30_A struct {
}

func (*E30_A) Public() {
	fmt.Println("You can call me")
}

func (E30_A) Private() {
	fmt.Println("Should not be able to call me")
}

type E30_B struct {
	*E30_A
}

func TestExample__30(t *testing.T) {
	const code = `
	b:public()
	b.E30_A:public()
	pcall(function()
		b:private()
	end)
	pcall(function()
		b.E30_A:private()
	end)
	pcall(function()
		b:Private()
	end)
	pcall(function()
		b.E30_A:Private()
	end)
	pcall(function()
		local a = -b.E30_A
		a:Private()
	end)
	`

	L := lua.NewState()
	defer L.Close()

	b := &E30_B{
		E30_A: &E30_A{},
	}

	mt := MT(L, E30_B{})
	mt.Blacklist("private", "Private")

	mt = MT(L, E30_A{})
	mt.Whitelist("public", "Public")

	L.SetGlobal("b", New(L, b))

	if err := L.DoString(code); err != nil {
		t.Fatal(err)
	}
	// Output:
	// You can call me
	// You can call me
}

type E_31 struct {
	S []string
}

func TestExample__31(t *testing.T) {
	const code = `
	x.S = {"a", "b", "", 3, true, "c"}
	`

	L := lua.NewState()
	defer L.Close()

	e := &E_31{}
	L.SetGlobal("x", New(L, e))

	if err := L.DoString(code); err != nil {
		t.Fatal(err)
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

type E_32 struct {
	S map[string]string
}

func TestExample__32(t *testing.T) {
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

	e := &E_32{}
	L.SetGlobal("x", New(L, e))

	if err := L.DoString(code); err != nil {
		t.Fatal(err)
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

type E_33 struct {
	Person
	P  Person
	P2 Person `luar:"other"`
}

func TestExample__33(t *testing.T) {
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

	e := &E_33{}
	L.SetGlobal("x", New(L, e))

	if err := L.DoString(code); err != nil {
		t.Fatal(err)
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

type E_34 struct {
	A string `luar:"q"`
	B int    `luar:"other"`
	C int    `luar:"-"`
}

func TestExample__34(t *testing.T) {
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

	e := &E_34{}
	L.SetGlobal("x", New(L, e))

	if err := L.DoString(code); err != nil {
		t.Fatal(err)
	}

	fmt.Println(e.A)
	fmt.Println(e.B)
	fmt.Println(e.C)
	// Output:
	// Cat
	// 675
	// 0
}

type E_35 struct {
	Fn  func(a string) (string, int)
	Fn2 func(a string, b ...int) string
}

func TestExample__35(t *testing.T) {
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

	e := &E_35{}
	L.SetGlobal("x", New(L, e))

	if err := L.DoString(code); err != nil {
		t.Fatal(err)
	}

	for ch := 'A'; ch <= 'C'; ch++ {
		str, i := e.Fn(string(ch))
		fmt.Printf("%s %d\n", str, i)
	}

	fmt.Println(e.Fn2("hello", 1, 2))
	fmt.Println(e.Fn2("hello", 1, 2, 3))

	if L.GetTop() != 0 {
		t.Fatal("expecting GetTop to return 0, got " + strconv.Itoa(L.GetTop()))
	}
	// Output:
	// >A< 1
	// >B< 2
	// >C< 3
	//
	// hello
}

type E_36 struct {
	F1 *lua.LFunction
}

func TestExample__36(t *testing.T) {
	const code = `
	x.F1 = function(str)
		print("Hello World")
	end
	`

	L := lua.NewState()
	defer L.Close()

	e := &E_36{}
	L.SetGlobal("x", New(L, e))

	if err := L.DoString(code); err != nil {
		t.Fatal(err)
	}

	L.Push(e.F1)
	L.Call(0, 0)

	// Output:
	// Hello World
}

func TestExample__37(t *testing.T) {
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
		t.Fatal(err)
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

type E38String string

func (s *E38String) ToUpper() {
	*s = E38String(strings.ToUpper(string(*s)))
}

func TestExample__38(t *testing.T) {
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

	str := E38String("Hello World")

	L.SetGlobal("a", New(L, &a))
	L.SetGlobal("s", New(L, s))
	L.SetGlobal("p", New(L, s[0]))
	L.SetGlobal("str", New(L, &str))

	if err := L.DoString(code); err != nil {
		t.Fatal(err)
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

func TestExampleLState(t *testing.T) {
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
		t.Fatal(err)
	}
	// Output:
	// 15
}

func TestExampleNewType(t *testing.T) {
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

func TestExample__Immutable1(t *testing.T) {
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

	L.SetGlobal("p", NewWithOptions(L, p, ReflectOptions{Immutable: true}))

	err := L.DoString(code)
	if err == nil {
		t.Fatal("Expected error, none thrown")
	}
	if !strings.Contains(err.Error(), "invalid operation on immutable struct") {
		t.Fatal("Expected invalid operation error, got:", err)
	}
}

func TestExample__Immutable2(t *testing.T) {
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

	L.SetGlobal("p", NewWithOptions(L, &p, ReflectOptions{Immutable: true}))

	err := L.DoString(code)
	if err == nil {
		t.Fatal("Expected error, none thrown")
	}
	if !strings.Contains(err.Error(), "attempt to call a non-function object") {
		t.Fatal("Expected call error, got:", err)
	}
}

func TestExample__Immutable3(t *testing.T) {
	// Accessing a field and calling a regular function on an immutable
	// struct - should be fine
	const code = `
	p:Hello()
	print(p.Name)
	`

	L := lua.NewState()
	defer L.Close()

	p := Person{
		Name: "Tim",
		Age:  66,
	}

	L.SetGlobal("p", NewWithOptions(L, p, ReflectOptions{Immutable: true}))

	if err := L.DoString(code); err != nil {
		t.Fatal(err)
	}
}

func TestExample__Immutable4(t *testing.T) {
	// Attempting to modify an immutable slice - should error
	const code = `
	s[1] = "hi"
	`

	L := lua.NewState()
	defer L.Close()

	s := []string{"first", "second"}
	L.SetGlobal("s", NewWithOptions(L, s, ReflectOptions{Immutable: true}))

	err := L.DoString(code)
	if err == nil {
		t.Fatal("Expected error, none thrown")
	}
	if !strings.Contains(err.Error(), "invalid operation on immutable slice") {
		t.Fatal("Expected invalid operation error, got:", err)
	}
}

func TestExample__Immutable5(t *testing.T) {
	// Attempting to append to an immutable slice - should error
	const code = `
	s = s:append("hi")
	`

	L := lua.NewState()
	defer L.Close()

	s := []string{"first", "second"}
	L.SetGlobal("s", NewWithOptions(L, s, ReflectOptions{Immutable: true}))

	err := L.DoString(code)
	if err == nil {
		t.Fatal("Expected error, none thrown")
	}
	if !strings.Contains(err.Error(), "invalid operation on immutable slice") {
		t.Fatal("Expected invalid operation error, got:", err)
	}
}

func TestExample__Immutable6(t *testing.T) {
	// Attempting to access a member of an immutable slice - should be fine
	const code = `
	print(s[1])
	`

	L := lua.NewState()
	defer L.Close()

	s := []string{"first", "second"}
	L.SetGlobal("s", NewWithOptions(L, s, ReflectOptions{Immutable: true}))

	if err := L.DoString(code); err != nil {
		t.Fatal(err)
	}
}

func TestExample__Immutable7(t *testing.T) {
	// Attempting to modify an immutable map - should error
	const code = `
	m["newKey"] = "hi"
	`

	L := lua.NewState()
	defer L.Close()

	m := map[string]string{"first": "foo", "second": "bar"}
	L.SetGlobal("m", NewWithOptions(L, m, ReflectOptions{Immutable: true}))

	err := L.DoString(code)
	if err == nil {
		t.Fatal("Expected error, none thrown")
	}
	if !strings.Contains(err.Error(), "invalid operation on immutable map") {
		t.Fatal("Expected invalid operation error, got:", err)
	}
}

func TestExample__Immutable8(t *testing.T) {
	// Attempting to access a member of an immutable map - should be fine
	const code = `
	print(m["first"])
	`

	L := lua.NewState()
	defer L.Close()

	m := map[string]string{"first": "foo", "second": "bar"}
	L.SetGlobal("m", NewWithOptions(L, m, ReflectOptions{Immutable: true}))

	if err := L.DoString(code); err != nil {
		t.Fatal(err)
	}
}

type Family struct {
	Mother Person
	Father Person
	Children []Person
}

func TestExample__Immutable9(t *testing.T) {
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

	L.SetGlobal("f", NewWithOptions(L, f, ReflectOptions{Immutable: true}))

	err := L.DoString(code)
	if err == nil {
		t.Fatal("Expected error, none thrown")
	}
	if !strings.Contains(err.Error(), "invalid operation on immutable struct") {
		t.Fatal("Expected invalid operation error, got:", err)
	}
}

func TestExample__Immutable10(t *testing.T) {
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
			{ Name: "Bill" },
		},
	}

	L.SetGlobal("f", NewWithOptions(L, f, ReflectOptions{Immutable: true}))

	err := L.DoString(code)
	if err == nil {
		t.Fatal("Expected error, none thrown")
	}
	if !strings.Contains(err.Error(), "invalid operation on immutable struct") {
		t.Fatal("Expected invalid operation error, got:", err)
	}
}

func TestExample__Immutable11(t *testing.T) {
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
			{ Name: "Bill" },
		},
	}

	L.SetGlobal("f", NewWithOptions(L, &f, ReflectOptions{Immutable: true}))

	err := L.DoString(code)
	if err == nil {
		t.Fatal("Expected error, none thrown")
	}
	if !strings.Contains(err.Error(), "invalid operation on immutable struct") {
		t.Fatal("Expected invalid operation error, got:", err)
	}
}

func TestExample__Immutable12(t *testing.T) {
	// Attempt to modify the value of an immutable pointer - should error
	const code = `
	_ = str ^ "world"
	`

	L := lua.NewState()
	defer L.Close()

	str := "hello"

	L.SetGlobal("str", NewWithOptions(L, &str, ReflectOptions{Immutable: true}))

	err := L.DoString(code)
	if err == nil {
		t.Fatal("Expected error, none thrown")
	}
	if !strings.Contains(err.Error(), "invalid operation for immutable pointer") {
		t.Fatal("Expected invalid operation error, got:", err)
	}
}

func TestExample__Immutable13(t *testing.T) {
	// Attempt to access the value of an immutable pointer - should be fine
	const code = `
	print(-str)
	`

	L := lua.NewState()
	defer L.Close()

	str := "hello"

	L.SetGlobal("str", NewWithOptions(L, &str, ReflectOptions{Immutable: true}))

	if err := L.DoString(code); err != nil {
		t.Fatal(err)
	}
}
