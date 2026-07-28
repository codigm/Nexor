package query

import (
	"github.com/olympus/titan/sql"
)

type Strategy int

const (
	StrategySQLFirst Strategy = iota
	StrategyVectorFirst
	StrategyParallel
)

type Optimizer struct {
	// Metadata stats (mocked)
	sqlSelectivity   float64
	vectorSelectivity float64
}

func NewOptimizer() *Optimizer {
	return &Optimizer{
		sqlSelectivity:    0.05, // 5% selectivity (highly selective)
		vectorSelectivity: 0.15, // 15% selectivity
	}
}

// Optimize decides the execution strategy for a hybrid query.
// It returns an execution plan based on the cost.
func (o *Optimizer) Optimize(stmt *sql.SelectStatement) Strategy {
	hasSQLFilter := stmt.Where != nil
	hasVectorFilter := false

	// Quick check for VECTOR_SIMILARITY in WHERE clause
	if hasSQLFilter {
		hasVectorFilter = containsVectorSimilarity(stmt.Where)
	}

	if hasSQLFilter && hasVectorFilter {
		// Both filters exist. Compare selectivity.
		// If SQL is very selective (< 10%), filter first.
		if o.sqlSelectivity < 0.10 {
			return StrategySQLFirst
		} else if o.vectorSelectivity < 0.10 {
			return StrategyVectorFirst
		}
		return StrategyParallel
	}

	return StrategySQLFirst // default
}

func containsVectorSimilarity(expr sql.Expr) bool {
	switch e := expr.(type) {
	case *sql.BinaryExpr:
		if _, ok := e.Left.(*sql.VectorSimilarityExpr); ok {
			return true
		}
		if _, ok := e.Right.(*sql.VectorSimilarityExpr); ok {
			return true
		}
		return containsVectorSimilarity(e.Left) || containsVectorSimilarity(e.Right)
	case *sql.VectorSimilarityExpr:
		return true
	}
	return false
}
