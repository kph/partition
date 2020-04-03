// Copyright © 2020 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package partition

type GPTHeader struct {
	Signature          [8]byte
	MinorVer           uint16
	MajorVer           uint16
	HeaderSize         uint32
	HeaderCRC          uint32
	MBZ                [8]byte
	CurrentLBA         uint64
	BackupLBA          uint64
	FirstUsableLBA     uint64
	LastUsableLBA      uint64
	UUID               [16]byte
	PartitionArrayLBA  uint64
	PartitionCount     uint32
	PartitionEntrySize uint32
	PartitionArrayCRC  uint32
	MBZ2               [92]byte
}

type GPTPartitonEntry struct {
	Type     [16]byte
	ID       [16]byte
	FirstLBA uint64
	LastLBA  uint64
	Flags    uint64
	Name     [36]uint16
}
