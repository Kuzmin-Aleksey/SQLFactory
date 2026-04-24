package llm

import "SQLFactory/internal/domain/entity"

type Context []entity.LLMContext

func (c *Context) Append(LLMContext entity.LLMContext) {
	if c == nil {
		panic("append to nil LLMContext")
	}
	if len(*c) == 0 {
		*c = []entity.LLMContext{LLMContext}
		return
	}

	previousId := (*c)[len(*c)-1].Id
	if previousId != 0 {
		LLMContext.PreviousId = &previousId
	}
	*c = append(*c, LLMContext)
}
