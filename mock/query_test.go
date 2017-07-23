package mock_godless

import (
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/johnny-morrice/godless/crdt"
	"github.com/johnny-morrice/godless/query"
)

func TestVisitJoin(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	joinQuery := query.Query{}
	joinQuery.TableKey = MAIN_TABLE_KEY
	joinQuery.OpCode = query.JOIN
	joinQuery.Join.Rows = []query.QueryRowJoin{
		query.QueryRowJoin{
			RowKey: "Hello Row",
			Entries: map[crdt.EntryName]crdt.PointText{
				"Entry A": "Point A",
			},
		},
		query.QueryRowJoin{
			RowKey: "Goodbye Row",
			Entries: map[crdt.EntryName]crdt.PointText{
				"Entry B": "Point B",
			},
		},
	}

	visitor := NewMockQueryVisitor(ctrl)
	c1 := visitor.EXPECT().VisitAST(nil)
	c2 := visitor.EXPECT().VisitParser(nil)
	c3 := visitor.EXPECT().VisitOpCode(query.JOIN)
	c4 := visitor.EXPECT().VisitTableKey(crdt.TableName(MAIN_TABLE_KEY))
	c5 := visitor.EXPECT().VisitJoin(&joinQuery.Join)
	c6 := visitor.EXPECT().VisitRowJoin(0, &joinQuery.Join.Rows[0])
	c7 := visitor.EXPECT().VisitRowJoin(1, &joinQuery.Join.Rows[1])
	c8 := visitor.EXPECT().LeaveJoin(&joinQuery.Join)

	gomock.InOrder(c1, c2, c3, c4, c5, c6, c7, c8)

	joinQuery.Visit(visitor)
}

func TestVisitSelect(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	innerWhereA := query.QueryWhere{}
	innerWhereA.OpCode = query.PREDICATE
	innerWhereA.Predicate = query.QueryPredicate{
		OpCode: query.STR_EQ,
		Keys:   []crdt.EntryName{"Index this", "Index that"},
	}

	innerWhereB := query.QueryWhere{}
	innerWhereB.OpCode = query.PREDICATE
	innerWhereB.Predicate = query.QueryPredicate{
		OpCode:        query.STR_EQ,
		Literals:      []string{"Match this"},
		IncludeRowKey: true,
	}

	outerWhere := query.QueryWhere{}
	outerWhere.OpCode = query.AND
	outerWhere.Clauses = []query.QueryWhere{
		innerWhereA,
		innerWhereB,
	}

	selectQuery := query.Query{}
	selectQuery.TableKey = MAIN_TABLE_KEY
	selectQuery.OpCode = query.SELECT
	selectQuery.Select.Limit = 5
	selectQuery.Select.Where = outerWhere

	visitor := NewMockQueryVisitor(ctrl)

	querySelect := &selectQuery.Select
	c1 := visitor.EXPECT().VisitAST(nil)
	c2 := visitor.EXPECT().VisitParser(nil)
	c3 := visitor.EXPECT().VisitOpCode(query.SELECT)
	c4 := visitor.EXPECT().VisitTableKey(MAIN_TABLE_KEY)
	c5 := visitor.EXPECT().VisitSelect(querySelect)
	c6 := visitor.EXPECT().VisitWhere(0, &selectQuery.Select.Where)
	c7 := visitor.EXPECT().VisitPredicate(&selectQuery.Select.Where.Predicate)
	c8 := visitor.EXPECT().VisitWhere(0, &selectQuery.Select.Where.Clauses[0])
	c9 := visitor.EXPECT().VisitPredicate(&selectQuery.Select.Where.Clauses[0].Predicate)
	c10 := visitor.EXPECT().LeaveWhere(&selectQuery.Select.Where.Clauses[0])
	c11 := visitor.EXPECT().VisitWhere(1, &selectQuery.Select.Where.Clauses[1])
	c12 := visitor.EXPECT().VisitPredicate(&selectQuery.Select.Where.Clauses[1].Predicate)
	c13 := visitor.EXPECT().LeaveWhere(&selectQuery.Select.Where.Clauses[1])
	c14 := visitor.EXPECT().LeaveWhere(&selectQuery.Select.Where)
	c15 := visitor.EXPECT().LeaveSelect(querySelect)

	gomock.InOrder(c1, c2, c3,
		c4, c5, c6,
		c7, c8, c9,
		c10, c11, c12,
		c13, c14, c15)

	selectQuery.Visit(visitor)
}
