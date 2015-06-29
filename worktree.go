package worktree

import (
	"log"
	"time"
)

type CommandTree struct {
	Reducer             func(inp []interface{}) interface{}
	LeafFunctions       []func(inp interface{}) interface{}
	LeafFunctionsInput  []interface{}
	nChildren           int
	LeafFunctionsOutput []interface{}
}

type LeafFunc func(inp interface{}) interface{}

func (t *CommandTree) AddMapper(f func(inp interface{}) interface{}, input interface{}) int {

	t.LeafFunctions = append(t.LeafFunctions, f)
	t.LeafFunctionsInput = append(t.LeafFunctionsInput, input)
	temp := t.nChildren
	t.nChildren += 1
	return temp
}

func (t *CommandTree) AddReducer(f func(inp []interface{}) interface{}) {
	t.Reducer = f
}

//type Inputs map[string]interface{}

type ResultFunction struct {
	Child int
	//Result reflect.Value
	Result interface{}
}

func (t *CommandTree) RunMergeAsync(_ interface{}) interface{} {
	// Execcute the tree

	channel := make(chan ResultFunction, t.nChildren)
	defer close(channel)

	for i, f := range t.LeafFunctions {
		go wrap(channel, i, f, t.LeafFunctionsInput[i])
	}

	remaining := t.nChildren
	for remaining > 0 {
		result := <-channel
		remaining -= 1
		res := make([]interface{}, 2)
		res[0] = result.Result
		res[1] = result.Child
		t.Reducer(res)
	}
	return nil

}

func (t *CommandTree) Run(_ interface{}) interface{} {
	// Execcute the tree

	channel := make(chan ResultFunction, t.nChildren)
	defer close(channel)
	t.LeafFunctionsOutput = make([]interface{}, t.nChildren)

	for i, f := range t.LeafFunctions {
		go wrap(channel, i, f, t.LeafFunctionsInput[i])
	}

	remaining := t.nChildren
	for remaining > 0 {
		result := <-channel
		remaining -= 1
		t.LeafFunctionsOutput[result.Child] = result.Result
	}
	return t.Reducer(t.LeafFunctionsOutput)

}

func wrap(c chan ResultFunction, child int, todo func(inp interface{}) interface{}, inp interface{}) {

	var result ResultFunction

	startTime := time.Now()
	result.Result = todo(inp)
	endTime := time.Now()
	log.Println("WRAP TOTAL ", endTime.Sub(startTime))
	result.Child = child

	c <- result

}
