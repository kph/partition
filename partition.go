// Copyright © 2020 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

// This is a package to parse partition tables
package partition

import (
	"errors"
	"fmt"
	"os"
)

const (
	blockSize = 512
)

var (
	ErrOpeningDev = errors.New("Error opening device")
	ErrReadingDev = errors.New("Error reading device")
	ErrReadCount  = errors.New("Read count mismatch on device")
)

type PartitionError struct {
	sentinel error
	wrapped  error
	Dev      string
}

func (e *PartitionError) Error() string {
	return fmt.Sprintf("%v %s: %v", e.sentinel, e.Dev, e.wrapped)
}

func (e *PartitionError) Is(target error) bool {
	return target == e.sentinel
}

func (e *PartitionError) Unwrap() error {
	return e.wrapped
}

func Analyze(dev string) (err error) {
	f, err := os.Open(dev)
	if err != nil {
		return &PartitionError{ErrOpeningDev, err, dev}
	}
	defer f.Close()
	cnt, err := f.Read(make([]byte, blockSize))
	if err != nil {
		return &PartitionError{ErrReadingDev, err, dev}
	}
	if cnt != blockSize {
		return fmt.Errorf("%w %s: expected %d got %d", ErrReadCount,
			dev, blockSize, cnt)
	}
	return nil
}
