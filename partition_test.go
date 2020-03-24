// Copyright © 2020 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package partition

import (
	"errors"
	"fmt"
	"syscall"
	"testing"
)

func TestNonexistent(t *testing.T) {
	err := Analyze("testdata/non-existent-file")
	fmt.Printf("Error returned is %v\n", err)
	fmt.Printf("Error unwrapped is %v\n", errors.Unwrap(err))
	fmt.Printf("Error unwrapped unwrapped is %v\n",
		errors.Unwrap(errors.Unwrap(err)))
	fmt.Printf("Error unwrapped unwrapped unwrapped is %v\n",
		errors.Unwrap(errors.Unwrap(errors.Unwrap(err))))

	if !errors.Is(err, syscall.ENOENT) {
		t.Error(err, "is not syscall.ENOENT")
	}
	if !errors.Is(err, ErrOpeningDev) {
		t.Error(err, "is not ErrOpeningDev")
	}
}

func TestHybrid(t *testing.T) {
	err := Analyze("testdata/hybrid.dat")
	fmt.Printf("Error returned is %v\n", err)
}
