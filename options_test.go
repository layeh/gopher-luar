package luar

import (
	"testing"

	"github.com/yuin/gopher-lua"
)

// NOTE: Not testing default options, that should be tested by the remaining
//       tests not depending on specific configuration values.

func Test_FieldNames_UnexportedNameStyle(t *testing.T) {
	L := lua.NewState()

	tim := StructTestPerson{
		Name: "Tim",
		Age:  30,
	}

	Configure(Options{
		FieldNames: UnexportedNameStyle,
	})

	L.SetGlobal("user", New(L, tim))

	testReturn(t, L, `return user.Name`, "nil")
	testReturn(t, L, `return user.name`, "Tim")
	testReturn(t, L, `return user.Age`, "nil")
	testReturn(t, L, `return user.age`, "30")

	// method names should be unaffected
	testReturn(t, L, `return user:Hello()`, "Hello, Tim")
	testReturn(t, L, `return user:hello()`, "Hello, Tim")

	Configure(DefaultOptions)
}

func Test_FieldNames_ExportedNameStyle(t *testing.T) {
	L := lua.NewState()

	tim := StructTestPerson{
		Name: "Tim",
		Age:  30,
	}

	Configure(Options{
		FieldNames: ExportedNameStyle,
	})

	L.SetGlobal("user", New(L, tim))

	testReturn(t, L, `return user.Name`, "Tim")
	testReturn(t, L, `return user.name`, "nil")
	testReturn(t, L, `return user.Age`, "30")
	testReturn(t, L, `return user.age`, "nil")

	// method names should be unaffected
	testReturn(t, L, `return user:Hello()`, "Hello, Tim")
	testReturn(t, L, `return user:hello()`, "Hello, Tim")

	Configure(DefaultOptions)
}

func Test_MethodNames_UnexportedNameStyle(t *testing.T) {
	L := lua.NewState()

	tim := StructTestPerson{
		Name: "Tim",
		Age:  30,
	}

	Configure(Options{
		MethodNames: UnexportedNameStyle,
	})

	L.SetGlobal("user", New(L, tim))

	testReturn(t, L, `return user:hello()`, "Hello, Tim")
	testReturn(t, L, `return user.Hello`, "nil")

	// fiels should be unaffected
	testReturn(t, L, `return user.Name`, "Tim")
	testReturn(t, L, `return user.name`, "Tim")

	Configure(DefaultOptions)
}

func Test_MethodNames_ExportedNameStyle(t *testing.T) {
	L := lua.NewState()

	tim := StructTestPerson{
		Name: "Tim",
		Age:  30,
	}

	Configure(Options{
		MethodNames: ExportedNameStyle,
	})

	L.SetGlobal("user", New(L, tim))

	testReturn(t, L, `return user:Hello()`, "Hello, Tim")
	testReturn(t, L, `return user.hello`, "nil")

	// fiels should be unaffected
	testReturn(t, L, `return user.Name`, "Tim")
	testReturn(t, L, `return user.name`, "Tim")

	Configure(DefaultOptions)
}
