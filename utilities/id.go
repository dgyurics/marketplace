package utilities

import (
	"errors"
	"fmt"
	"strconv"
	"sync"
	"time"
)

// IDGenerator generates unique 64-bit IDs based on timestamp, machine ID, and sequence number.
// The ID format is:
//   - 47 bits: timestamp (milliseconds since Unix epoch)
//   - 8  bits: machine ID (0-255)
//   - 8  bits: sequence ID (0-254)
//
// The most significant bit is unused
type IDGenerator struct {
	machineID     uint8
	seqID         uint8
	lastTimestamp int64
	mu            sync.Mutex
}

var (
	idGenerator *IDGenerator
	initOnce    sync.Once
)

func InitIDGenerator(machineID uint8) {
	initOnce.Do(func() {
		idGenerator = &IDGenerator{
			machineID: machineID,
		}
	})
}

// GenerateID returns a unique 64-bit ID composed of the current timestamp, machine ID, and sequence ID.
// It ensures IDs are unique within the same millisecond. The most significant bit is unused.
func GenerateID() (uint64, error) {
	if idGenerator == nil {
		return 0, errors.New("ID generator not initialized")
	}

	idGenerator.mu.Lock()
	defer idGenerator.mu.Unlock()

	timestamp := time.Now().UTC().UnixMilli()

	// Check for clock skew
	if timestamp < idGenerator.lastTimestamp {
		return 0, fmt.Errorf("clock moved backwards: current %d, last %d", timestamp, idGenerator.lastTimestamp)
	}

	if idGenerator.lastTimestamp != timestamp {
		idGenerator.seqID = 0
		idGenerator.lastTimestamp = timestamp
	} else {
		idGenerator.seqID++
		if idGenerator.seqID > 254 {
			// Wait for the next millisecond
			time.Sleep(time.Millisecond)
			timestamp = time.Now().UTC().UnixMilli()
			idGenerator.seqID = 0
			idGenerator.lastTimestamp = timestamp
		}
	}

	// Construct ID: | 47-bit timestamp (bits 16-62) | 8-bit machine ID | 8-bit sequence ID |
	var id uint64
	id = uint64(timestamp) << 16
	id |= uint64(idGenerator.machineID) << 8
	id |= uint64(idGenerator.seqID)
	return id, nil
}

func GenerateIDString() (string, error) {
	id, err := GenerateID()
	if err != nil {
		return "", err
	}
	return strconv.FormatUint(id, 10), nil
}

func DecodeID(id uint64) (timestamp time.Time, machineID, seqID uint8) {
	seqID = uint8(id & 0xFF)            // bits 0-7
	machineID = uint8((id >> 8) & 0xFF) // bits 8-15
	timestampMillis := int64(id >> 16)  // bits 16-63
	timestamp = time.UnixMilli(timestampMillis)
	return
}
