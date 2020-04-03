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

var (
	ErrOpeningDev       = errors.New("Error opening device")
	ErrReadingDev       = errors.New("Error reading device")
	ErrSeekingDev       = errors.New("Error seeking device")
	ErrUnexpectedSeek   = errors.New("Unexpected position seeking device")
	ErrMultipleBootable = errors.New("Multiple bootable partitions")
)

type PartitionError struct {
	sentinel error
	wrapped  error
	message  string
}

func (e PartitionError) Error() string {
	return e.message
}

func (e PartitionError) Is(target error) bool {
	return target == e.sentinel
}

func (e PartitionError) Unwrap() error {
	return e.wrapped
}

func Analyze(dev string) (err error) {
	f, err := os.Open(dev)
	if err != nil {
		return &PartitionError{ErrOpeningDev, err,
			fmt.Sprintf("%v %s: %v", ErrOpeningDev, dev, err)}
	}
	defer f.Close()
	t := &PartitionTable{}
	t.Table = make([]PartitionEntry, 0)
	err = t.ParseBootRecord(f, dev, 0)
	if err != nil {
		return err
	}
	//fmt.Println(t)
	index, err := t.GetBootable()
	fmt.Printf("Total partitions: %d\nBootable: %d %v\n%v", len(t.Table),
		index, err, t)

	return nil
}
