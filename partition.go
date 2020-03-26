// Copyright © 2020 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

// This is a package to parse partition tables
package partition

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"os"
)

var (
	ErrOpeningDev = errors.New("Error opening device")
	ErrReadingDev = errors.New("Error reading device")
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

type CHS struct {
	Head   byte
	Sector byte // sector in 5-0; 7-6 are high bits of cylinder
	Cyl    byte // bits 7-0 of cylinder
}

func (c *CHS) String() string {
	cyl := c.Cyl + ((c.Sector >> 6) << 8)
	sector := c.Sector & 0x2f
	return fmt.Sprintf("%d/%d/%d", cyl, c.Head, sector)
}

type PartitionEntry struct {
	Status   byte
	First    CHS
	PartType byte
	Last     CHS
	Lba      uint32
	Sectors  uint32
}

func (p *PartitionEntry) String() string {
	return fmt.Sprintf("%02x %v %02x %v %d %d", p.Status, p.First, p.PartType,
		p.Last, p.Lba, p.Sectors)
}

type MBR struct {
	Tbd        [446]byte
	Partitions [4]PartitionEntry
	Signature  uint16
}

func (m *MBR) String() (s string) {
	for i := 0; i < 4; i++ {
		s += m.Partitions[i].String() + "\n"
	}
	s += fmt.Sprintf("Signature: %04x", m.Signature)
	return
}

func findEBR(f io.ReadSeeker, dev string, base int64) (err error) {
	p, err := f.Seek(base, io.SeekStart)
	if err != nil {
		panic(err)
	}
	if p != base {
		panic("seeked to the wrong place")
	}
	mbr := MBR{}
	err = binary.Read(f, binary.LittleEndian, &mbr)
	if err != nil {
		return &PartitionError{ErrReadingDev, err, dev}
	}
	fmt.Println(mbr.String())

	for i := 0; i < 4; i++ {
		if mbr.Partitions[i].PartType == 0x05 {
			offset := base + (int64(mbr.Partitions[i].Lba) * 512)
			findEBR(f, dev, offset)
		}
	}
	return nil
}

func Analyze(dev string) (err error) {
	f, err := os.Open(dev)
	if err != nil {
		return &PartitionError{ErrOpeningDev, err, dev}
	}
	defer f.Close()
	return findEBR(f, dev, 0)
}
