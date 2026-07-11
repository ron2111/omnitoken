package cacheflow_test

import (
	"fmt"

	"github.com/ron2111/omnitoken"
	"github.com/ron2111/omnitoken/cacheflow"
)

func ExampleAligner_AlignPromptToProfile() {
	engine, err := omnitoken.ForModel("gpt-4o")
	if err != nil {
		panic(err)
	}

	report := cacheflow.NewAligner(engine).AlignPromptToProfile("hello world", cacheflow.ProfileOpenAI)

	fmt.Println(report.CurrentTokens)
	fmt.Println(report.PaddingNeeded)
	fmt.Println(report.IsEligible)

	// Output:
	// 2
	// 1022
	// false
}
