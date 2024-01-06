package main

import (
	"fmt"
	"sync"
	"testing"
)

// TestPushTransaction checks if transactions are correctly pushed onto the stack
func TestPushTransaction(t *testing.T) {
	ts := &TxStack{}
	ts.Push()

	if ts.size != 1 {
		t.Errorf("Expected stack size of 1 after one push, got %d", ts.size)
	}

	if ts.top == nil || ts.top.store == nil {
		t.Errorf("Expected top transaction to be non-nil with initialized store")
	}

	// Test pushing multiple transactions
	for i := 0; i < 5; i++ {
		ts.Push()
	}
	if ts.size != 6 { // already 1 pushed before
		t.Errorf("Expected stack size of 6 after pushing multiple times, got %d", ts.size)
	}
}

// TestPopTransaction checks popping transactions from the stack, including edge cases
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

// TestCommit checks committing a transaction and its effect on the GlobalStore
func TestCommit(t *testing.T) {
	// Reset GlobalStore for testing
	GlobalStore = make(map[string]string)
	ts := &TxStack{}
	ts.Push()
	ts.top.store["key1"] = "value1"

	ts.Commit()

	if GlobalStore["key1"] != "value1" {
		t.Errorf("Expected GlobalStore to have committed value 'value1' for 'key1', found %s", GlobalStore["key1"])
	}

	// Edge case: Committing with no active transactions
	ts = &TxStack{} // new stack with no transactions
	ts.Commit()     // should not cause error

	// Ensure GlobalStore is unchanged after committing with no active transactions
	if len(GlobalStore) != 1 {
		t.Errorf("GlobalStore should be unchanged when committing with no active transactions, found %d items", len(GlobalStore))
	}
}

// TestRaceCondition tests for race conditions in transaction operations
func TestRaceCondition(t *testing.T) {
	GlobalStore = make(map[string]string) // Resetting GlobalStore before the test
	ts := &TxStack{}
	var wg sync.WaitGroup

	// Define a mutex to synchronize access to the GlobalStore in the test
	var mu sync.Mutex

	// Run several concurrent transactions
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			ts.Push() // Push a new transaction

			// Locking the store to simulate transactional update
			mu.Lock()
			// Simulate updating the transaction's store with unique keys for each transaction
			for j := 0; j < 5; j++ { // each transaction changes 5 values
				key := fmt.Sprintf("key%d", j) // unique keys across transactions
				val := fmt.Sprintf("value%d_from_tx%d", j, i)
				ts.top.store[key] = val
			}
			mu.Unlock()

			// Committing the transaction should affect the GlobalStore
			ts.Commit()
		}(i)
	}

	wg.Wait() // Wait for all goroutines to complete

	// Check for consistency in GlobalStore
	// The store should have multiple keys set by the transactions
	if len(GlobalStore) < 5 {
		t.Errorf("Expected GlobalStore to have many values after concurrent commits, found only %d", len(GlobalStore))
	}
}
