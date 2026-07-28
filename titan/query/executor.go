package query

import (
	"fmt"

	"github.com/olympus/titan/sql"
)

type RowBatch struct {
	Rows [][]interface{} // Columnar format is better, but row-based is easier for mock
}

// Vectorized execution — processes 1,024 rows per batch.
type Executor struct {
	BatchSize int
}

func NewExecutor() *Executor {
	return &Executor{
		BatchSize: 1024,
	}
}

func (e *Executor) Execute(stmt sql.Statement, strategy Strategy) (*RowBatch, error) {
	switch s := stmt.(type) {
	case *sql.SelectStatement:
		return e.executeSelect(s, strategy)
	case *sql.CreateTableStatement:
		return e.executeCreate(s)
	case *sql.InsertStatement:
		return e.executeInsert(s)
	default:
		return nil, fmt.Errorf("unsupported statement type")
	}
}

func (e *Executor) executeSelect(stmt *sql.SelectStatement, strategy Strategy) (*RowBatch, error) {
	// Vectorized processing simulation
	fmt.Printf("Executing SELECT using strategy %v\n", strategy)
	
	// Simulated batched pipelined execution
	batch := &RowBatch{
		Rows: make([][]interface{}, 0, e.BatchSize),
	}

	// Fake data
	batch.Rows = append(batch.Rows, []interface{}{"Rahul", 1500.50})
	batch.Rows = append(batch.Rows, []interface{}{"Priya", 2200.00})

	return batch, nil
}

func (e *Executor) executeCreate(stmt *sql.CreateTableStatement) (*RowBatch, error) {
	fmt.Printf("Executing CREATE TABLE %s\n", stmt.TableName)
	return &RowBatch{}, nil
}

func (e *Executor) executeInsert(stmt *sql.InsertStatement) (*RowBatch, error) {
	fmt.Printf("Executing INSERT INTO %s\n", stmt.TableName)
	return &RowBatch{}, nil
}
