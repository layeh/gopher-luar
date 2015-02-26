// Package gopher-luar provides custom type reflection to gopher-lua.
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
// will be automatically converted from and to Lua types, respectively.
//
// Example:
//  fn := func(name string, age uint) string {
//    return fmt.Sprintf("Hello %s, age %d", name, age)
//  }
//  L.SetGlobal("fn", New(L, fn))
//  ---
//  print(fn("Tim", 5)) -- prints "Hello Tim, age 5"
//
// Map types
//
// Map types can be accessed and modified like a normal Lua table a meta table.
// Its length can also be queried using the # operator.
//
// Rather than using pairs to create an iterator over Go maps, calling the
// value (e.g. map_variable()) will return an iterator for the map.
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
// Slice types have a meta table with __len and __index, which allows for
// accessing slice items.
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
