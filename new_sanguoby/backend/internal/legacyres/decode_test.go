package legacyres

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func decodeTestArchivePath(t *testing.T) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("locate test file")
	}
	path := filepath.Clean(filepath.Join(filepath.Dir(file), "..", "..", "..", "..", "sanguobaye_c-master", "src", "dat.lib.orig"))
	if _, err := os.Stat(path); err != nil {
		t.Skipf("legacy archive not found: %v", err)
	}
	return path
}

// ---- City Decoding Tests ----

func TestDecodeCityFromResource57(t *testing.T) {
	archive, err := Open(decodeTestArchivePath(t))
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}

	raw, err := archive.Item(57, 1)
	if err != nil {
		t.Fatalf("Item(57, 1) error = %v", err)
	}
	if len(raw) < CityStructSize {
		t.Fatalf("city data too short: %d bytes", len(raw))
	}

	city, n, err := DecodeCity(raw)
	if err != nil {
		t.Fatalf("DecodeCity error = %v", err)
	}
	if n != CityStructSize {
		t.Fatalf("decoded size = %d, want %d", n, CityStructSize)
	}
	t.Logf("City[0]: Index=%d Belong=%d SatrapID=%d Farm=%d/%d Com=%d/%d Pop=%d/%d Money=%d Food=%d Dev=%d Cal=%d",
		city.Index, city.Belong, city.SatrapID,
		city.Farming, city.FarmingLimit, city.Commerce, city.CommerceLimit,
		city.Population, city.PopulationLimit, city.Money, city.Food,
		city.PeopleDevotion, city.AvoidCalamity)

	if city.Index != 0 {
		t.Errorf("first city Index=%d, want 0", city.Index)
	}
	if city.PeopleDevotion > 100 || city.AvoidCalamity > 100 {
		t.Errorf("devotion/calamity out of range: %d/%d", city.PeopleDevotion, city.AvoidCalamity)
	}
	if city.Money > 30000 || city.Food > 30000 {
		t.Errorf("money/food out of range: %d/%d", city.Money, city.Food)
	}
}

func TestDecodeCitiesWithNamesMatchesDongzhuoMap(t *testing.T) {
	archive, err := Open(decodeTestArchivePath(t))
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}

	cities, names, err := archive.DecodeCitiesWithNames()
	if err != nil {
		t.Fatalf("DecodeCitiesWithNames error = %v", err)
	}
	if len(cities) != 38 {
		t.Fatalf("decoded %d cities, want 38", len(cities))
	}

	// Known董卓-era city positions in the roster.
	expected := map[int]string{0: "西凉", 1: "北平", 5: "平原", 10: "长安"}
	for idx, want := range expected {
		if names[idx] != want {
			t.Errorf("names[%d] = %q, want %q", idx, names[idx], want)
		}
	}

	// All cities should have sane economy values.
	for i, c := range cities {
		if c.PeopleDevotion > 100 || c.AvoidCalamity > 100 {
			t.Errorf("city[%d] %s devotion/calamity out of range: %d/%d", i, names[i], c.PeopleDevotion, c.AvoidCalamity)
		}
		if c.Money > 30000 || c.Food > 30000 {
			t.Errorf("city[%d] %s money/food out of range: %d/%d", i, names[i], c.Money, c.Food)
		}
	}

	for i := 0; i < 12; i++ {
		c := cities[i]
		t.Logf("  city[%2d] %-4s 归属%d 农%d/%d 商%d/%d 民忠%d 防灾%d 人口%d 金%d 粮%d",
			i, names[i], c.Belong, c.Farming, c.FarmingLimit, c.Commerce, c.CommerceLimit,
			c.PeopleDevotion, c.AvoidCalamity, c.Population, c.Money, c.Food)
	}
}

func TestDecodeAllCitiesPeriod1(t *testing.T) {
	archive, err := Open(decodeTestArchivePath(t))
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}

	raw, err := archive.Item(57, 1)
	if err != nil {
		t.Fatalf("Item(57, 1) error = %v", err)
	}

	// Resource 57 item length is 1180; figure out how many cities fit
	cities, err := DecodeCities(raw, len(raw)/CityStructSize)
	if err != nil {
		t.Fatalf("DecodeCities error = %v", err)
	}
	if len(cities) == 0 {
		t.Fatal("No cities decoded")
	}
	t.Logf("Decoded %d cities from %d raw bytes (ItemLength=%d, %d bytes/city)",
		len(cities), len(raw), len(raw), CityStructSize)

	if len(cities) != 38 {
		t.Fatalf("decoded %d cities, want 38", len(cities))
	}

	// Verify at least some cities have reasonable data
	hasMoney := 0
	hasFood := 0
	owned := 0
	for _, c := range cities {
		if c.Money > 0 {
			hasMoney++
		}
		if c.Food > 0 {
			hasFood++
		}
		if c.Belong > 0 {
			owned++
		}
	}
	t.Logf("Cities with money: %d, food: %d, owned (belong>0): %d", hasMoney, hasFood, owned)
	if hasMoney == 0 {
		t.Error("No cities have money")
	}
}

// ---- Person Decoding Tests ----

func TestDecodePersonsFromResource61(t *testing.T) {
	archive, err := Open(decodeTestArchivePath(t))
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}

	raw, err := archive.Item(61, 1)
	if err != nil {
		t.Fatalf("Item(61, 1) error = %v", err)
	}

	persons, err := DecodePersons(raw)
	if err != nil {
		t.Fatalf("DecodePersons error = %v", err)
	}
	if len(persons) == 0 {
		t.Fatal("No persons decoded")
	}
	t.Logf("Decoded %d persons from %d raw bytes (%d bytes each)",
		len(persons), len(raw), PersonStructSize)

	// Verify first person has reasonable values
	p := persons[0]
	t.Logf("Person[0]: Index=%d NameItem=%d Lv=%d Str=%d Int=%d Loy=%d Char=%d ArmsType=%d Age=%d",
		p.Index, p.NameItem, p.Level, p.Force, p.IQ, p.Devotion, p.Character, p.ArmsType, p.Age)

	// Sanity checks on first person — should be 董卓 with strong stats.
	if p.Force > 100 || p.IQ > 100 || p.Devotion > 100 {
		t.Errorf("Person[0] stats out of range: Str=%d Int=%d Loy=%d", p.Force, p.IQ, p.Devotion)
	}
	if p.ArmsType > 5 {
		t.Errorf("Person[0] ArmsType=%d out of range [0-5]", p.ArmsType)
	}
	if p.Level == 0 && p.Force == 0 && p.IQ == 0 {
		t.Error("Person[0] appears empty — possible decode issue")
	}

	// Validate the first 19 leaders (rulers) have sane stats.
	for i := 0; i < 19 && i < len(persons); i++ {
		ps := persons[i]
		if ps.Force == 0 || ps.Force > 100 {
			t.Errorf("Leader[%d] Force=%d out of [1,100]", i, ps.Force)
		}
		if ps.IQ == 0 || ps.IQ > 100 {
			t.Errorf("Leader[%d] IQ=%d out of [1,100]", i, ps.IQ)
		}
		if ps.ArmsType > 5 {
			t.Errorf("Leader[%d] ArmsType=%d out of [0,5]", i, ps.ArmsType)
		}
	}
}

func TestDecodePersonsWithNamesMatchesDongzhuoRoster(t *testing.T) {
	archive, err := Open(decodeTestArchivePath(t))
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}

	persons, names, err := archive.DecodePersonsWithNames()
	if err != nil {
		t.Fatalf("DecodePersonsWithNames error = %v", err)
	}
	if len(persons) != len(names) {
		t.Fatalf("persons=%d names=%d length mismatch", len(persons), len(names))
	}

	// The 董卓 scenario opens with these leaders in order.
	expected := map[int]string{
		0:  "董卓",
		1:  "曹操",
		2:  "袁绍",
		3:  "袁术",
		4:  "孙坚",
		13: "刘备",
	}
	for idx, want := range expected {
		if idx >= len(names) {
			t.Fatalf("person index %d out of range", idx)
		}
		if names[idx] != want {
			t.Errorf("names[%d] = %q, want %q", idx, names[idx], want)
		}
	}

	// Log first 10 for visibility
	for i := 0; i < 10 && i < len(persons); i++ {
		p := persons[i]
		t.Logf("  %2d %s 武%d 智%d 忠%d 性格%s 兵种%s 年龄%d",
			i, names[i], p.Force, p.IQ, p.Devotion,
			CharacterNames[p.Character], ArmsTypeNames[p.ArmsType], p.Age)
	}
}

// ---- City Name Decoding Tests ----

func TestDecodeCityNamesGBK(t *testing.T) {
	archive, err := Open(decodeTestArchivePath(t))
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}

	names, err := archive.DecodeAllCityNames()
	if err != nil {
		t.Fatalf("DecodeAllCityNames error = %v", err)
	}
	if len(names) == 0 {
		t.Fatal("No city names decoded")
	}
	t.Logf("Decoded %d city names", len(names))

	// First name should be recognizable Chinese
	if names[0] == "" {
		t.Log("First city name is empty")
	} else {
		t.Logf("City[0] name = %q (%X)", names[0], []byte(names[0]))
	}

	// Count non-empty
	nonEmpty := 0
	for _, name := range names {
		if name != "" {
			nonEmpty++
		}
	}
	t.Logf("Non-empty city names: %d/%d", nonEmpty, len(names))

	// Spot-check for known cities
	knownCities := map[string]bool{
		"洛阳": true, "长安": true, "许昌": true, "陈留": true,
		"邺城": true, "平原": true, "下邳": true, "建业": true,
		"吴郡": true, "江夏": true, "成都": true, "荆州": true,
	}
	found := 0
	for _, name := range names {
		if knownCities[name] {
			found++
			delete(knownCities, name)
		}
	}
	t.Logf("Found %d known city names", found)
	if len(knownCities) > 0 {
		t.Logf("Not found: %v", knownCities)
	}

	// Print first 10 names for visual verification
	for i := 0; i < min(10, len(names)); i++ {
		t.Logf("  City[%d] = %q", i, names[i])
	}
}

func TestDecodeCityNameDirectGBK(t *testing.T) {
	// 长安 in GBK is B3 A4 B0 B2; decrypted bytes are transcoded to UTF-8.
	raw := []byte{0xb3, 0xa4, 0xb0, 0xb2, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}
	name, err := DecodeGBKName(raw)
	if err != nil {
		t.Fatalf("DecodeGBKName error = %v", err)
	}
	if name != "长安" {
		t.Errorf("name = %q, want 长安", name)
	}
}

func TestDecodeCityNameTrimsNullPadding(t *testing.T) {
	// A 2-char name followed by 0x00 should trim cleanly. 北海: B1 B1 BA A3
	raw := []byte{0xb1, 0xb1, 0xba, 0xa3, 0x00, 0xcc, 0xcc, 0xcc, 0xcc, 0xcc}
	name, err := DecodeGBKName(raw)
	if err != nil {
		t.Fatalf("DecodeGBKName error = %v", err)
	}
	if name != "北海" {
		t.Errorf("name = %q, want 北海", name)
	}
}

// ---- Period-wide Tests ----

func TestDecodeAllPeriods(t *testing.T) {
	archive, err := Open(decodeTestArchivePath(t))
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}

	for period := 1; period <= 4; period++ {
		cityRaw, err := archive.Item(57, uint16(period))
		if err != nil {
			t.Fatalf("Item(57, %d) error = %v", period, err)
		}
		cities, err := DecodeCities(cityRaw, len(cityRaw)/CityStructSize)
		if err != nil {
			t.Fatalf("DecodeCities period %d error = %v", period, err)
		}

		personRaw, err := archive.Item(61, uint16(period))
		if err != nil {
			t.Fatalf("Item(61, %d) error = %v", period, err)
		}
		persons, err := DecodePersons(personRaw)
		if err != nil {
			t.Fatalf("DecodePersons period %d error = %v", period, err)
		}

		t.Logf("Period %d: %d cities (%d bytes), %d persons (%d bytes)",
			period, len(cities), len(cityRaw), len(persons), len(personRaw))
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

