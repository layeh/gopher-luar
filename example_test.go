package luar_test

import (
	"fmt"
	"strconv"

	"github.com/layeh/gopher-luar"
	"github.com/yuin/gopher-lua"
)

type Person struct {
	Name   string
	Age    int
	Friend *Person
}

func (p Person) Hello() string {
	return "Hello, " + p.Name
}

func (p Person) String() string {
	return p.Name + " (" + strconv.Itoa(p.Age) + ")"
}

func Example_1() {
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

	L.SetGlobal("user1", luar.New(L, tim))
	L.SetGlobal("user2", luar.New(L, john))

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

func Example_2() {
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

	L.SetGlobal("things", luar.New(L, things))
	L.SetGlobal("thangs", luar.New(L, thangs))

	if err := L.DoString(code); err != nil {
		panic(err)
	}

	fmt.Println(things[0])
	fmt.Println(thangs["GHI"])
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
	// cookie
	// 789
}

func Example_3() {
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

	L.SetGlobal("user1", luar.New(L, tim))
	L.SetGlobal("Person", luar.NewType(L, Person{}))
	L.SetGlobal("People", luar.NewType(L, map[string]*Person{}))

	if err := L.DoString(code); err != nil {
		panic(err)
	}

	everyone := L.GetGlobal("everyone").(*lua.LUserData).Value.(map[string]*Person)
	fmt.Println(len(everyone))
	// Output:
	// John
	// Tim
	// 2
}

func Example_4() {
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

	L.SetGlobal("person", luar.New(L, tim))
	L.SetGlobal("getHello", luar.New(L, fn))

	if err := L.DoString(code); err != nil {
		panic(err)
	}
	// Output:
	// Hello, Tim
}

func Example_5() {
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

	L.SetGlobal("ch", luar.New(L, ch))

	if err := L.DoString(code); err != nil {
		panic(err)
	}
	// Output:
	// Tim	true
	// John	true
	// nil	false
}

func Example_6() {
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

	L.SetGlobal("countries", luar.New(L, countries))

	if err := L.DoString(code); err != nil {
		panic(err)
	}
	// Output:
	// Canada
	// France
	// Japan
}

func Example_7() {
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

	L.SetGlobal("fn", luar.New(L, fn))

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

func Example_8() {
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

	L.SetGlobal("fn", luar.New(L, fn))

	if err := L.DoString(code); err != nil {
		panic(err)
	}
	// Output:
	// 3
	// 2
	// 1
	// 4
}

func Example_9() {
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

	L.SetGlobal("items", luar.New(L, items))

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

func Example_10() {
	const code = `
	ints = newInts(1)
	print(#ints, ints:capacity())

	ints = newInts(0, 10)
	print(#ints, ints:capacity())
	`

	L := lua.NewState()
	defer L.Close()

	type ints []int

	L.SetGlobal("newInts", luar.NewType(L, ints{}))

	if err := L.DoString(code); err != nil {
		panic(err)
	}
	// Output:
	// 1	1
	// 0	10
}

func Example_11() {
	const code = `
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

	L.SetGlobal("p1", luar.New(L, &p1))
	L.SetGlobal("p1_alias", luar.New(L, &p1))
	L.SetGlobal("p2", luar.New(L, &p2))

	if err := L.DoString(code); err != nil {
		panic(err)
	}
	// Output:
	// true
	// true
	// false
}

func Example_12() {
	const code = `
	print(p1)
	print(p2)
	print(a)
	print(b)
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

	a := struct {
		A string
	}{
		A: "hello",
	}

	b := make(chan string)

	L.SetGlobal("p1", luar.New(L, &p1))
	L.SetGlobal("p2", luar.New(L, &p2))
	L.SetGlobal("a", luar.New(L, a))
	L.SetGlobal("b", luar.New(L, b))

	if err := L.DoString(code); err != nil {
		panic(err)
	}
	// Output:
	// Tim (99)
	// John (2)
	// <struct { A string } Value>
	// <chan string Value>
}

func ExampleNewType() {
	L := lua.NewState()
	defer L.Close()

	type Song struct {
		Title  string
		Artist string
	}

	L.SetGlobal("Song", luar.NewType(L, Song{}))
	L.DoString(`
		s = Song()
		s.Title = "Montana"
		s.Artist = "Tycho"
		print(s.Artist .. " - " .. s.Title)
	`)
	// Output:
	// Tycho - Montana
}
