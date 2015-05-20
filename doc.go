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
// Channel types have the following methods defined:
//  receive():    Receives data from the channel. Returns nil plus false if the
//                channel is closed.
//  send(data):   Sends data to the channel.
//  close():      Closes the channel.
//
// Taking the length (#) of a channel returns how many unread items are in its
// buffer.
//
// Example:
//  ch := make(chan string)
//  L.SetGlobal("ch", New(L, ch))
//  ---
//  ch:receive()      -- equivalent to v, ok := ch
//  ch:send("hello")  -- equivalent to ch <- "hello"
//  ch:close()        -- equivalent to close(ch)
//
// Function types
//
// Function types can be called from Lua. Its arguments and returned values
// will be automatically converted from and to Lua types, respectively (see
// exception below). However, a function that uses luar.LState can bypass the
// automatic argument and return value conversion (see luar.LState
// documentation for example).
//
// Example:
//  fn := func(name string, age uint) string {
//    return fmt.Sprintf("Hello %s, age %d", name, age)
//  }
//  L.SetGlobal("fn", New(L, fn))
//  ---
//  print(fn("Tim", 5)) -- prints "Hello Tim, age 5"
//
// A special conversion case happens when function returns a lua.LValue slice.
// In that case, luar will automatically unpack the slice.
//
// Example:
//  fn := func() []lua.LValue {
//    return []lua.LValue{lua.LString("Hello"), lua.LNumber(2.5)}
//  }
//  L.SetGlobal("fn", New(L, fn))
//  ---
//  x, y = fn()
//  print(x) -- prints "Hello"
//  print(y) -- prints "2.5"
//
// Map types
//
// Map types can be accessed and modified like a normal Lua table a meta table.
// Its length can also be queried using the # operator.
//
// Rather than using pairs to create an map iterator, calling the value (e.g.
// map_variable()) will return an iterator for the map.
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
//  for k, v = places() do
//    print(k .. ": " .. v)
//  end
//
// Slice types
//
// Like map types, slices be accessed, be modified, and have their length
// queried. Additionally, the following methods are defined for slices:
//  append(items...):   Appends the items to the slice. Returns a slice with
//                      the items appended.
//  capacity():         Returns the slice capacity.
//
// For consistency with other Lua code, slices use one-based indexing.
//
// Example:
//  letters := []string{"a", "e", "i"}
//  L.SetGlobal("letters", New(L, letters))
//  ---
//  letters = letters:append("o", "u")
//
// Struct types
//
// Struct types can have their fields accessed and modified and their methods
// called. First letters of field/method names are automatically converted to
// uppercase.
//
// Example:
//  type Person {
//    Name string
//  }
//  func (p Person) SayHello() {
//    fmt.Printf("Hello, %s\n", p.Name)
//  }
//
//  tim := Person{"Tim"}
//  L.SetGlobal("tim", New(L, tim))
//  ---
//  tim:SayHello() -- same as tim:sayHello()
//
// Pointer types
//
// Pointers to structs operate the same way structs do. Pointers can also be
// dereferenced using the unary minus (-) operator.
//
// Example:
//  str := "hello"
//  L.SetGlobal("strptr", New(L, &str))
//  ---
//  print(-strptr) -- prints "hello"
//
// The length operator returns true if the pointer is non-nil.
//
// The pointed to value can changed using the pow (^) operator.
//
// Example:
//  str := "hello"
//  L.SetGlobal("strptr", New(L, &str))
//  ---
//  print(str^"world") -- prints "world", and str's value is now "world"
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
