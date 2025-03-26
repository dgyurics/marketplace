package utilities

import (
	"errors"
	"sync"
)

const (
	maxMachineId = 0
)

type IDGenerator struct {
	machineID uint16
}

var (
	defaultIDGenerator *IDGenerator
	initOnce           sync.Once
)

func InitIDGenerator(machineID uint16) error {
	var err error
	initOnce.Do(func() {
		if machineID > maxMachineId {
			err = errors.New("machine ID is too large")
			return
		}
		defaultIDGenerator = &IDGenerator{
			machineID: machineID,
		}
	})
	return err
}

func GenerateID() (int64, error) {
	if defaultIDGenerator == nil {
		return 0, errors.New("IDGenerator not initialized")
	}
	return 0, nil
}
