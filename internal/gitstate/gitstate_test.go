package gitstate

import (
	"reflect"
	"testing"
)

func TestParseStatusPorcelainV2ZHandlesSpacesAndRenames(t *testing.T) {
	input := []byte(
		"1 M. N... 100644 100644 100644 abcdef abcdef path with spaces.txt\x00" +
			"? new file.txt\x00" +
			"2 R. N... 100644 100644 100644 abcdef abcdef R100 renamed file.txt\x00old file.txt\x00" +
			"! ignored.txt\x00",
	)

	dirty, untracked := parseStatusPorcelainV2Z(input)

	wantDirty := []string{"path with spaces.txt", "old file.txt -> renamed file.txt"}
	if !reflect.DeepEqual(dirty, wantDirty) {
		t.Fatalf("dirty paths mismatch:\nwant %#v\ngot  %#v", wantDirty, dirty)
	}
	wantUntracked := []string{"new file.txt"}
	if !reflect.DeepEqual(untracked, wantUntracked) {
		t.Fatalf("untracked paths mismatch:\nwant %#v\ngot  %#v", wantUntracked, untracked)
	}
}
