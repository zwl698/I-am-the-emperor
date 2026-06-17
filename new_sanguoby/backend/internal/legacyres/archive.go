package legacyres

import (
	"encoding/binary"
	"errors"
	"fmt"
	"math"
	"os"
)

const (
	resourceHeaderSize = 12
	resourceIndexSize  = 4
)

var (
	ErrResourceNotFound = errors.New("legacy resource not found")
	ErrItemOutOfRange   = errors.New("legacy resource item out of range")
	ErrInvalidArchive   = errors.New("invalid legacy archive")
)

type Header struct {
	Address    uint32 `json:"address"`
	Length     uint32 `json:"length"`
	ID         uint16 `json:"id"`
	ItemCount  uint16 `json:"itemCount"`
	ItemLength uint16 `json:"itemLength"`
	Key        byte   `json:"key"`
	Reserved   byte   `json:"reserved"`
}

// PersonStructSize is the per-person record size in scenario resource 61.
// Verified empirically: the 0x64 (=100, max loyalty) anchor repeats every 15
// bytes, and decoding at stride 15 yields the correct 董卓-era roster
// (index 0 = 董卓, 1 = 曹操, 2 = 袁绍 ...). This is the compact scenario record,
// not the larger in-memory C PersonType struct.
const PersonStructSize = 15

// CityStructSize is the per-city record size in scenario resource 57.
// Verified empirically: the index byte (column 0) increments cleanly 0..37
// at stride 31, and 38 cities × 31 + 2 (start year U16) = 1180 bytes matches
// the resource 57 item length exactly.
const CityStructSize = 31

type Archive struct {
	path string
	data []byte
}

func Open(path string) (*Archive, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	if len(data) < resourceHeaderSize {
		return nil, fmt.Errorf("%w: file too small", ErrInvalidArchive)
	}
	return &Archive{path: path, data: data}, nil
}

func (a *Archive) Path() string {
	return a.path
}

func (a *Archive) Resource(id uint16) (Header, error) {
	if id == 0 {
		return Header{}, ErrResourceNotFound
	}

	indexOffset := int(id-1) * 4
	if indexOffset+4 > len(a.data) {
		return Header{}, fmt.Errorf("%w: resource %d index outside file", ErrResourceNotFound, id)
	}

	address := binary.LittleEndian.Uint32(a.data[indexOffset : indexOffset+4])
	if address == 0 || address == math.MaxUint32 {
		return Header{}, fmt.Errorf("%w: resource %d", ErrResourceNotFound, id)
	}

	headerOffset := int(address)
	if headerOffset < 0 || headerOffset+resourceHeaderSize > len(a.data) {
		return Header{}, fmt.Errorf("%w: resource %d header outside file", ErrInvalidArchive, id)
	}

	header := Header{
		Address:    address,
		Length:     binary.LittleEndian.Uint32(a.data[headerOffset : headerOffset+4]),
		ID:         binary.LittleEndian.Uint16(a.data[headerOffset+4 : headerOffset+6]),
		ItemCount:  binary.LittleEndian.Uint16(a.data[headerOffset+6 : headerOffset+8]),
		ItemLength: binary.LittleEndian.Uint16(a.data[headerOffset+8 : headerOffset+10]),
		Key:        a.data[headerOffset+10],
		Reserved:   a.data[headerOffset+11],
	}
	if header.ID != id {
		return Header{}, fmt.Errorf("%w: index %d points to resource %d", ErrInvalidArchive, id, header.ID)
	}
	return header, nil
}

func (a *Archive) List(maxID uint16) []Header {
	resources := make([]Header, 0)
	for id := uint16(1); id <= maxID; id++ {
		header, err := a.Resource(id)
		if err == nil {
			resources = append(resources, header)
		}
	}
	return resources
}

func (a *Archive) Item(resourceID uint16, itemIndex uint16) ([]byte, error) {
	header, err := a.Resource(resourceID)
	if err != nil {
		return nil, err
	}
	if itemIndex == 0 || itemIndex > header.ItemCount {
		return nil, fmt.Errorf("%w: resource %d item %d", ErrItemOutOfRange, resourceID, itemIndex)
	}

	itemOffset, itemLength, err := a.itemBounds(header, itemIndex)
	if err != nil {
		return nil, err
	}

	start := int(header.Address) + int(itemOffset)
	end := start + int(itemLength)
	if start < 0 || end < start || end > len(a.data) {
		return nil, fmt.Errorf("%w: resource %d item %d outside file", ErrInvalidArchive, resourceID, itemIndex)
	}

	item := append([]byte(nil), a.data[start:end]...)
	if header.Key != 0 {
		for i := range item {
			item[i] -= header.Key
		}
	}
	return item, nil
}

// RawItem returns the item data WITHOUT decryption.
// Use this for resources where the "encryption" is actually a different encoding scheme
// (e.g., resource 58 city names store raw GBK bytes that should not be XOR-decrypted).
func (a *Archive) RawItem(resourceID uint16, itemIndex uint16) ([]byte, error) {
	header, err := a.Resource(resourceID)
	if err != nil {
		return nil, err
	}
	if itemIndex == 0 || itemIndex > header.ItemCount {
		return nil, fmt.Errorf("%w: resource %d item %d", ErrItemOutOfRange, resourceID, itemIndex)
	}

	itemOffset, itemLength, err := a.itemBounds(header, itemIndex)
	if err != nil {
		return nil, err
	}

	start := int(header.Address) + int(itemOffset)
	end := start + int(itemLength)
	if start < 0 || end < start || end > len(a.data) {
		return nil, fmt.Errorf("%w: resource %d item %d outside file", ErrInvalidArchive, resourceID, itemIndex)
	}

	return append([]byte(nil), a.data[start:end]...), nil
}

func (a *Archive) itemBounds(header Header, itemIndex uint16) (uint16, uint16, error) {
	if header.ItemLength != 0 {
		return uint16(resourceHeaderSize) + (itemIndex-1)*header.ItemLength, header.ItemLength, nil
	}

	if header.ItemCount == 1 {
		if header.Length < resourceHeaderSize {
			return 0, 0, fmt.Errorf("%w: resource %d length shorter than header", ErrInvalidArchive, header.ID)
		}
		return resourceHeaderSize, uint16(header.Length - resourceHeaderSize), nil
	}

	indexOffset := int(header.Address) + resourceHeaderSize + int(itemIndex-1)*resourceIndexSize
	if indexOffset+resourceIndexSize > len(a.data) {
		return 0, 0, fmt.Errorf("%w: resource %d item %d index outside file", ErrInvalidArchive, header.ID, itemIndex)
	}
	index := a.data[indexOffset : indexOffset+resourceIndexSize]
	return binary.LittleEndian.Uint16(index[0:2]), binary.LittleEndian.Uint16(index[2:4]), nil
}
