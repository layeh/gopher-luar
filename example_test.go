package luar_test

import (
	"fmt"

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

func Example_1() {
	const code = `
	print(user1.Name)
	print(user1.Age)
	print(user1.Hello())

	print(user2.Name)
	print(user2.Age)
	print(user2.Hello())
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
	print(stream())
	stream("John")
	print(stream())
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

	L.SetGlobal("stream", luar.New(L, ch))

	if err := L.DoString(code); err != nil {
		panic(err)
	}
	// Output:
	// Tim	true
	// John	true
	// nil	false
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
