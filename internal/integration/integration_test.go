package integration

import "testing"

func TestNew(t *testing.T) {
	New(t, PrintDebugLog)

	t.Run("subTest", func(t *testing.T) {
		New(t)
	})

	t.Run("subTest/withSlash", func(t *testing.T) {
		New(t)
	})

	t.Run("thisIsASuperLongTestNameThatWillNeedToBeTruncated", func(t *testing.T) {
		New(t)
	})
}
