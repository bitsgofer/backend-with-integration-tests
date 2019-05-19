package pgdata

import (
	"context"
	"database/sql"
	"time"

	"github.com/bitsgofer/backend-with-integration-tests/internal/config"
	_ "github.com/lib/pq"
	"github.com/pkg/errors"
)

type Client struct {
	db             *sql.DB
	defaultTimeout time.Duration
}

type TestCase struct {
	ID            int
	Input, Output string
}

type Problem struct {
	ID              int
	Description     string
	ExampleTestCase TestCase
}

var ErrNotFound = errors.New("not found")

func NewClient(c config.PGConfig) (*Client, error) {
	db, err := newDB(c)
	if err != nil {
		return nil, err
	}

	if err := setupTables(db); err != nil {
		return nil, err
	}

	return &Client{
		db:             db,
		defaultTimeout: time.Second * 10,
	}, nil
}

func (c *Client) Close() {
	c.db.Close()
}

func newDB(c config.PGConfig) (*sql.DB, error) {
	db, err := sql.Open("postgres", c.ConnStr)
	if err != nil {
		return nil, errors.Wrapf(err, "cannot connect to DB")
	}

	db.SetMaxIdleConns(c.MaxIdle)
	db.SetMaxOpenConns(c.MaxOpen)

	return db, nil
}

func setupTables(db *sql.DB) error {
	const createTableTestCases = `
CREATE TABLE test_cases(
	id SERIAL PRIMARY KEY,
	input text NOT NULL,
	output text NOT NULL
);`
	const createTableProblems = `
CREATE TABLE problems(
	id SERIAL PRIMARY KEY,
	description text NOT NULL,
	example_test_case_id integer REFERENCES test_cases(id)
);`
	const createTableTestCaseUsage = `
CREATE TABLE test_case_assignments(
	problem_id integer REFERENCES problems(id),
	test_case_id integer REFERENCES test_cases(id)
);`

	if _, err := db.Exec(createTableTestCases); err != nil {
		return errors.Wrapf(err, "cannot create table 'test_cases'")
	}
	if _, err := db.Exec(createTableProblems); err != nil {
		return errors.Wrapf(err, "cannot create table 'problems'")
	}
	if _, err := db.Exec(createTableTestCaseUsage); err != nil {
		return errors.Wrapf(err, "cannot create table 'test_case_usage'")
	}

	return nil
}

func rollbackOrPanic(tx *sql.Tx) {
	if err := tx.Rollback(); err != nil {
		panic("cannot rollback transaction")
	}
}

func (c *Client) NewProblem(description, sampleInput, sampleOutput string) (Problem, error) {
	ctx, cancel := context.WithTimeout(context.Background(), c.defaultTimeout)
	defer cancel()

	tx, err := c.db.BeginTx(ctx, nil)
	if err != nil {
		return Problem{}, errors.Wrapf(err, "cannot begin transaction")
	}

	problem, err := insertProblem(ctx, tx, description, sampleInput, sampleOutput)
	if err != nil {
		return Problem{}, err
	}

	if err := tx.Commit(); err != nil {
		defer rollbackOrPanic(tx)
		return Problem{}, errors.Wrapf(err, "cannot commit transaction")
	}

	return problem, nil
}

func insertProblem(ctx context.Context, tx *sql.Tx, description, sampleInput, sampleOutput string) (Problem, error) {
	const query = `INSERT INTO problems (description, example_test_case_id) VALUES ($1, $2) RETURNING id;`

	tc, err := insertTestCase(ctx, tx, sampleInput, sampleOutput)
	if err != nil {
		return Problem{}, errors.Wrapf(err, "cannot insert example test case")
	}

	description = sanitizeStr(description)
	var pl Problem
	if err := tx.QueryRowContext(ctx, query, description, tc.ID).Scan(&pl.ID); err != nil {
		return Problem{}, errors.Wrapf(err, "query failed")
	}

	if err := assignTestCaseToProblem(ctx, tx, pl.ID, tc.ID); err != nil {
		return Problem{}, errors.Wrapf(err, "cannot assign test case to problem")
	}

	pl.Description = description
	pl.ExampleTestCase = tc
	return pl, nil
}

func (c *Client) FindProblemByID(id int) (Problem, error) {
	ctx, cancel := context.WithTimeout(context.Background(), c.defaultTimeout)
	defer cancel()

	problem, err := findProblemByID(ctx, c.db, id)
	if err != nil {
		return Problem{}, err
	}

	return problem, nil
}

func findProblemByID(ctx context.Context, db *sql.DB, id int) (Problem, error) {
	const query = `
SELECT pl.id, pl.description, tc.id, tc.input, tc.output
FROM problems AS pl INNER JOIN test_cases AS tc
	ON pl.example_test_case_id = tc.id
WHERE pl.id = $1 LIMIT 1;`

	var tc TestCase
	var pl Problem
	if err := db.QueryRowContext(ctx, query, id).Scan(&pl.ID, &pl.Description, &tc.ID, &tc.Input, &tc.Output); err == sql.ErrNoRows {
		return Problem{}, ErrNotFound
	} else if err != nil {
		return Problem{}, errors.Wrapf(err, "query failed")
	}

	pl.ExampleTestCase = tc
	return pl, nil
}

func (c *Client) NewTestCase(problemID int, input, output string) (TestCase, error) {
	ctx, cancel := context.WithTimeout(context.Background(), c.defaultTimeout)
	defer cancel()

	tx, err := c.db.BeginTx(ctx, nil)
	if err != nil {
		return TestCase{}, errors.Wrapf(err, "cannot begin transaction")
	}

	tc, err := insertTestCase(ctx, tx, input, output)
	if err != nil {
		return TestCase{}, errors.Wrapf(err, "cannot insert test case")
	}
	if err := assignTestCaseToProblem(ctx, tx, problemID, tc.ID); err != nil {
		return TestCase{}, errors.Wrapf(err, "cannot assign test case to problem")
	}

	if err := tx.Commit(); err != nil {
		defer rollbackOrPanic(tx)
		return TestCase{}, errors.Wrapf(err, "cannot commit transaction")
	}

	return tc, nil

}

func insertTestCase(ctx context.Context, tx *sql.Tx, input, output string) (TestCase, error) {
	const query = `INSERT INTO test_cases (input, output) VALUES ($1, $2) RETURNING id;`

	input = sanitizeStr(input)
	output = sanitizeStr(output)
	var tc TestCase
	if err := tx.QueryRowContext(ctx, query, input, output).Scan(&tc.ID); err != nil {
		return TestCase{}, errors.Wrapf(err, "query failed")
	}

	tc.Input = input
	tc.Output = output
	return tc, nil
}

func assignTestCaseToProblem(ctx context.Context, tx *sql.Tx, problemID int, testCaseIDs ...int) error {
	const query = `INSERT INTO test_case_assignments (problem_id, test_case_id) VALUES ($1, $2);`

	for _, tcID := range testCaseIDs {
		if _, err := tx.ExecContext(ctx, query, problemID, tcID); err != nil {
			defer rollbackOrPanic(tx)
			return errors.Wrapf(err, "query failed")
		}
	}

	return nil
}

func (c *Client) FindTestCaseByID(id int) (TestCase, error) {
	ctx, cancel := context.WithTimeout(context.Background(), c.defaultTimeout)
	defer cancel()

	tc, err := findTestCaseByID(ctx, c.db, id)
	if err != nil {
		return TestCase{}, err
	}

	return tc, nil
}

func findTestCaseByID(ctx context.Context, db *sql.DB, id int) (TestCase, error) {
	const query = `SELECT id, input, output FROM test_cases WHERE id = $1 LIMIT 1;`

	var tc TestCase
	if err := db.QueryRowContext(ctx, query, id).Scan(&tc.ID, &tc.Input, &tc.Output); err == sql.ErrNoRows {
		return TestCase{}, ErrNotFound
	} else if err != nil {
		return TestCase{}, errors.Wrapf(err, "query failed")
	}

	return tc, nil
}

func (c *Client) FindTestCasesByProblemID(id int) ([]TestCase, error) {
	ctx, cancel := context.WithTimeout(context.Background(), c.defaultTimeout)
	defer cancel()

	testCases, err := findTestCasesByProblemID(ctx, c.db, id)
	if err != nil {
		return nil, err
	}

	return testCases, nil
}

func findTestCasesByProblemID(ctx context.Context, db *sql.DB, problemID int) ([]TestCase, error) {
	const query = `
SELECT tc.id, tc.input, tc.output
FROM problems AS pl
	INNER JOIN test_case_assignments AS usg
		ON pl.id = usg.problem_id
	INNER JOIN test_cases AS tc
		ON tc.id = usg.test_case_id
WHERE pl.id = $1
ORDER BY tc.id;`

	var tcs []TestCase
	rows, err := db.QueryContext(ctx, query, problemID)
	if err == sql.ErrNoRows {
		return nil, nil
	} else if err != nil {
		return nil, errors.Wrapf(err, "query failed")
	}

	for rows.Next() {
		var tc TestCase
		if err := rows.Scan(&tc.ID, &tc.Input, &tc.Output); err != nil {
			return nil, errors.Wrapf(err, "cannot scan")
		}
		tcs = append(tcs, tc)
	}

	return tcs, nil
}

func sanitizeStr(s string) string {
	// TODO(exklamationmark): actually sanitize the values for use in SQL
	return s
}
