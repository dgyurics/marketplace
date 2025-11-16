package utilities

import (
	"sync"
	"testing"
	"time"
)

func resetIDGenerator() {
	idGenerator = nil
	initIDGenerator = sync.Once{}
}

func TestGenerateID_Initialized(t *testing.T) {
	resetIDGenerator()
	InitIDGenerator(1) // initialize with machine ID 1

	id1, err := GenerateID()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if id1 == 0 {
		t.Fatalf("expected a non-zero ID, got %d", id1)
	}
}

func TestGenerateID_Increasing(t *testing.T) {
	resetIDGenerator()
	InitIDGenerator(2) // initialize with machine ID 2

	id1, err1 := GenerateID()
	time.Sleep(time.Millisecond)
	id2, err2 := GenerateID()

	if err1 != nil || err2 != nil {
		t.Fatalf("expected no errors, got %v and %v", err1, err2)
	}
	if id2 <= id1 {
		t.Fatalf("expected id2 > id1, got id1=%d, id2=%d", id1, id2)
	}
}

func TestGenerateID_Uninitialized(t *testing.T) {
	resetIDGenerator()
	idGenerator = nil // manually reset to simulate uninitialized state

	_, err := GenerateID()
	if err == nil {
		t.Fatal("expected an error when generating ID without initialization")
	}
}

func TestGenerateID_UniqueIDsBeyondSequenceLimit(t *testing.T) {
	resetIDGenerator()
	InitIDGenerator(3) // initialize with machine ID 3

	numIDs := 1000
	idSet := make(map[uint64]struct{})

	for i := 0; i < numIDs; i++ {
		id, err := GenerateID()
		if err != nil {
			t.Fatalf("GenerateID failed: %v", err)
		}
		if _, exists := idSet[id]; exists {
			t.Fatalf("duplicate ID generated: %d", id)
		}
		idSet[id] = struct{}{}
	}

	if len(idSet) != numIDs {
		t.Fatalf("expected %d unique IDs, got %d", numIDs, len(idSet))
	}
}

func TestGenerateID_Concurrent(t *testing.T) {
	resetIDGenerator()
	InitIDGenerator(4)

	const numIDs = 10000
	idSet := make(map[uint64]struct{})
	var mu sync.Mutex
	var wg sync.WaitGroup

	wg.Add(numIDs)
	for i := 0; i < numIDs; i++ {
		go func() {
			defer wg.Done()
			id, err := GenerateID()
			if err != nil {
				t.Errorf("GenerateID failed: %v", err)
				return
			}
			mu.Lock()
			defer mu.Unlock()
			if _, exists := idSet[id]; exists {
				t.Errorf("duplicate ID generated: %d", id)
			}
			idSet[id] = struct{}{}
		}()
	}
	wg.Wait()

	if len(idSet) != numIDs {
		t.Fatalf("expected %d unique IDs, got %d", numIDs, len(idSet))
	}
}
