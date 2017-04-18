package mock_godless

import (
	"github.com/golang/mock/gomock"
	lib "github.com/johnny-morrice/godless"

	"testing"
)

func TestVisitJoin(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	joinQuery := lib.Query{}
	joinQuery.TableKey = "Table Key"
	joinQuery.OpCode = lib.JOIN
	joinQuery.Join.Rows = []lib.QueryRowJoin{
		lib.QueryRowJoin{
			RowKey: "Hello Row",
			Entries: map[string]string{
				"Entry A": "Value A",
			},
		},
		lib.QueryRowJoin{
			RowKey: "Goodbye Row",
			Entries: map[string]string{
				"Entry B": "Value B",
			},
		},
	}

	visitor := NewMockQueryVisitor(ctrl)
	visitor.EXPECT().VisitTableKey("Table Key")
	visitor.EXPECT().VisitAST(nil)
	visitor.EXPECT().VisitParser(nil)
	visitor.EXPECT().VisitOpCode(lib.JOIN)
	visitor.EXPECT().VisitJoin(&joinQuery.Join)
	visitor.EXPECT().VisitRowJoin(0, &joinQuery.Join.Rows[0])
	visitor.EXPECT().VisitRowJoin(1, &joinQuery.Join.Rows[1])

	joinQuery.Visit(visitor)
}

func TestVisitSelect(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	innerWhereA := lib.QueryWhere{}
	innerWhereA.OpCode = lib.PREDICATE
	innerWhereA.Predicate = lib.QueryPredicate{
		OpCode: lib.STR_NEQ,
		Keys: []string{"Index this", "Index that"},
	}

	innerWhereB := lib.QueryWhere{}
	innerWhereB.OpCode = lib.PREDICATE
	innerWhereB.Predicate = lib.QueryPredicate{
		OpCode: lib.STR_EQ,
		Literals: []string{"Match this"},
		IncludeRowKey: true,
	}

	outerWhere := lib.QueryWhere{}
	outerWhere.OpCode = lib.AND
	outerWhere.Clauses = []lib.QueryWhere{
		innerWhereA,
		innerWhereB,
	}

	selectQuery := lib.Query{}
	selectQuery.TableKey = "Table Key"
	selectQuery.OpCode = lib.SELECT
	selectQuery.Select.Limit = 5
	selectQuery.Select.Where = outerWhere

	visitor := NewMockQueryVisitor(ctrl)
	visitor.EXPECT().VisitTableKey("Table Key")
	visitor.EXPECT().VisitAST(nil)
	visitor.EXPECT().VisitParser(nil)
	visitor.EXPECT().VisitOpCode(lib.SELECT)
	visitor.EXPECT().VisitSelect(&selectQuery.Select)
	visitor.EXPECT().VisitWhere(0, &selectQuery.Select.Where)
	visitor.EXPECT().VisitPredicate(&selectQuery.Select.Where.Predicate)
	visitor.EXPECT().VisitWhere(0, &selectQuery.Select.Where.Clauses[0])
	visitor.EXPECT().VisitPredicate(&selectQuery.Select.Where.Clauses[0].Predicate)
	visitor.EXPECT().LeaveWhere(&selectQuery.Select.Where.Clauses[0])
	visitor.EXPECT().VisitWhere(1, &selectQuery.Select.Where.Clauses[1])
	visitor.EXPECT().VisitPredicate(&selectQuery.Select.Where.Clauses[1].Predicate)
	visitor.EXPECT().LeaveWhere(&selectQuery.Select.Where.Clauses[1])
	visitor.EXPECT().LeaveWhere(&selectQuery.Select.Where)

	selectQuery.Visit(visitor)
}
