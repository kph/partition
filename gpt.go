// Copyright © 2020 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package partition

import (
	"encoding/binary"
	"fmt"
	"io"
	"strings"
	"unicode/utf16"
	"unsafe"

	"github.com/satori/uuid"
)

const (
	BlockSize    = 512 // For the time being
	GPTHeaderLBA = 1 * BlockSize
)

var (
	LinuxFSDataPart = uuid.Must(uuid.FromString("0FC63DAF-8483-4772-8E79-3D69D8477DE4"))
	LinuxBootPart   = uuid.Must(uuid.FromString("BC13C2FF-59E6-4262-A352-B275FD6F7172"))
)

type msuuid struct {
	TimeLow          uint32
	TimeMid          uint16
	TimeHiAndVersion uint16
	ClockSeq         [2]byte
	Node             [6]byte
}

func (u msuuid) nonZero() bool {
	if u.TimeLow != 0 || u.TimeMid != 0 || u.TimeHiAndVersion != 0 {
		return true
	}
	return false
}

func (u msuuid) UUID() uuid.UUID {
	// Abandon hope all ye who enter here
	// https://developer.apple.com/library/archive/technotes/tn2166/_index.html#//apple_ref/doc/uid/DTS10003927-CH1-SUBSECTION11
	u.TimeLow = binary.BigEndian.Uint32((*[4]byte)(unsafe.Pointer(&u.TimeLow))[:])
	u.TimeMid = binary.BigEndian.Uint16((*[2]byte)(unsafe.Pointer(&u.TimeMid))[:])
	u.TimeHiAndVersion = binary.BigEndian.Uint16((*[2]byte)(unsafe.Pointer(&u.TimeHiAndVersion))[:])
	uuid, _ := uuid.FromBytes((*[16]byte)(unsafe.Pointer(&u))[:])
	return uuid
}
func (u msuuid) String() string {
	return u.UUID().String()
}

type GPTHeader struct {
	Signature          [8]byte
	MinorVer           uint16
	MajorVer           uint16
	HeaderSize         uint32
	HeaderCRC          uint32
	MBZ                uint32
	CurrentLBA         uint64
	BackupLBA          uint64
	FirstUsableLBA     uint64
	LastUsableLBA      uint64
	UUID               msuuid
	PartitionArrayLBA  uint64
	PartitionCount     uint32
	PartitionEntrySize uint32
	PartitionArrayCRC  uint32
	MBZ2               [92]byte
}

func (g GPTHeader) String() (s string) {
	return fmt.Sprintf("Signature %x Ver %d.%d HeaderSize %04x HeaderCRC %04x CurrentLBA %d BackupLBA %d FirstUsableLBA %d LastUsableLBA %d UUID %v PartitionArrayLBA %d PartitionCount %d PartitionEntrySize %d PartitionArrayCRC %x",
		g.Signature, g.MajorVer, g.MinorVer, g.HeaderSize, g.HeaderCRC,
		g.CurrentLBA, g.BackupLBA, g.FirstUsableLBA, g.LastUsableLBA,
		g.UUID, g.PartitionArrayLBA, g.PartitionCount,
		g.PartitionEntrySize, g.PartitionArrayCRC)
}

type GPTPartitionEntry struct {
	Type     msuuid
	ID       msuuid
	FirstLBA uint64
	LastLBA  uint64
	Flags    uint64
	Name     PartitionName
}

func (g GPTPartitionEntry) IsUsed() bool {
	return g.Type.nonZero()
}

type PartitionName [36]uint16

func (n PartitionName) String() string {
	return strings.Trim(string(utf16.Decode(((*[36]uint16)(&n))[:])), "\x00")
}

func (t *PartitionTable) ParseGPT(f io.ReadSeeker, dev string) (err error) {
	p, err := f.Seek(GPTHeaderLBA, io.SeekStart)
	if err != nil {
		return &PartitionError{ErrSeekingDev, err,
			fmt.Sprintf("%v %s offset %d: %v",
				ErrSeekingDev, dev, GPTHeaderLBA, err)}
	}
	if p != GPTHeaderLBA {
		return &PartitionError{ErrUnexpectedSeek, nil,
			fmt.Sprintf("%v %s offset %d seeked to %d instead",
				ErrUnexpectedSeek, dev, GPTHeaderLBA, p)}
	}
	gpt := GPTHeader{}
	err = binary.Read(f, binary.LittleEndian, &gpt)
	if err != nil {
		return &PartitionError{ErrReadingDev, err,
			fmt.Sprintf("%v %s offset %d: %v",
				ErrReadingDev, dev, GPTHeaderLBA, err)}
	}
	fmt.Println(gpt.String())

	for partNum := uint32(0); partNum < gpt.PartitionCount; partNum++ {
		offset := int64(gpt.PartitionArrayLBA*BlockSize) +
			int64((partNum * gpt.PartitionEntrySize))
		p, err := f.Seek(offset, io.SeekStart)
		if err != nil {
			return &PartitionError{ErrSeekingDev, err,
				fmt.Sprintf("%v %s offset %d: %v",
					ErrSeekingDev, dev, offset, err)}
		}
		if p != offset {
			return &PartitionError{ErrUnexpectedSeek, nil,
				fmt.Sprintf("%v %s offset %d seeked to %d instead",
					ErrUnexpectedSeek, dev, offset, p)}
		}
		part := GPTPartitionEntry{}
		err = binary.Read(f, binary.LittleEndian, &part)
		if err != nil {
			return &PartitionError{ErrReadingDev, err,
				fmt.Sprintf("%v %s offset %d: %v",
					ErrReadingDev, dev, offset, err)}
		}
		if part.IsUsed() {
			t.GPTTable = append(t.GPTTable, part)
		}
		//fmt.Println(part)
	}

	return nil
}
