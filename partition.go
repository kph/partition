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
	ErrOpeningDev     = errors.New("Error opening device")
	ErrReadingDev     = errors.New("Error reading device")
	ErrSeekingDev     = errors.New("Error seeking device")
	ErrUnexpectedSeek = errors.New("Unexpected position seeking device")
)

type PartitionError struct {
	sentinel error
	wrapped  error
	message  string
}

func (e *PartitionError) Error() string {
	return e.message
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

func (c *CHS) IsZero() bool {
	return c.Cyl == 0 && c.Head == 0 && c.Sector == 0
}

type PartitionType byte

const (
	PartitionTypeEmpty = PartitionType(0x00)

	PartitionTypeDOSExtended   = PartitionType(0x05)
	PartitionTypeWin98Extended = PartitionType(0x0f)
	PartitionTypeLinuxExtended = PartitionType(0x85)

	PartitionTypeLinuxSwap = PartitionType(0x82)
	PartitionTypeLinuxData = PartitionType(0x83)
	PartitionTypeLinuxLVM  = PartitionType(0x8e)
	PartitionTypeLinuxRAID = PartitionType(0xfd)
)

func (p PartitionType) String() string {
	switch p {
	case PartitionTypeEmpty:
		return "Empty"

	case PartitionTypeDOSExtended:
		return "DOS Extended"
	case PartitionTypeWin98Extended:
		return "Win98 Extended"
	case PartitionTypeLinuxExtended:
		return "Linux Extended"

	case PartitionTypeLinuxSwap:
		return "Linux Swap"
	case PartitionTypeLinuxData:
		return "Linux Data"
	case PartitionTypeLinuxLVM:
		return "Linux LVM"
	case PartitionTypeLinuxRAID:
		return "Linux RAID"
	}
	return fmt.Sprintf("%02x", int(p))
}

type PartitionStatus byte

const (
	PartitionStatusUnbootable = PartitionStatus(0x00)
	PartitionStatusBootable   = PartitionStatus(0x80)
)

func (s PartitionStatus) String() string {
	switch s {
	case PartitionStatusUnbootable:
		return "Unbootable"
	case PartitionStatusBootable:
		return "Bootable"
	}
	return fmt.Sprintf("Unexpected %02x", int(s))
}

type PartitionEntry struct {
	Status  PartitionStatus
	First   CHS
	Type    PartitionType
	Last    CHS
	Lba     uint32
	Sectors uint32
}

func (p *PartitionEntry) String() string {
	return fmt.Sprintf("%v %v %0v %v %d %d", p.Status, p.First, p.Type,
		p.Last, p.Lba, p.Sectors)
}

func (p *PartitionEntry) IsExtended() bool {
	return p.Type == PartitionTypeDOSExtended ||
		p.Type == PartitionTypeWin98Extended ||
		p.Type == PartitionTypeLinuxExtended
}

func (p *PartitionEntry) IsUsed() bool {
	return p.Status != 0 || !p.First.IsZero() || p.Type != PartitionTypeEmpty ||
		!p.Last.IsZero() || p.Lba != 0 || p.Sectors != 0
}

type BootRecord struct {
	Tbd        [446]byte
	Partitions [4]PartitionEntry
	Signature  uint16
}

func (m *BootRecord) String() (s string) {
	for i := 0; i < 4; i++ {
		s += m.Partitions[i].String() + "\n"
	}
	s += fmt.Sprintf("Signature: %04x", m.Signature)
	return
}

type PartitionTable struct {
	Table []PartitionEntry
}

func (t *PartitionTable) ParseBootRecord(f io.ReadSeeker, dev string, base int64) (err error) {
	p, err := f.Seek(base, io.SeekStart)
	if err != nil {
		return &PartitionError{ErrSeekingDev, err,
			fmt.Sprintf("%v %s offset %d: %v",
				ErrSeekingDev, dev, base, err)}
	}
	if p != base {
		return &PartitionError{ErrUnexpectedSeek, nil,
			fmt.Sprintf("%v %s offset %d seeked to %d instead",
				ErrUnexpectedSeek, dev, base, p)}
	}
	br := BootRecord{}
	err = binary.Read(f, binary.LittleEndian, &br)
	if err != nil {
		return &PartitionError{ErrReadingDev, err,
			fmt.Sprintf("%v %s offset %d: %v",
				ErrReadingDev, dev, base, err)}
	}
	fmt.Println(br.String())

	for i := 0; i < 4; i++ {
		if br.Partitions[i].IsUsed() && !br.Partitions[i].IsExtended() {
			t.Table = append(t.Table, br.Partitions[i])
		} else {
			if base == 0 {
				t.Table = append(t.Table, PartitionEntry{})
			}
		}
	}

	for i := 0; i < 4; i++ {
		if br.Partitions[i].IsExtended() {
			offset := base + (int64(br.Partitions[i].Lba) * 512)
			err = t.ParseBootRecord(f, dev, offset)
			if err != nil {
				return err
			}
		}
	}
	return nil
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
	fmt.Println(t)
	return nil
}
