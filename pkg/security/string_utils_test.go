package security

import "testing"

func TestGenerateUniqueID(t *testing.T) {
	numIDs := 1
	ids := make(map[string]bool)

	for i := 0; i < numIDs; i++ {
		id := GenerateRandomString(10)
		if ids[id] {
			t.Errorf("Duplicate ID generated: %s", id)
		}
		ids[id] = true
	}
}
