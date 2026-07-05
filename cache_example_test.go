package omnitoken_test

import (
	"fmt"

	"github.com/ron2111/omnitoken"
)

func ExampleCacheAligner_AlignPromptToProfile() {
	engine, err := omnitoken.ForModel("gpt-4o")
	if err != nil {
		panic(err)
	}

	report := omnitoken.NewCacheAligner(engine).AlignPromptToProfile("hello world", omnitoken.CacheProfileOpenAI)

	fmt.Println(report.CurrentTokens)
	fmt.Println(report.PaddingNeeded)
	fmt.Println(report.IsEligible)

	// Output:
	// 2
	// 1022
	// false
}
