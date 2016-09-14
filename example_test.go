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

type outputLogger struct {
	Lines []string
}

func (l *outputLogger) Log(lines... interface{}) {
	if len(lines) == 0 {
		l.Lines = append(l.Lines, "")
		return
	}

	linesStr := []string{}
	for _, line := range lines {
		linesStr = append(linesStr, fmt.Sprintf("%v", line))
	}
	l.Lines = append(l.Lines, strings.Join(linesStr, "\t"))
}

func (l *outputLogger) Equals(other []string) bool {
	if len(l.Lines) != len(other) {
		return false
	}
	for i, line := range(l.Lines) {
		if line != other[i] {
			return false
		}
	}
	return true
}

func newOutputLogger() *outputLogger {
	return &outputLogger{[]string{}}
}

func newStateWithLogger() (*lua.LState, *outputLogger) {
	L := lua.NewState()
	logger := newOutputLogger()
	L.SetGlobal("log", New(L, logger.Log))
	return L, logger
}

func TestStructUsage(t *testing.T) {
	const code = `
	log(user1.Name)
	log(user1.Age)
	log(user1:Hello())

	log(user2.Name)
	log(user2.Age)
	hello = user2.Hello
	log(hello(user2))
	`

	L, logger := newStateWithLogger()
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
	
	expected := []string{
		"Tim",
		"30",
		"Hello, Tim",
		"John",
		"40",
		"Hello, John",
	}

	if !logger.Equals(expected) {
		t.Fatalf("Unexpected output. Expected:\n%s\n\nActual:\n%s", expected, logger.Lines)
	}
}

func TestMapAndSlice(t *testing.T) {
	const code = `
	for i = 1, #things do
		log(things[i])
	end
	things[1] = "cookie"

	log()

	log(thangs.ABC)
	log(thangs.DEF)
	log(thangs.GHI)
	thangs.GHI = 789
	thangs.ABC = nil
	`

	L, logger := newStateWithLogger()
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

	logger.Log()
	logger.Log(things[0])
	logger.Log(thangs["GHI"])
	_, ok := thangs["ABC"]
	logger.Log(ok)
	
	expected := []string{
		"cake",
		"wallet",
		"calendar",
		"phone",
		"speaker",
		"",
		"123",
		"456",
		"<nil>",
		"",
		"cookie",
		"789",
		"false",
	}

	if !logger.Equals(expected) {
		t.Fatalf("Unexpected output. Expected:\n%s\n\nActual:\n%s", expected, logger.Lines)
	}
}

func TestStructConstructorAndMap(t *testing.T) {
	const code = `
	user2 = Person()
	user2.Name = "John"
	user2.Friend = user1
	log(user2.Name)
	log(user2.Friend.Name)

	everyone = People()
	everyone["tim"] = user1
	everyone["john"] = user2
	`

	L, logger := newStateWithLogger()
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
	logger.Log(len(everyone))
	
	expected := []string{
		"John",
		"Tim",
		"2",
	}

	if !logger.Equals(expected) {
		t.Fatalf("Unexpected output. Expected:\n%s\n\nActual:\n%s", expected, logger.Lines)
	}
}

func TestGoFunc(t *testing.T) {
	const code = `
	log(getHello(person))
	`

	L, logger := newStateWithLogger()
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
	
	expected := []string{
		"Hello, Tim",
	}

	if !logger.Equals(expected) {
		t.Fatalf("Unexpected output. Expected:\n%s\n\nActual:\n%s", expected, logger.Lines)
	}
}

func TestChan(t *testing.T) {
	const code = `
	log(ch:receive())
	ch:send("John")
	log(ch:receive())
	`

	L, logger := newStateWithLogger()
	defer L.Close()

	ch := make(chan string)
	go func() {
		ch <- "Tim"
		name, ok := <-ch
		logger.Log(name, ok)
		close(ch)
	}()

	L.SetGlobal("ch", New(L, ch))

	if err := L.DoString(code); err != nil {
		t.Fatal(err)
	}
	
	expected := []string{
		"Tim	true",
		"John	true",
		"<nil>	false",
	}

	if !logger.Equals(expected) {
		t.Fatalf("Unexpected output. Expected:\n%s\n\nActual:\n%s", expected, logger.Lines)
	}
}

func TestMap(t *testing.T) {
	const code = `
	local sorted = {}
	for k, v in countries() do
		table.insert(sorted, v)
	end
	table.sort(sorted)
	for i = 1, #sorted do
		log(sorted[i])
	end
	`

	L, logger := newStateWithLogger()
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
	
	expected := []string{
		"Canada",
		"France",
		"Japan",
	}

	if !logger.Equals(expected) {
		t.Fatalf("Unexpected output. Expected:\n%s\n\nActual:\n%s", expected, logger.Lines)
	}
}

func TestFuncVariadic(t *testing.T) {
	const code = `
	fn("a", 1, 2, 3)
	fn("b")
	fn("c", 4)
	`

	L, logger := newStateWithLogger()
	defer L.Close()

	fn := func(str string, extra ...int) {
		logger.Log(str)
		for _, x := range extra {
			logger.Log(x)
		}
	}

	L.SetGlobal("fn", New(L, fn))

	if err := L.DoString(code); err != nil {
		t.Fatal(err)
	}
	
	expected := []string{
		"a",
		"1",
		"2",
		"3",
		"b",
		"c",
		"4",
	}

	if !logger.Equals(expected) {
		t.Fatalf("Unexpected output. Expected:\n%s\n\nActual:\n%s", expected, logger.Lines)
	}
}

func TestLuaFuncVariadic(t *testing.T) {
	const code = `
	for _, x in ipairs(fn(1, 2, 3)) do
		log(x)
	end
	for _, x in ipairs(fn()) do
		log(x)
	end
	for _, x in ipairs(fn(4)) do
		log(x)
	end
	`

	L, logger := newStateWithLogger()
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
	
	expected := []string{
		"3",
		"2",
		"1",
		"4",
	}

	if !logger.Equals(expected) {
		t.Fatalf("Unexpected output. Expected:\n%s\n\nActual:\n%s", expected, logger.Lines)
	}
}

func TestSlice(t *testing.T) {
	const code = `
	log(#items)
	log(items:capacity())
	items = items:append("hello", "world")
	log(#items)
	log(items:capacity())
	log(items[1])
	log(items[2])
	`

	L, logger := newStateWithLogger()
	defer L.Close()

	items := make([]string, 0, 10)

	L.SetGlobal("items", New(L, items))

	if err := L.DoString(code); err != nil {
		t.Fatal(err)
	}
	
	expected := []string{
		"0",
		"10",
		"2",
		"10",
		"hello",
		"world",
	}

	if !logger.Equals(expected) {
		t.Fatalf("Unexpected output. Expected:\n%s\n\nActual:\n%s", expected, logger.Lines)
	}
}

func TestSliceCapacity(t *testing.T) {
	const code = `
	ints = newInts(1)
	log(#ints, ints:capacity())

	ints = newInts(0, 10)
	log(#ints, ints:capacity())
	`

	L, logger := newStateWithLogger()
	defer L.Close()

	type ints []int

	L.SetGlobal("newInts", NewType(L, ints{}))

	if err := L.DoString(code); err != nil {
		t.Fatal(err)
	}
	
	expected := []string{
		"1	1",
		"0	10",
	}

	if !logger.Equals(expected) {
		t.Fatalf("Unexpected output. Expected:\n%s\n\nActual:\n%s", expected, logger.Lines)
	}
}

func TestStructPtrEquality(t *testing.T) {
	const code = `
	log(-p1 == -p1)
	log(-p1 == -p1_alias)
	log(p1 == p1)
	log(p1 == p1_alias)
	log(p1 == p2)
	`

	L, logger := newStateWithLogger()
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
	
	expected := []string{
		"true",
		"true",
		"true",
		"true",
		"false",
	}

	if !logger.Equals(expected) {
		t.Fatalf("Unexpected output. Expected:\n%s\n\nActual:\n%s", expected, logger.Lines)
	}
}

func TestStructStringer(t *testing.T) {
	const code = `
	log(p1)
	log(p2)
	`

	L, logger := newStateWithLogger()
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
	
	expected := []string{
		"Tim (99)",
		"John (2)",
	}

	if !logger.Equals(expected) {
		t.Fatalf("Unexpected output. Expected:\n%s\n\nActual:\n%s", expected, logger.Lines)
	}
}

func TestPtrMethod(t *testing.T) {
	const code = `
	log(p:AddNumbers(1, 2, 3, 4, 5))
	`

	L, logger := newStateWithLogger()
	defer L.Close()

	p := Person{
		Name: "Tim",
	}

	L.SetGlobal("p", New(L, &p))

	if err := L.DoString(code); err != nil {
		t.Fatal(err)
	}
	
	expected := []string{
		"Tim counts: 15",
	}

	if !logger.Equals(expected) {
		t.Fatalf("Unexpected output. Expected:\n%s\n\nActual:\n%s", expected, logger.Lines)
	}
}

func TestStruct(t *testing.T) {
	const code = `
	log(p:hello())
	log(p.age)
	`

	L, logger := newStateWithLogger()
	defer L.Close()

	p := Person{
		Name: "Tim",
		Age:  66,
	}

	L.SetGlobal("p", New(L, &p))

	if err := L.DoString(code); err != nil {
		t.Fatal(err)
	}
	
	expected := []string{
		"Hello, Tim",
		"66",
	}

	if !logger.Equals(expected) {
		t.Fatalf("Unexpected output. Expected:\n%s\n\nActual:\n%s", expected, logger.Lines)
	}
}

type OneString [1]string

func (o OneString) Log() string {
	return o[0]
}

func TestArray(t *testing.T) {
	const code = `
	log(#e.V, e.V[1], e.V[2])
	e.V[1] = "World"
	e.V[2] = "Hello"
	log(#e.V, e.V[1], e.V[2])

	log(#arr, arr[1])
	log(arr:Log())
	`

	type Elem struct {
		V [2]string
	}

	L, logger := newStateWithLogger()
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
	
	expected := []string{
		"2	Hello	World",
		"2	World	Hello",
		"1	Test",
		"Test",
	}

	if !logger.Equals(expected) {
		t.Fatalf("Unexpected output. Expected:\n%s\n\nActual:\n%s", expected, logger.Lines)
	}
}

func TestLuaFunc(t *testing.T) {
	const code = `
	log(fn("tim", 5))
	`

	L, logger := newStateWithLogger()
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
	
	expected := []string{
		"tim	tim	tim	tim	tim",
	}

	if !logger.Equals(expected) {
		t.Fatalf("Unexpected output. Expected:\n%s\n\nActual:\n%s", expected, logger.Lines)
	}
}

func TestPtrIndirection(t *testing.T) {
	const code = `
	log(-ptr)
	`

	L, logger := newStateWithLogger()
	defer L.Close()

	str := "hello"

	L.SetGlobal("ptr", New(L, &str))

	if err := L.DoString(code); err != nil {
		t.Fatal(err)
	}
	
	expected := []string{
		"hello",
	}

	if !logger.Equals(expected) {
		t.Fatalf("Unexpected output. Expected:\n%s\n\nActual:\n%s", expected, logger.Lines)
	}
}

func TestPtrEquality(t *testing.T) {
	const code = `
	log(ptr1 == nil)
	log(ptr2 == nil)
	log(ptr1 == ptr2)
	`

	L, logger := newStateWithLogger()
	defer L.Close()

	var ptr1 *string
	str := "hello"

	L.SetGlobal("ptr1", New(L, ptr1))
	L.SetGlobal("ptr2", New(L, &str))

	if err := L.DoString(code); err != nil {
		t.Fatal(err)
	}
	
	expected := []string{
		"true",
		"false",
		"false",
	}

	if !logger.Equals(expected) {
		t.Fatalf("Unexpected output. Expected:\n%s\n\nActual:\n%s", expected, logger.Lines)
	}
}

func TestPtrAssignment(t *testing.T) {
	const code = `
	log(-str)
	log(str ^ "world")
	log(-str)
	`

	L, logger := newStateWithLogger()
	defer L.Close()

	str := "hello"

	L.SetGlobal("str", New(L, &str))

	if err := L.DoString(code); err != nil {
		t.Fatal(err)
	}
	
	expected := []string{
		"hello",
		"world",
		"world",
	}

	if !logger.Equals(expected) {
		t.Fatalf("Unexpected output. Expected:\n%s\n\nActual:\n%s", expected, logger.Lines)
	}
}

type AnonymousFieldsA struct {
	*AnonymousFieldsB
}

type AnonymousFieldsB struct {
	Value *string
	Person
}

func TestAnonymousFields(t *testing.T) {
	const code = `
	log(a.Value == nil)
	a.Value = str_ptr()
	_ = a.Value ^ "hello"
	log(a.Value == nil)
	log(-a.Value)
	log(a.Name)
	`

	L, logger := newStateWithLogger()
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
		t.Fatal(err)
	}
	
	expected := []string{
		"true",
		"false",
		"hello",
		"Tim",
	}

	if !logger.Equals(expected) {
		t.Fatalf("Unexpected output. Expected:\n%s\n\nActual:\n%s", expected, logger.Lines)
	}
}

func TestEmptyFunc(t *testing.T) {
	const code = `
	log(fn == nil)
	`

	L, logger := newStateWithLogger()
	defer L.Close()

	var fn func()

	L.SetGlobal("fn", New(L, fn))

	if err := L.DoString(code); err != nil {
		t.Fatal(err)
	}
	
	expected := []string{
		"true",
	}

	if !logger.Equals(expected) {
		t.Fatalf("Unexpected output. Expected:\n%s\n\nActual:\n%s", expected, logger.Lines)
	}
}

func TestFuncArray(t *testing.T) {
	const code = `
	fn(arr)
	`

	L, logger := newStateWithLogger()
	defer L.Close()

	arr := [3]int{1, 2, 3}
	fn := func(val [3]int) {
		logger.Log(val[0], val[1], val[2])
	}

	L.SetGlobal("fn", New(L, fn))
	L.SetGlobal("arr", New(L, arr))

	if err := L.DoString(code); err != nil {
		t.Fatal(err)
	}
	
	expected := []string{
		"1	2	3",
	}

	if !logger.Equals(expected) {
		t.Fatalf("Unexpected output. Expected:\n%s\n\nActual:\n%s", expected, logger.Lines)
	}
}

func TestComplex(t *testing.T) {
	const code = `
	b = a
	`

	L, logger := newStateWithLogger()
	defer L.Close()

	a := complex(float64(1), float64(2))

	L.SetGlobal("a", New(L, a))

	if err := L.DoString(code); err != nil {
		t.Fatal(err)
	}
	b := L.GetGlobal("b").(*lua.LUserData).Value.(complex128)
	logger.Log(a == b)
	
	expected := []string{
		"true",
	}

	if !logger.Equals(expected) {
		t.Fatalf("Unexpected output. Expected:\n%s\n\nActual:\n%s", expected, logger.Lines)
	}
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

func TestTypeAlias(t *testing.T) {
	const code = `
	log(a:Test())
	local len1 = b:Len()
	b:Append("!")
	log(len1, b:len())
	log(c.x, c:y())
	`

	L, logger := newStateWithLogger()
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
	
	expected := []string{
		"I'm a \"chan string\" alias",
		"2	3",
		"15	1",
	}

	if !logger.Equals(expected) {
		t.Fatalf("Unexpected output. Expected:\n%s\n\nActual:\n%s", expected, logger.Lines)
	}
}

type StructPtrFuncB struct {
}

func (*StructPtrFuncB) Test() string {
	return "Pointer test"
}

type StructPtrFuncA struct {
	B StructPtrFuncB
}

func TestStructPtrFunc(t *testing.T) {
	const code = `
	log(a.b:Test())
	`

	L, logger := newStateWithLogger()
	defer L.Close()

	a := StructPtrFuncA{}
	L.SetGlobal("a", New(L, &a))

	if err := L.DoString(code); err != nil {
		t.Fatal(err)
	}
	logger.Log(a.B.Test())
	
	expected := []string{
		"Pointer test",
		"Pointer test",
	}

	if !logger.Equals(expected) {
		t.Fatalf("Unexpected output. Expected:\n%s\n\nActual:\n%s", expected, logger.Lines)
	}
}

type HiddenFieldNamesA struct {
	Name   string `luar:"name"`
	Name2  string `luar:"Name"`
	Str    string
	Hidden bool `luar:"-"`
}

func TestHiddenFieldNames(t *testing.T) {
	const code = `
	log(a.name)
	log(a.Name)
	log(a.str)
	log(a.Str)
	log(a.Hidden)
	log(a.hidden)
	`

	L, logger := newStateWithLogger()
	defer L.Close()

	a := &HiddenFieldNamesA{
		Name:   "tim",
		Name2:  "bob",
		Str:    "asd123",
		Hidden: true,
	}

	L.SetGlobal("a", New(L, a))

	if err := L.DoString(code); err != nil {
		t.Fatal(err)
	}
	
	expected := []string{
		"tim",
		"bob",
		"asd123",
		"asd123",
		"<nil>",
		"<nil>",
	}

	if !logger.Equals(expected) {
		t.Fatalf("Unexpected output. Expected:\n%s\n\nActual:\n%s", expected, logger.Lines)
	}
}

func TestStructPtrAssignment(t *testing.T) {
	const code = `
	log(a.Name)
	_ = a ^ -b
	log(a.Name)
	`

	L, logger := newStateWithLogger()
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
	
	expected := []string{
		"tim",
		"bob",
	}

	if !logger.Equals(expected) {
		t.Fatalf("Unexpected output. Expected:\n%s\n\nActual:\n%s", expected, logger.Lines)
	}
}

type PtrNonPtrChanMethodsA chan string

func (*PtrNonPtrChanMethodsA) Test() string {
	return "Test"
}

func (PtrNonPtrChanMethodsA) Test2() string {
	return "Test2"
}

func TestPtrNonPtrChanMethods(t *testing.T) {
	const code = `
	log(b:Test())
	log(b:Test2())
	`

	L, logger := newStateWithLogger()
	defer L.Close()

	a := make(PtrNonPtrChanMethodsA)
	b := &a

	logger.Log(b.Test())
	logger.Log(b.Test2())

	L.SetGlobal("b", New(L, b))

	if err := L.DoString(code); err != nil {
		t.Fatal(err)
	}
	
	expected := []string{
		"Test",
		"Test2",
		"Test",
		"Test2",
	}

	if !logger.Equals(expected) {
		t.Fatalf("Unexpected output. Expected:\n%s\n\nActual:\n%s", expected, logger.Lines)
	}
}

type StructFieldA string

type StructFieldB struct {
	StructFieldA
}

func TestStructField(t *testing.T) {
	const code = `
	a.StructFieldA = "world"
	`

	L, logger := newStateWithLogger()
	defer L.Close()

	a := StructFieldB{}
	a.StructFieldA = "hello"
	logger.Log(a.StructFieldA)

	L.SetGlobal("a", New(L, &a))

	if err := L.DoString(code); err != nil {
		t.Fatal(err)
	}
	logger.Log(a.StructFieldA)
	
	expected := []string{
		"hello",
		"world",
	}

	if !logger.Equals(expected) {
		t.Fatalf("Unexpected output. Expected:\n%s\n\nActual:\n%s", expected, logger.Lines)
	}
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

func TestStructBlacklist(t *testing.T) {
	const code = `
	log(b:public())
	log(b.StructBlacklistA:public())
	pcall(function()
		log(b:private())
	end)
	pcall(function()
		log(b.StructBlacklistA:private())
	end)
	pcall(function()
		log(b:Private())
	end)
	pcall(function()
		log(b.StructBlacklistA:Private())
	end)
	pcall(function()
		local a = -b.StructBlacklistA
		log(a:Private())
	end)
	`

	L, logger := newStateWithLogger()
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
		t.Fatal(err)
	}
	
	expected := []string{
		"You can call me",
		"You can call me",
	}

	if !logger.Equals(expected) {
		t.Fatalf("Unexpected output. Expected:\n%s\n\nActual:\n%s", expected, logger.Lines)
	}
}

type SliceAssignmentA struct {
	S []string
}

func TestSliceAssignment(t *testing.T) {
	const code = `
	x.S = {"a", "b", "", 3, true, "c"}
	`

	L, logger := newStateWithLogger()
	defer L.Close()

	e := &SliceAssignmentA{}
	L.SetGlobal("x", New(L, e))

	if err := L.DoString(code); err != nil {
		t.Fatal(err)
	}
	for _, v := range e.S {
		logger.Log(v)
	}
	
	expected := []string{
		"a",
		"b",
		"",
		"3",
		"true",
		"c",
	}

	if !logger.Equals(expected) {
		t.Fatalf("Unexpected output. Expected:\n%s\n\nActual:\n%s", expected, logger.Lines)
	}
}

type SliceTableAssignmentA struct {
	S map[string]string
}

func TestSliceTableAssignment(t *testing.T) {
	const code = `
	x.S = {
		33,
		a = 123,
		b = nil,
		c = "hello",
		d = false
	}
	`

	L, logger := newStateWithLogger()
	defer L.Close()

	e := &SliceTableAssignmentA{}
	L.SetGlobal("x", New(L, e))

	if err := L.DoString(code); err != nil {
		t.Fatal(err)
	}

	logger.Log(len(e.S))
	logger.Log(e.S["a"])
	logger.Log(e.S["b"])
	logger.Log(e.S["c"])
	logger.Log(e.S["d"])
	
	expected := []string{
		"3",
		"123",
		"",
		"hello",
		"false",
	}

	if !logger.Equals(expected) {
		t.Fatalf("Unexpected output. Expected:\n%s\n\nActual:\n%s", expected, logger.Lines)
	}
}

type FieldNameResolutionA struct {
	Person
	P  Person
	P2 Person `luar:"other"`
}

func TestFieldNameResolution(t *testing.T) {
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

	L, logger := newStateWithLogger()
	defer L.Close()

	e := &FieldNameResolutionA{}
	L.SetGlobal("x", New(L, e))

	if err := L.DoString(code); err != nil {
		t.Fatal(err)
	}

	logger.Log(e.Name)
	logger.Log(e.Age)
	logger.Log(e.P.Name)
	logger.Log(e.P.Age)
	logger.Log(e.P.Friend.Name)
	logger.Log(e.P.Friend.Age)
	logger.Log(e.P2.Name)
	logger.Log(e.P2.Age)
	
	expected := []string{
		"Bill",
		"33",
		"Tim",
		"94",
		"Bob",
		"77",
		"Dale",
		"26",
	}

	if !logger.Equals(expected) {
		t.Fatalf("Unexpected output. Expected:\n%s\n\nActual:\n%s", expected, logger.Lines)
	}
}

type PCallA struct {
	A string `luar:"q"`
	B int    `luar:"other"`
	C int    `luar:"-"`
}

func TestPCall(t *testing.T) {
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

	L, logger := newStateWithLogger()
	defer L.Close()

	e := &PCallA{}
	L.SetGlobal("x", New(L, e))

	if err := L.DoString(code); err != nil {
		t.Fatal(err)
	}

	logger.Log(e.A)
	logger.Log(e.B)
	logger.Log(e.C)
	
	expected := []string{
		"Cat",
		"675",
		"0",
	}

	if !logger.Equals(expected) {
		t.Fatalf("Unexpected output. Expected:\n%s\n\nActual:\n%s", expected, logger.Lines)
	}
}

type LuaFuncDefinitionA struct {
	Fn  func(a string) (string, int)
	Fn2 func(a string, b ...int) string
}

func TestLuaFuncDefinition(t *testing.T) {
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

	L, logger := newStateWithLogger()
	defer L.Close()

	e := &LuaFuncDefinitionA{}
	L.SetGlobal("x", New(L, e))

	if err := L.DoString(code); err != nil {
		t.Fatal(err)
	}

	for ch := 'A'; ch <= 'C'; ch++ {
		str, i := e.Fn(string(ch))
		logger.Log(str, i)
	}

	logger.Log(e.Fn2("hello", 1, 2))
	logger.Log(e.Fn2("hello", 1, 2, 3))

	if L.GetTop() != 0 {
		t.Fatal("expecting GetTop to return 0, got " + strconv.Itoa(L.GetTop()))
	}
	
	expected := []string{
		">A<	1",
		">B<	2",
		">C<	3",
		"",
		"hello",
	}

	if !logger.Equals(expected) {
		t.Fatalf("Unexpected output. Expected:\n%s\n\nActual:\n%s", expected, logger.Lines)
	}
}

type LuaFuncPtrA struct {
	F1 *lua.LFunction
}

func TestLuaFuncPtr(t *testing.T) {
	const code = `
	x.F1 = function(str)
		log("Hello World")
	end
	`

	L, logger := newStateWithLogger()
	defer L.Close()

	e := &LuaFuncPtrA{}
	L.SetGlobal("x", New(L, e))

	if err := L.DoString(code); err != nil {
		t.Fatal(err)
	}

	L.Push(e.F1)
	L.Call(0, 0)

	
	expected := []string{
		"Hello World",
	}

	if !logger.Equals(expected) {
		t.Fatalf("Unexpected output. Expected:\n%s\n\nActual:\n%s", expected, logger.Lines)
	}
}

func TestSliceAndArrayTypes(t *testing.T) {
	const code = `
	for i, x in s() do
		log(i, x)
	end
	for i, x in e() do
		log(i, x)
	end
	for i, x in a() do
		log(i, x)
	end
	for i, x in ap() do
		log(i, x)
	end
	`

	L, logger := newStateWithLogger()
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

	
	expected := []string{
		"1	hello",
		"2	there",
		"3	tim",
		"1	x",
		"2	y",
		"1	x",
		"2	y",
	}

	if !logger.Equals(expected) {
		t.Fatalf("Unexpected output. Expected:\n%s\n\nActual:\n%s", expected, logger.Lines)
	}
}

type StructArrayAndSliceA string

func (s *StructArrayAndSliceA) ToUpper() {
	*s = StructArrayAndSliceA(strings.ToUpper(string(*s)))
}

func TestStructArrayAndSlice(t *testing.T) {
	const code = `
	log(a[1]:AddNumbers(1, 2, 3, 4, 5))
	log(s[1]:AddNumbers(1, 2, 3, 4))
	log(s[1].LastAddSum)
	log(p:AddNumbers(1, 2, 3, 4, 5))
	log(p.LastAddSum)

	log(p.Age)
	p:IncreaseAge()
	log(p.Age)

	log(-str)
	str:ToUpper()
	log(-str)
	`

	L, logger := newStateWithLogger()
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
		t.Fatal(err)
	}

	
	expected := []string{
		"Tim counts: 15",
		"Tim counts: 10",
		"10",
		"Tim counts: 15",
		"15",
		"32",
		"33",
		"Hello World",
		"HELLO WORLD",
	}

	if !logger.Equals(expected) {
		t.Fatalf("Unexpected output. Expected:\n%s\n\nActual:\n%s", expected, logger.Lines)
	}
}

func TestLStateFunc(t *testing.T) {
	const code = `
	log(sum(1, 2, 3, 4, 5))
	`

	L, logger := newStateWithLogger()
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
	
	expected := []string{
		"15",
	}

	if !logger.Equals(expected) {
		t.Fatalf("Unexpected output. Expected:\n%s\n\nActual:\n%s", expected, logger.Lines)
	}
}

func TestNewType(t *testing.T) {
	L, logger := newStateWithLogger()
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
		log(s.Artist .. " - " .. s.Title)
	`)
	
	expected := []string{
		"Tycho - Montana",
	}

	if !logger.Equals(expected) {
		t.Fatalf("Unexpected output. Expected:\n%s\n\nActual:\n%s", expected, logger.Lines)
	}
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
	if !strings.Contains(err.Error(), "attempt to call a non-function object") {
		t.Fatal("Expected call error, got:", err)
	}
}

func TestImmutableStructFieldAccess(t *testing.T) {
	// Accessing a field and calling a regular function on an immutable
	// struct - should be fine
	const code = `
	log(p:Hello())
	log(p.Name)
	`

	L, logger := newStateWithLogger()
	defer L.Close()

	p := Person{
		Name: "Tim",
		Age:  66,
	}

	L.SetGlobal("p", New(L, p, ReflectOptions{Immutable: true}))

	if err := L.DoString(code); err != nil {
		t.Fatal(err)
	}

	expected := []string{
		"Hello, Tim",
		"Tim",
	}

	if !logger.Equals(expected) {
		t.Fatalf("Unexpected output. Expected:\n%s\n\nActual:\n%s", expected, logger.Lines)
	}
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

func TestImmutableSliceAccess(t *testing.T) {
	// Attempting to access a member of an immutable slice - should be fine
	const code = `
	log(s[1])
	`

	L, logger := newStateWithLogger()
	defer L.Close()

	s := []string{"first", "second"}
	L.SetGlobal("s", New(L, s, ReflectOptions{Immutable: true}))

	if err := L.DoString(code); err != nil {
		t.Fatal(err)
	}

	expected := []string{
		"first",
	}

	if !logger.Equals(expected) {
		t.Fatalf("Unexpected output. Expected:\n%s\n\nActual:\n%s", expected, logger.Lines)
	}
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

func TestImmutableMapAccess(t *testing.T) {
	// Attempting to access a member of an immutable map - should be fine
	const code = `
	log(m["first"])
	`

	L, logger := newStateWithLogger()
	defer L.Close()

	m := map[string]string{"first": "foo", "second": "bar"}
	L.SetGlobal("m", New(L, m, ReflectOptions{Immutable: true}))

	if err := L.DoString(code); err != nil {
		t.Fatal(err)
	}

	expected := []string{
		"foo",
	}

	if !logger.Equals(expected) {
		t.Fatalf("Unexpected output. Expected:\n%s\n\nActual:\n%s", expected, logger.Lines)
	}
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

func TestImmutablePointerAccess(t *testing.T) {
	// Attempt to access the value of an immutable pointer - should be fine
	const code = `
	log(-str)
	`

	L, logger := newStateWithLogger()
	defer L.Close()

	str := "hello"

	L.SetGlobal("str", New(L, &str, ReflectOptions{Immutable: true}))

	if err := L.DoString(code); err != nil {
		t.Fatal(err)
	}

	expected := []string{
		"hello",
	}

	if !logger.Equals(expected) {
		t.Fatalf("Unexpected output. Expected:\n%s\n\nActual:\n%s", expected, logger.Lines)
	}
}

type TransparentPtrAccessB struct {
	Str *string
}

type TransparentPtrAccessA struct {
	B *TransparentPtrAccessB
}

func TestTransparentPtrAccess(t *testing.T) {
	// Access an undefined pointer field - should auto populate with zero
	// value as if a non-pointer object
	const code = `
	log(b.Str)
	`

	L, logger := newStateWithLogger()
	defer L.Close()

	val := "foo"
	b := TransparentPtrAccessB{}
	b.Str = &val

	L.SetGlobal("b", New(L, &b, ReflectOptions{TransparentPointers: true}))

	if err := L.DoString(code); err != nil {
		t.Fatal(err)
	}
	
	expected := []string{
		"foo",
	}

	if !logger.Equals(expected) {
		t.Fatalf("Unexpected output. Expected:\n%s\n\nActual:\n%s", expected, logger.Lines)
	}
}

func TestTransparentPtrAssignment(t *testing.T) {
	// Assign one pointer value to another, with the left side
	// transparent - requires indirection of the right side since
	// the left behaves like a non-pointer field. They should
	// also be separate objects at that point - no shared address.
	// This is distinct from regular pointer assignment, where
	// modifying a value would change it for both references.
	const code = `
	a.B = -b
	log(a.B.Str)
	b.Str = "new value"
	log(b.Str)
	log(a.B.Str)
	`

	L, logger := newStateWithLogger()
	defer L.Close()

	val := "assigned ptr value"
	a := TransparentPtrAccessA{}
	b := TransparentPtrAccessB{
		Str: &val,

	}
	L.SetGlobal("a", New(L, &a, ReflectOptions{TransparentPointers: true}))
	L.SetGlobal("b", New(L, &b, ReflectOptions{TransparentPointers: true}))

	if err := L.DoString(code); err != nil {
		t.Fatal(err)
	}
	
	expected := []string{
		"assigned ptr value",
		"new value",
		"assigned ptr value",
	}

	if !logger.Equals(expected) {
		t.Fatalf("Unexpected output. Expected:\n%s\n\nActual:\n%s", expected, logger.Lines)
	}
}

func TestTransparentNestedStructPtrAccess(t *testing.T) {
	// Access an undefined nested pointer field - should auto populate
	// with zero values as if a non-pointer object
	const code = `
	log(a.B.Str)
	`

	L, logger := newStateWithLogger()
	defer L.Close()

	a := TransparentPtrAccessA{}

	L.SetGlobal("a", New(L, &a, ReflectOptions{TransparentPointers: true}))

	if err := L.DoString(code); err != nil {
		t.Fatal(err)
	}
	
	expected := []string{
		"",
	}

	if !logger.Equals(expected) {
		t.Fatalf("Unexpected output. Expected:\n%s\n\nActual:\n%s", expected, logger.Lines)
	}
}

func TestTransparentNestedStructPtrAssignment(t *testing.T) {
	// Set an undefined nested pointer field - should get assigned like
	// a regular non-pointer field
	const code = `
	a.B.Str = "hello, world!"
	log(a.B.Str)
	`

	L, logger := newStateWithLogger()
	defer L.Close()

	a := TransparentPtrAccessA{}

	L.SetGlobal("a", New(L, &a, ReflectOptions{TransparentPointers: true}))

	if err := L.DoString(code); err != nil {
		t.Fatal(err)
	}
	
	expected := []string{
		"hello, world!",
	}

	if !logger.Equals(expected) {
		t.Fatalf("Unexpected output. Expected:\n%s\n\nActual:\n%s", expected, logger.Lines)
	}
}

func TestTransparentPtrEquality(t *testing.T) {
	// Check equality on a pointer field - should act like a plain field
	const code = `
	log(b.Str == "foo")
	`

	L, logger := newStateWithLogger()
	defer L.Close()

	b := TransparentPtrAccessB{}
	val := "foo"
	b.Str = &val

	L.SetGlobal("b", New(L, &b, ReflectOptions{TransparentPointers: true}))

	if err := L.DoString(code); err != nil {
		t.Fatal(err)
	}

	
	expected := []string{
		"true",
	}

	if !logger.Equals(expected) {
		t.Fatalf("Unexpected output. Expected:\n%s\n\nActual:\n%s", expected, logger.Lines)
	}
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

func TestTransparentStructSliceField(t *testing.T) {
	// Access an undefined slice field - should be automatically created
	const code = `
	log(#a.List)
	`

	L, logger := newStateWithLogger()
	defer L.Close()

	a := TransparentStructSliceFieldA{}

	L.SetGlobal("a", New(L, &a, ReflectOptions{TransparentPointers: true}))

	if err := L.DoString(code); err != nil {
		t.Fatal(err)
	}

	
	expected := []string{
		"0",
	}

	if !logger.Equals(expected) {
		t.Fatalf("Unexpected output. Expected:\n%s\n\nActual:\n%s", expected, logger.Lines)
	}
}

func TestTransparentStructSliceAppend(t *testing.T) {
	// Append to an undefined slice field - should be fine
	const code = `
	a.List = a.List:append("hi")
	log(#a.List)
	`

	L, logger := newStateWithLogger()
	defer L.Close()

	a := TransparentStructSliceFieldA{}

	L.SetGlobal("a", New(L, &a, ReflectOptions{TransparentPointers: true}))

	if err := L.DoString(code); err != nil {
		t.Fatal(err)
	}

	expected := []string{
		"1",
	}

	if !logger.Equals(expected) {
		t.Fatalf("Unexpected output. Expected:\n%s\n\nActual:\n%s", expected, logger.Lines)
	}
}

func TestTransparentNestedStructVar(t *testing.T) {
	// Assign the value of a pointer field to a variable - variable should
	// inherit the transparent reflect options
	const code = `
	b = a.B
	log(b.Str)
	`

	L, logger := newStateWithLogger()
	defer L.Close()

	val := "hello, world!"
	a := TransparentPtrAccessA{
		&TransparentPtrAccessB{&val},
	}

	L.SetGlobal("a", New(L, &a, ReflectOptions{TransparentPointers: true}))

	if err := L.DoString(code); err != nil {
		t.Fatal(err)
	}

	
	expected := []string{
		"hello, world!",
	}

	if !logger.Equals(expected) {
		t.Fatalf("Unexpected output. Expected:\n%s\n\nActual:\n%s", expected, logger.Lines)
	}
}

func TestTransparentSliceElementVar(t *testing.T) {
	// Assign a slice value to a variable - variable should inherit the
	// transparent reflect options
	const code = `
	a = list[1]
	log(a.B.Str)
	`

	val := "hello, world!"
	L, logger := newStateWithLogger()
	defer L.Close()

	list := []TransparentPtrAccessA{
		{&TransparentPtrAccessB{&val}},
	}

	L.SetGlobal("list", New(L, &list, ReflectOptions{TransparentPointers: true}))

	if err := L.DoString(code); err != nil {
		t.Fatal(err)
	}

	
	expected := []string{
		"hello, world!",
	}

	if !logger.Equals(expected) {
		t.Fatalf("Unexpected output. Expected:\n%s\n\nActual:\n%s", expected, logger.Lines)
	}
}

func TestImmutableTransparentPtrFieldAccess(t *testing.T) {
	// Access a transparent pointer field on an immutable
	// struct - should be fine
	const code = `
	log(b.Str)
	`

	L, logger := newStateWithLogger()
	defer L.Close()

	val := "foo"
	b := TransparentPtrAccessB{}
	b.Str = &val

	L.SetGlobal("b", New(L, &b, ReflectOptions{Immutable: true, TransparentPointers: true}))

	if err := L.DoString(code); err != nil {
		t.Fatal(err)
	}

	expected := []string{
		"foo",
	}

	if !logger.Equals(expected) {
		t.Fatalf("Unexpected output. Expected:\n%s\n\nActual:\n%s", expected, logger.Lines)
	}
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
