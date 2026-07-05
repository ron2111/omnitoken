package omnitoken_test

import (
	"fmt"

	"github.com/ron2111/omnitoken"
)

func ExampleForModel() {
	engine, err := omnitoken.ForModel("gpt-4o")
	if err != nil {
		panic(err)
	}

	fmt.Println(engine.CountTokens("hello world"))
	fmt.Println(engine.Decode(engine.EncodeOrdinary("hello world")))

	// Output:
	// 2
	// hello world
}
