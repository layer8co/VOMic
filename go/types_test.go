package vomic

import (
	"testing"
)

// Tests whether if the SessionInit struct is being Unmarshaled currectly from
// it's byte encoding or not.
func TestSessionInitUnmarshal(t *testing.T) {
	sInit := SessionInit{
		Flags:    1,
		Hostname: "Redmi Note 8",
	}

	bytes, err := sInit.MarshalBinary()
	if err != nil {
		t.Fatal(err)
	}

	sInit2 := SessionInit{}

	err = sInit2.UnmarshalBinary(bytes)

	if err != nil {
		t.Fatal(err)
	}

	if sInit2 != sInit {
		t.Fatalf("Two structs are not the same. sent: %v \n Now got: %v", sInit, sInit2)
	}

	t.Logf("%#v", sInit2)
}

func TestSessionInitFlags(t *testing.T) {
	sInit := SessionInit{
		Flags:    1,
		Hostname: "Testing",
	}

	yes := sInit.Flags.HasFlag(FlagStereo)

	if !yes {
		t.Fatal("Expected to have the binary flag sterio")
	}

}
