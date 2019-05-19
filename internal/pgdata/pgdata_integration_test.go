// +build integration

package pgdata

import (
	"sort"
	"testing"

	"github.com/bitsgofer/backend-with-integration-tests/internal/integration"
	"github.com/google/go-cmp/cmp"
)

func TestNew(t *testing.T) {
	setup := integration.New(t)

	c, err := NewClient(setup.ServerConfig.Postgres)
	if err != nil {
		t.Fatalf("New(): want no error, got %v", err)
	}
	defer c.Close()
}

func TestProblem(t *testing.T) {
	setup := integration.New(t)
	c := mustGetClient(t, setup)
	defer c.Close()

	pl, err := c.NewProblem("hello", "", "hello, world!")
	if err != nil {
		t.Fatalf("NewProblem(): want no error, got %v", err)
	}

	inDB := Problem{
		ID:          1,
		Description: "hello",
		ExampleTestCase: TestCase{
			ID:     1,
			Input:  "",
			Output: "hello, world!",
		},
	}
	if want, got := inDB, pl; !cmp.Equal(want, got) {
		t.Fatalf("NewProblem():\n want= %v;\n  got= %v;\n diff= %v", want, got, cmp.Diff(want, got))
	}

	pl, err = c.FindProblemByID(1)
	if err != nil {
		t.Fatalf("FindProblemByID(): want no error, got %v", err)
	}
	if want, got := inDB, pl; !cmp.Equal(want, got) {
		t.Fatalf("FindProblemByID():\n want= %v;\n  got= %v;\n diff= %v", want, got, cmp.Diff(want, got))
	}

}

func TestTestCase(t *testing.T) {
	setup := integration.New(t)
	c := mustGetClient(t, setup)
	defer c.Close()

	// insert problem with ID= 1 and also 1 test case with ID= 1
	pl, err := c.NewProblem("hello", "", "hello, world!")
	if err != nil {
		t.Fatalf("NewProblem(): want no error, got %v", err)
	}

	tc, err := c.NewTestCase(1, "hi", "there")
	if err != nil {
		t.Fatalf("NewTestCase(): want no error, got %v", err)
	}

	inDB := TestCase{
		ID:     2,
		Input:  "hi",
		Output: "there",
	}
	if want, got := inDB, tc; !cmp.Equal(want, got) {
		t.Fatalf("NewTestCase():\n want= %v;\n  got= %v;\n diff= %v", want, got, cmp.Diff(want, got))
	}

	tc, err = c.FindTestCaseByID(2)
	if err != nil {
		t.Fatalf("FindTestCaseByID(): want no error, got %v", err)
	}
	if want, got := inDB, tc; !cmp.Equal(want, got) {
		t.Fatalf("FindTestCaseByID():\n want= %v;\n  got= %v;\n diff= %v", want, got, cmp.Diff(want, got))
	}

	assingedTestCases := []TestCase{
		TestCase{
			ID:     1,
			Input:  "",
			Output: "hello, world!",
		},
		TestCase{
			ID:     2,
			Input:  "hi",
			Output: "there",
		},
	}
	tcs, err := c.FindTestCasesByProblemID(pl.ID)
	if err != nil {
		t.Fatalf("FindTestCasesByProblemID(): want no error, got %v", err)
	}
	sort.Slice(tcs, func(i, j int) bool {
		return tcs[i].ID < tcs[j].ID
	})
	if want, got := assingedTestCases, tcs; !cmp.Equal(want, got) {
		t.Fatalf("FindTestCasesByProblemID():\n want= %v;\n  got= %v;\n diff= %v", want, got, cmp.Diff(want, got))
	}

}

func mustGetClient(t *testing.T, s *integration.Setup) *Client {
	c, err := NewClient(s.ServerConfig.Postgres)
	if err != nil {
		t.Fatalf("NewClient(): want no error, got %v", err)
	}

	return c
}
