package main

import (
	"fmt"
	"sync"
	"testing"
)

// Assuming a GlobalStore and Tx structure defined as follows in main application:
// var GlobalStore sync.Map
// type Tx struct { ... }
// type TxStack struct { ... }
// And TxStack, Tx, Set, Get, Delete methods are defined appropriately for sync.Map

func TestPushTransaction(t *testing.T) {
	ts := &TxStack{}
	ts.Push()

	if ts.size != 1 {
		t.Errorf("Expected stack size of 1 after one push, got %d", ts.size)
	}

	if ts.top == nil {
		t.Errorf("Expected top transaction to be non-nil after push")
	}

	// Test pushing multiple transactions
	for i := 0; i < 5; i++ {
		ts.Push()
	}
	if ts.size != 6 { // already 1 pushed before
		t.Errorf("Expected stack size of 6 after pushing multiple times, got %d", ts.size)
	}
}

func TestPopTransaction(t *testing.T) {
	ts := &TxStack{}
	ts.Push()
	ts.Pop()

	if ts.size != 0 {
		t.Errorf("Expected stack size of 0 after pushing and popping once, got %d", ts.size)
	}

	// Edge case: popping from an empty stack
	ts.Pop() // should not cause error or change in size
	if ts.size != 0 {
		t.Errorf("Popping from an empty stack should leave size unchanged, got %d", ts.size)
	}

	// Test popping after multiple pushes
	for i := 0; i < 5; i++ {
		ts.Push()
	}
	for i := 0; i < 5; i++ {
		ts.Pop()
	}
	if ts.size != 0 {
		t.Errorf("Expected stack size of 0 after pushing and popping multiple times, got %d", ts.size)
	}
}

func TestCommit(t *testing.T) {
	// Reset GlobalStore for testing
	GlobalStore = sync.Map{} // Updated to sync.Map
	ts := &TxStack{}
	ts.Push()

	var key = "key1"
	var value = "value1"
	ts.top.store.Store(key, value) // Correct usage of sync.Map

	ts.Commit()

	// Retrieve value from GlobalStore
	valueFromStore, ok := GlobalStore.Load(key)
	if !ok || valueFromStore != value {
		t.Errorf("Expected GlobalStore to have committed value '%s' for '%s', found %s", value, key, valueFromStore)
	}

	// Edge case: Committing with no active transactions
	ts = &TxStack{} // new stack with no transactions
	ts.Commit()     // should not cause error

	// Ensure GlobalStore is unchanged after committing with no active transactions
	var count int
	GlobalStore.Range(func(k, v interface{}) bool {
		count++
		return true
	})
	if count != 1 { // count should be 1 because one value was added in this test
		t.Errorf("GlobalStore should have only one value after committing with no active transactions, found %d items", count)
	}
}

func TestRaceCondition(t *testing.T) {
	GlobalStore = sync.Map{} // Resetting GlobalStore before the test
	var wg sync.WaitGroup

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(txNum int) {
			defer wg.Done()
			localTs := &TxStack{} // Consider giving each transaction its own TxStack if applicable
			localTs.Push()        // Push a new transaction

			// Simulate updating the transaction's store with unique keys for each transaction
			for j := 0; j < 5; j++ {
				key := fmt.Sprintf("key%d_from_tx%d", j, txNum)
				val := fmt.Sprintf("value%d", j)
				localTs.top.store.Store(key, val)
			}

			localTs.Commit()
		}(i)
	}

	wg.Wait()

	// Check for consistency in GlobalStore
	var count int
	GlobalStore.Range(func(k, v interface{}) bool {
		count++
		return true
	})
	if count < 50 { // Expecting at least 50 entries (10 transactions x 5 keys each)
		t.Errorf("Expected GlobalStore to have many values after concurrent commits, found only %d", count)
	}
}
