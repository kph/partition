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
	SentinelErr *sentinelError
	Dev         string
}

type sentinelError struct {
	e error
	w *wrappedError
}

type wrappedError struct {
	e error
}

func (e *PartitionError) Error() string {
	return fmt.Sprintf("%v %s: %v", e.SentinelErr, e.Dev, e.SentinelErr.w)
}

func (e *PartitionError) Unwrap() error {
	return e.SentinelErr
}

func (e *sentinelError) Error() string {
	return e.e.Error()
}

func (e *sentinelError) Is(target error) bool {
	return target == e.e
}

func (e *sentinelError) Unwrap() error {
	return e.w
}

func (e *wrappedError) Error() string {
	return e.e.Error()
}

func (e *wrappedError) Unwrap() error {
	return errors.Unwrap(e.e)
}

func Analyze(dev string) (err error) {
	f, err := os.Open(dev)
	if err != nil {
		return &PartitionError{
			SentinelErr: &sentinelError{ErrOpeningDev,
				&wrappedError{err}},
			Dev: dev}
	}
	defer f.Close()
	cnt, err := f.Read(make([]byte, blockSize))
	if err != nil {
		return &PartitionError{
			SentinelErr: &sentinelError{ErrReadingDev,
				&wrappedError{err}},
			Dev: dev}
	}
	if cnt != blockSize {
		return fmt.Errorf("%w %s: expected %d got %d", ErrReadCount,
			dev, blockSize, cnt)
	}
	return nil
}
