package legacyres

// Decoding layer for legacy C binary data structures from dat.lib.
//
// Struct layouts match the original C source (attribute.h) with packed,
// little-endian fields. Key sizes (verified against real dat.lib data):
//   - CityType:  37 bytes per city  (no trailing padding, size is even)
//   - PersonType: 20 bytes per person (19 bytes fields + 1 byte trailing padding)
//   - City names: 10 bytes, resource 58 key=192 IS real subtract-decryption,
//     after which the bytes are GBK and must be transcoded to UTF-8.

import (
	"encoding/binary"
	"errors"
	"fmt"
	"strings"

	"golang.org/x/text/encoding/simplifiedchinese"
)

var (
	ErrCityDecode   = errors.New("legacy city decode error")
	ErrPersonDecode = errors.New("legacy person decode error")
	ErrNameDecode   = errors.New("legacy name decode error")
	ErrShortBuffer  = errors.New("buffer too short")
)

// ---- City scenario record (31 bytes) ----
//
// Resource 57 stores one item per period; each item is CITY_MAX (38) city
// records of 31 bytes followed by a U16 start year. The on-disk scenario layout
// is more compact than the in-memory C CityType struct. Field offsets were
// verified empirically against the 董卓 scenario (city[0]=西凉, city[10]=长安),
// validating民忠/防灾∈[0,100], 金钱/粮食∈[0,30000], and population as a U32:
//
//   0     Index            u8   城池索引 (0-based)
//   1     Belong           u8   归属 (ruler person index + 1; 0 = neutral)
//   2     SatrapID         u8   太守索引
//   3-4   FarmingLimit     u16  农业上限
//   5-6   Farming          u16  农业开发度
//   7-8   CommerceLimit    u16  商业上限
//   9-10  Commerce         u16  商业开发度
//   11    PeopleDevotion   u8   民忠 0-100
//   12    AvoidCalamity    u8   防灾 0-100
//   13-16 PopulationLimit  u32  人口上限
//   17-20 Population       u32  人口
//   21-22 Money            u16  金钱
//   23-24 Food             u16  粮食
//   25-30 reserved/queue indices

type LegacyCity struct {
	Index           uint8 // 城池索引 (0-based)
	Belong          uint8 // 归属 (ruler index + 1; 0 = neutral)
	SatrapID        uint8 // 太守索引
	FarmingLimit    uint16
	Farming         uint16
	CommerceLimit   uint16
	Commerce        uint16
	PeopleDevotion  uint8 // 民忠
	AvoidCalamity   uint8 // 防灾
	PopulationLimit uint32
	Population      uint32
	Money           uint16
	Food            uint16
}

func DecodeCity(raw []byte) (*LegacyCity, int, error) {
	if len(raw) < CityStructSize {
		return nil, 0, fmt.Errorf("%w: got %d want %d", ErrShortBuffer, len(raw), CityStructSize)
	}
	c := &LegacyCity{
		Index:           raw[0],
		Belong:          raw[1],
		SatrapID:        raw[2],
		FarmingLimit:    binary.LittleEndian.Uint16(raw[3:5]),
		Farming:         binary.LittleEndian.Uint16(raw[5:7]),
		CommerceLimit:   binary.LittleEndian.Uint16(raw[7:9]),
		Commerce:        binary.LittleEndian.Uint16(raw[9:11]),
		PeopleDevotion:  raw[11],
		AvoidCalamity:   raw[12],
		PopulationLimit: binary.LittleEndian.Uint32(raw[13:17]),
		Population:      binary.LittleEndian.Uint32(raw[17:21]),
		Money:           binary.LittleEndian.Uint16(raw[21:23]),
		Food:            binary.LittleEndian.Uint16(raw[23:25]),
	}
	return c, CityStructSize, nil
}

func DecodeCities(raw []byte, maxCities int) ([]*LegacyCity, error) {
	cities := make([]*LegacyCity, 0, maxCities)
	offset := 0
	for i := 0; i < maxCities && offset+CityStructSize <= len(raw); i++ {
		city, n, err := DecodeCity(raw[offset:])
		if err != nil {
			return nil, fmt.Errorf("%w: city %d: %v", ErrCityDecode, i, err)
		}
		cities = append(cities, city)
		offset += n
	}
	return cities, nil
}

// DecodeCitiesWithNames decodes all scenario cities from resource 57 for the
// requested period (1-4) and pairs each with its name from resource 58.
func (a *Archive) DecodeCitiesWithNames() ([]*LegacyCity, []string, error) {
	return a.DecodeCitiesWithNamesForPeriod(1)
}

func (a *Archive) DecodeCitiesWithNamesForPeriod(period uint16) ([]*LegacyCity, []string, error) {
	raw, err := a.Item(57, period)
	if err != nil {
		return nil, nil, fmt.Errorf("cities resource: %w", err)
	}
	cities, err := DecodeCities(raw, len(raw)/CityStructSize)
	if err != nil {
		return nil, nil, err
	}
	allNames, err := a.DecodeAllCityNames()
	if err != nil {
		return nil, nil, err
	}
	names := make([]string, len(cities))
	for i, c := range cities {
		if int(c.Index) < len(allNames) {
			names[i] = allNames[c.Index]
		}
	}
	return cities, names, nil
}

// ---- Person record (15 bytes, scenario format) ----
//
// The scenario person record in resource 61 is a compact 15-byte layout
// (verified by decoding the 董卓 scenario: index 0 = 董卓, index 1 = 曹操, ...).
// This differs from the in-memory C PersonType struct; the scenario stores only
// the authored initial attributes. Field offsets within each 15-byte record:
//
//   0  Index     u8  武将索引 (0-based)
//   1  NameItem  u8  名字资源项号 (1-based = Index+1)
//   2  Level     u8  等级
//   3  Force     u8  武力 1-100
//   4  IQ        u8  智力 1-100
//   5  Devotion  u8  忠诚 0-100
//   6  Character u8  性格 0-4
//   7  Experience u8 经验
//   8  Thew      u8  体力 (scenario stores 0; runtime resets to 100)
//   9  ArmsType  u8  兵种 0-5
//   10 Equip0    u8  装备1
//   11 Equip1    u8  装备2
//   12 Reserved0 u8
//   13 Reserved1 u8
//   14 Age       u8  年龄

type LegacyPerson struct {
	Index      uint8 // 武将索引 (0-based)
	NameItem   uint8 // 名字资源项号 (1-based)
	Level      uint8
	Force      uint8    // 武力
	IQ         uint8    // 智力
	Devotion   uint8    // 忠诚
	Character  uint8    // 性格 0-4
	Experience uint8    // 经验
	Thew       uint8    // 体力
	ArmsType   uint8    // 兵种 0-5
	Equip      [2]uint8 // 装备道具
	Age        uint8    // 年龄
}

func DecodePerson(raw []byte) (*LegacyPerson, int, error) {
	if len(raw) < PersonStructSize {
		return nil, 0, fmt.Errorf("%w: got %d want %d", ErrShortBuffer, len(raw), PersonStructSize)
	}
	p := &LegacyPerson{
		Index:      raw[0],
		NameItem:   raw[1],
		Level:      raw[2],
		Force:      raw[3],
		IQ:         raw[4],
		Devotion:   raw[5],
		Character:  raw[6],
		Experience: raw[7],
		Thew:       raw[8],
		ArmsType:   raw[9],
		Equip:      [2]uint8{raw[10], raw[11]},
		Age:        raw[14],
	}
	return p, PersonStructSize, nil
}

func DecodePersons(raw []byte) ([]*LegacyPerson, error) {
	persons := make([]*LegacyPerson, 0)
	offset := 0
	for offset+PersonStructSize <= len(raw) {
		p, n, err := DecodePerson(raw[offset:])
		if err != nil {
			return nil, fmt.Errorf("%w: person at offset %d: %v", ErrPersonDecode, offset, err)
		}
		persons = append(persons, p)
		offset += n
	}
	return persons, nil
}

// DecodePersonName reads a single general name from resource 62 (period 1).
// Names are 8-byte GBK items, 1-based, with key subtract decryption applied by
// Archive.Item, then transcoded to UTF-8.
func (a *Archive) DecodePersonName(nameItem uint8) (string, error) {
	return a.DecodePersonNameForPeriod(1, nameItem)
}

func (a *Archive) DecodePersonNameForPeriod(period uint16, nameItem uint8) (string, error) {
	item, err := a.Item(personNameResourceID(period), uint16(nameItem))
	if err != nil {
		return "", err
	}
	return DecodeGBKName(padTo(item, 10))
}

// DecodePersonsWithNames decodes all scenario persons from resource 61 item 1
// and pairs each with its period-appropriate name resource.
func (a *Archive) DecodePersonsWithNames() ([]*LegacyPerson, []string, error) {
	return a.DecodePersonsWithNamesForPeriod(1)
}

func (a *Archive) DecodePersonsWithNamesForPeriod(period uint16) ([]*LegacyPerson, []string, error) {
	raw, err := a.Item(61, period)
	if err != nil {
		return nil, nil, fmt.Errorf("persons resource: %w", err)
	}
	persons, err := DecodePersons(raw)
	if err != nil {
		return nil, nil, err
	}
	names := make([]string, len(persons))
	for i, p := range persons {
		name, nerr := a.DecodePersonNameForPeriod(period, p.NameItem)
		if nerr != nil {
			names[i] = ""
			continue
		}
		names[i] = name
	}
	return persons, names, nil
}

func personNameResourceID(period uint16) uint16 {
	switch period {
	case 2:
		return 70
	case 3:
		return 71
	case 4:
		return 72
	default:
		return 62
	}
}

// padTo right-pads (or truncates) a byte slice to length n with zeros.
func padTo(b []byte, n int) []byte {
	if len(b) >= n {
		return b[:n]
	}
	out := make([]byte, n)
	copy(out, b)
	return out
}

// ---- City Name (10 bytes raw GBK, NO decryption needed) ----

// DecodeGBKName decodes a 10-byte city name buffer that has ALREADY been
// decrypted by Archive.Item (subtract key 192). The decrypted bytes are GBK
// and are transcoded to UTF-8 here. Null/0xCC padding is trimmed.
func DecodeGBKName(decrypted []byte) (string, error) {
	if len(decrypted) < 10 {
		return "", fmt.Errorf("%w: got %d want 10 for GBK name", ErrShortBuffer, len(decrypted))
	}

	// Trim at the first null terminator; original names are null/0xCC padded.
	end := len(decrypted)
	for i := 0; i < len(decrypted); i++ {
		if decrypted[i] == 0 {
			end = i
			break
		}
	}
	gbk := decrypted[:end]
	if len(gbk) == 0 {
		return "", fmt.Errorf("%w: decoded empty name", ErrNameDecode)
	}

	utf8Bytes, err := simplifiedchinese.GBK.NewDecoder().Bytes(gbk)
	if err != nil {
		return "", fmt.Errorf("%w: GBK transcode: %v", ErrNameDecode, err)
	}
	name := strings.TrimSpace(string(utf8Bytes))
	if name == "" {
		return "", fmt.Errorf("%w: empty after transcode", ErrNameDecode)
	}
	return name, nil
}

// DecodeAllCityNames reads all city name items from resource 58.
// Archive.Item applies the key=192 subtract decryption, then DecodeGBKName
// transcodes the resulting GBK bytes to UTF-8.
func (a *Archive) DecodeAllCityNames() ([]string, error) {
	header, err := a.Resource(58)
	if err != nil {
		return nil, fmt.Errorf("city names resource: %w", err)
	}
	names := make([]string, 0, header.ItemCount)
	for i := uint16(1); i <= header.ItemCount; i++ {
		item, err := a.Item(58, i) // Item applies key=192 decryption
		if err != nil {
			return nil, fmt.Errorf("city name item %d: %w", i, err)
		}
		name, err := DecodeGBKName(item)
		if err != nil {
			names = append(names, "")
			continue
		}
		names = append(names, name)
	}
	return names, nil
}

// ArmsTypeNames maps numeric arms type to Chinese labels.
var ArmsTypeNames = map[uint8]string{
	0: "骑兵",
	1: "步兵",
	2: "弓箭兵",
	3: "水军",
	4: "极兵",
	5: "玄兵",
}

// CharacterNames maps numeric character trait to Chinese labels.
var CharacterNames = map[uint8]string{
	0: "卤莽",
	1: "怕死",
	2: "贪财",
	3: "大志",
	4: "忠义",
}

// CityStateNames map numeric state to Chinese labels.
var CityStateNames = map[uint8]string{
	0: "正常",
	1: "饥荒",
	2: "旱灾",
	3: "水灾",
	4: "暴动",
}
