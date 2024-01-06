package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// GlobalStore holds the global store values
var GlobalStore = make(map[string]string)

// store is map of string:string
type store = map[string]string

// Tx points to a key:value store
type Tx struct {
	next  *Tx
	store store // every transaction will have its local store
}

// TxStack is a list of active and closed transactions
type TxStack struct {
	top  *Tx
	size int
}

// Push creates a new active transaction
func (ts *TxStack) Push() {
	temp := Tx{store: make(store)}
	temp.next = ts.top
	ts.top = &temp
	ts.size++
}

// Pop deletes a transaction from stack
func (ts *TxStack) Pop() {
	// Pop the transaction from stack
	if ts.top == nil {
		fmt.Println("No Active Transactions")
	} else {
		ts.top = ts.top.next
		ts.size--
	}
}

// Commit writes changes to the store within TxStack
func (ts *TxStack) Commit() {
	activeTx := ts.Peek()
	if activeTx != nil {
		for key, val := range activeTx.store {
			GlobalStore[key] = val
			if activeTx.next != nil {
				// update the parent transaction
				activeTx.next.store[key] = val
			}
		}
	} else {
		fmt.Println("Nothing to commit")
	}
}

// Peek returns the active transaction
func (ts *TxStack) Peek() *Tx {
	return ts.top
}

// Get value of key from Store
func Get(key string, txStack *TxStack) {
	activeTx := txStack.Peek()
	if activeTx == nil {
		if val, ok := GlobalStore[key]; ok {
			fmt.Println(val)
		} else {
			fmt.Println(key, "not set")
		}
	} else {
		if val, ok := activeTx.store[key]; ok {
			fmt.Println(val)
		} else {
			fmt.Printf("Key not found %s\n", key)
		}
	}
}

// Set key, value to store
func Set(key string, value string, txStack *TxStack) {
	activeTx := txStack.Peek()
	if activeTx == nil {
		GlobalStore[key] = value
	} else {
		activeTx.store[key] = value
	}
}

// Delete key from Store
func Delete(key string, txStack *TxStack) {
	activeTx := txStack.Peek()
	if activeTx == nil {
		delete(GlobalStore, key)
	} else {
		delete(activeTx.store, key)
		for activeTx.next != nil {
			delete(activeTx.next.store, key)
			activeTx = activeTx.next
			if activeTx.next == nil {
				break
			}
		}
	}
}

// Driver function
func main() {
	r := bufio.NewReader(os.Stdin)
	elements := &TxStack{}
	for {
		fmt.Print("> ")
		text, _ := r.ReadString('\n')
		operation := strings.Fields(text)
		if len(operation) == 0 {
			continue
		}
		switch strings.ToUpper(operation[0]) {
		case "READ":
			// prevents panic
			if len(operation) == 1 {
				printError(true, operation[0])
				continue
			}
			Get(operation[1], elements)
		case "WRITE":
			// prevents panic
			if len(operation) < 3 {
				printError(true, operation[0])
				continue
			}
			Set(operation[1], operation[2], elements)
		case "DELETE":
			// prevents panic
			if len(operation) < 2 {
				printError(true, operation[0])
				continue
			}
			Delete(operation[1], elements)
		case "START":
			// prevents panic
			if len(operation) > 1 {
				printError(false, operation[0])
				continue
			}
			elements.Push()
		case "COMMIT":
			// prevents panic
			if len(operation) > 1 {
				printError(false, operation[0])
				continue
			}
			elements.Commit()
			elements.Pop()
		case "ABORT":
			// prevents panic
			if len(operation) > 1 {
				printError(false, operation[0])
				continue
			}
			elements.Pop()
		case "QUIT":
			// prevents panic
			if len(operation) > 1 {
				printError(false, operation[0])
				continue
			}
			fmt.Println("Exiting...")
			os.Exit(0)
		default:
			fmt.Println("Invalid Command:", operation[0], "\n", "Valid commands are: READ, WRITE, DELETE, START, COMMIT, ABORT, QUIT")
		}
	}
}

func printError(required bool, operation string) {
	if required {
		fmt.Printf("Maybe `%s` missing an argument?\n", operation)
	} else {
		fmt.Printf("`%s` doesn't require an argument!\n", operation)
	}
}
