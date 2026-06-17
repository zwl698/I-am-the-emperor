package legacyres

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func legacyArchivePath(t *testing.T) string {
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

func TestOpenListsKnownLegacyResources(t *testing.T) {
	archive, err := Open(legacyArchivePath(t))
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}

	resources := archive.List(120)
	if got, want := len(resources), 87; got != want {
		t.Fatalf("List(120) count = %d, want %d", got, want)
	}

	cityNames, err := archive.Resource(58)
	if err != nil {
		t.Fatalf("Resource(58) error = %v", err)
	}
	if cityNames.Address != 192843 || cityNames.Length != 392 || cityNames.ID != 58 || cityNames.ItemCount != 43 || cityNames.ItemLength != 10 || cityNames.Key != 192 {
		t.Fatalf("Resource(58) = %+v", cityNames)
	}

	generals, err := archive.Resource(61)
	if err != nil {
		t.Fatalf("Resource(61) error = %v", err)
	}
	if generals.ItemCount != 4 || generals.ItemLength != 3000 || generals.Key != 0 {
		t.Fatalf("Resource(61) = %+v", generals)
	}
}

func TestItemReadsFixedLengthAndDecrypts(t *testing.T) {
	archive, err := Open(legacyArchivePath(t))
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}

	item, err := archive.Item(58, 1)
	if err != nil {
		t.Fatalf("Item(58, 1) error = %v", err)
	}
	if len(item) != 10 {
		t.Fatalf("Item(58, 1) len = %d, want 10", len(item))
	}
	if !bytes.HasPrefix(item, []byte{0xce, 0xf7, 0xc1, 0xb9}) {
		t.Fatalf("Item(58, 1) decrypted prefix = % x, want GBK bytes for 洛阳", item[:4])
	}

	generalScenario, err := archive.Item(61, 1)
	if err != nil {
		t.Fatalf("Item(61, 1) error = %v", err)
	}
	if len(generalScenario) != 3000 {
		t.Fatalf("Item(61, 1) len = %d, want 3000", len(generalScenario))
	}
}

func TestItemReadsVariableLengthResource(t *testing.T) {
	archive, err := Open(legacyArchivePath(t))
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}

	item, err := archive.Item(64, 1)
	if err != nil {
		t.Fatalf("Item(64, 1) error = %v", err)
	}
	if len(item) == 0 {
		t.Fatal("Item(64, 1) returned empty item")
	}

	header, err := archive.Resource(64)
	if err != nil {
		t.Fatalf("Resource(64) error = %v", err)
	}
	if header.ItemLength != 0 || header.ItemCount < 2 {
		t.Fatalf("Resource(64) should be variable length, got %+v", header)
	}
}

func TestItemRejectsInvalidIndex(t *testing.T) {
	archive, err := Open(legacyArchivePath(t))
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}

	_, err = archive.Item(58, 44)
	if !errors.Is(err, ErrItemOutOfRange) {
		t.Fatalf("Item(58, 44) error = %v, want %v", err, ErrItemOutOfRange)
	}
}
