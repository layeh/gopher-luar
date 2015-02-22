// Package luar provides custom type reflection to gopher-lua.
//
// Notice
//
// This package is currently in development, and its behavior may change. This
// message will be removed once the package is considered stable.
//
// Basic types
//
// Go bool, number, and string types are converted to the equivalent basic
// Lua type.
//
// Example:
//  New(L, "Hello World") -> lua.LString("Hello World")
//  New(L, uint(834))     -> lua.LNumber(uint(834))
//
// Channel types
//
// Channel types have a meta table with the __call method defined. Passing no
// arguments when calling is considered a channel receive, and passing one
// argument is considered a channel send.
//
// Example:
//  ch := make(chan string)
//  L.SetGlobal("ch", New(L, ch))
//  ---
//  ch()         -- equivalent to v, ok := ch
//  ch("hello")  -- equivalent to ch <- hello
//
// TODO: close channels from Lua
//
// Function types
//
// Function types have a meta table with the __call method defined. Function
// arguments and returned values will be converted from and to Lua types,
// respectively.
//
// Example:
//  fn := func(name string, age uint) string {
//    return fmt.Sprintf("Hello %s, age %d", name, age)
//  }
//  L.SetGlobal("fn", New(L, fn))
//  ---
//  print(fn("Tim", 5)) -- prints "Hello Tim, age 5"
//
// TODO: variadic functions
//
// Map types
//
// Map types have a meta table with __len, __index, and __newindex defined.
// This allows map values to be fetched and stored.
//
// Example:
//  places := map[string]string{
//    "NA": "North America",
//    "EU": "European Union",
//  }
//  L.SetGlobal("places", New(L, places))
//  ---
//  print(#places)       -- prints "2"
//  print(places.NA)     -- prints "North America"
//  print(places["EU"])  -- prints "European Union"
//
// TODO: implement __call for creating an iterator
//
// Slice types
//
// Slice types have a meta table with __len and __index, which allows for
// accessing slice items.
//
// TODO: slice modification
//
// Struct types
//
// Struct types have a meta table with __index and __newindex. This allows
// accessing struct fields, setting struct fields, and calling struct methods.
//
// Type types
//
// Type constructors can be created using NewType. When called, it returns a
// new variable which is the same type of variable that was passed to NewType.
//
// Example:
//  type Person struct {
//    Name string
//  }
//  L.SetGlobal("Person", NewType(L, Person{}))
//  ---
//  p = Person()
//  p.Name = "John"
//  print("Hello, " .. p.Name)  // prints "Hello, John"
package luar
