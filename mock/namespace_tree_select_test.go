package mock_godless

import (
	"fmt"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/johnny-morrice/godless/api"
	"github.com/johnny-morrice/godless/crdt"
	"github.com/johnny-morrice/godless/crypto"
	"github.com/johnny-morrice/godless/internal/eval"
	"github.com/johnny-morrice/godless/internal/testutil"
	"github.com/johnny-morrice/godless/query"
	"github.com/pkg/errors"
)

func TestRunQuerySelectSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mock := NewMockNamespaceTree(ctrl)

	hash, cryptoErr := __SELECT_PUBLIC_KEY.Hash()

	testutil.AssertNil(t, cryptoErr)

	keyStore := &crypto.KeyStore{}
	cryptoErr = keyStore.PutPrivateKey(__SELECT_PRIVATE_KEY)

	testutil.AssertNil(t, cryptoErr)

	whereA := query.QueryWhere{
		OpCode: query.PREDICATE,
		Predicate: query.QueryPredicate{
			OpCode:   query.STR_EQ,
			Literals: []string{"Hi"},
			Keys:     []crdt.EntryName{"Entry A"},
		},
	}

	whereB := query.QueryWhere{
		OpCode: query.PREDICATE,
		Predicate: query.QueryPredicate{
			OpCode:   query.STR_EQ,
			Literals: []string{"Hi"},
			Keys:     []crdt.EntryName{"Entry B"},
		},
	}

	whereC := query.QueryWhere{
		OpCode: query.PREDICATE,
		Predicate: query.QueryPredicate{
			OpCode:   query.STR_NEQ,
			Literals: []string{"Hello World"},
			Keys:     []crdt.EntryName{"Entry B"},
		},
	}

	whereD := query.QueryWhere{
		OpCode: query.PREDICATE,
		Predicate: query.QueryPredicate{
			OpCode:   query.STR_EQ,
			Literals: []string{"Apple"},
			Keys:     []crdt.EntryName{"Entry C"},
		},
	}

	whereE := query.QueryWhere{
		OpCode: query.PREDICATE,
		Predicate: query.QueryPredicate{
			OpCode:   query.STR_EQ,
			Literals: []string{"Orange"},
			Keys:     []crdt.EntryName{"Entry D"},
		},
	}

	whereF := query.QueryWhere{
		OpCode: query.PREDICATE,
		Predicate: query.QueryPredicate{
			OpCode:   query.STR_EQ,
			Literals: []string{"Train"},
			Keys:     []crdt.EntryName{"Entry E"},
		},
	}

	whereG := query.QueryWhere{
		OpCode: query.PREDICATE,
		Predicate: query.QueryPredicate{
			OpCode:   query.STR_EQ,
			Literals: []string{"Bus"},
			Keys:     []crdt.EntryName{"Entry E"},
		},
	}

	whereH := query.QueryWhere{
		OpCode: query.PREDICATE,
		Predicate: query.QueryPredicate{
			OpCode:   query.STR_EQ,
			Literals: []string{"Boat"},
			Keys:     []crdt.EntryName{"Entry E"},
		},
	}

	whereI := query.QueryWhere{
		OpCode: query.PREDICATE,
		Predicate: query.QueryPredicate{
			OpCode:        query.STR_EQ,
			IncludeRowKey: true,
			Literals:      []string{"Row F0"},
		},
	}

	queries := []*query.Query{
		// One result
		&query.Query{
			OpCode:   query.SELECT,
			TableKey: MAIN_TABLE_KEY,
			Select: query.QuerySelect{
				Limit: 2,
				Where: whereA,
			},
		},
		// Multiple results
		&query.Query{
			OpCode:   query.SELECT,
			TableKey: MAIN_TABLE_KEY,
			Select: query.QuerySelect{
				Limit: 5,
				Where: whereB,
			},
		},
		// STR_NEQ
		&query.Query{
			OpCode:   query.SELECT,
			TableKey: MAIN_TABLE_KEY,
			Select: query.QuerySelect{
				Limit: 2,
				Where: whereC,
			},
		},
		// AND
		&query.Query{
			OpCode:   query.SELECT,
			TableKey: MAIN_TABLE_KEY,
			Select: query.QuerySelect{
				Limit: 2,
				Where: query.QueryWhere{
					OpCode:  query.AND,
					Clauses: []query.QueryWhere{whereD, whereE},
				},
			},
		},
		// OR
		&query.Query{
			OpCode:   query.SELECT,
			TableKey: MAIN_TABLE_KEY,
			Select: query.QuerySelect{
				Limit: 2,
				Where: query.QueryWhere{
					OpCode:  query.OR,
					Clauses: []query.QueryWhere{whereF, whereG},
				},
			},
		},
		// No results
		&query.Query{
			OpCode:   query.SELECT,
			TableKey: MAIN_TABLE_KEY,
			Select: query.QuerySelect{
				Limit: 2,
				Where: whereH,
			},
		},
		// Row key
		&query.Query{
			OpCode:   query.SELECT,
			TableKey: MAIN_TABLE_KEY,
			Select: query.QuerySelect{
				Limit: 2,
				Where: whereI,
			},
		},
		// No where or limits.
		&query.Query{
			OpCode:   query.SELECT,
			TableKey: ALT_TABLE_KEY,
		},
		// Signed,
		&query.Query{
			OpCode:     query.SELECT,
			TableKey:   ALT_TABLE_KEY,
			PublicKeys: []crypto.PublicKeyHash{hash},
		},
	}

	responseA := api.RESPONSE_QUERY
	responseA.QueryResponse.Entries = streamA()

	responseB := api.RESPONSE_QUERY
	responseB.QueryResponse.Entries = joinNamespaceStream(append(streamB(), streamC()...))

	responseC := api.RESPONSE_QUERY
	responseC.QueryResponse.Entries = streamC()

	responseD := api.RESPONSE_QUERY
	responseD.QueryResponse.Entries = streamD()

	responseE := api.RESPONSE_QUERY
	responseE.QueryResponse.Entries = streamE()

	responseF := api.RESPONSE_QUERY

	responseG := api.RESPONSE_QUERY
	responseG.QueryResponse.Entries = streamF()

	responseH := api.RESPONSE_QUERY
	responseH.QueryResponse.Entries = joinNamespaceStream(append(streamG(), streamH()...))

	responseI := api.RESPONSE_QUERY
	responseI.QueryResponse.Entries = streamH()

	expect := []api.APIResponse{
		responseA,
		responseB,
		responseC,
		responseD,
		responseE,
		responseF,
		responseG,
		responseH,
		responseI,
	}

	if len(queries) != len(expect) {
		panic("mismatched input and expect")
	}

	mock.EXPECT().LoadTraverse(gomock.Any()).Return(nil).Do(feedNamespace).Times(len(queries))

	for i, q := range queries {
		selector := eval.MakeNamespaceTreeSelect(mock, keyStore)
		q.Visit(selector)
		actual := selector.RunQuery()
		expected := expect[i]
		if !expected.Equals(actual) {
			if actual.QueryResponse.Entries == nil {
				t.Error("actual.QueryResponse.Entries was nil")
			}

			if actual.Err != nil {
				t.Error("actual.Err was", actual.Err)
			}

			t.Error("Case", i, "Expected", expected, "but receieved", actual)
		}
	}
}

func TestRunQuerySelectFailure(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mock := NewMockNamespaceTree(ctrl)

	mock.EXPECT().LoadTraverse(gomock.Any()).Return(errors.New("Expected Error"))

	failQuery := &query.Query{
		OpCode:   query.SELECT,
		TableKey: MAIN_TABLE_KEY,
		Select: query.QuerySelect{
			Limit: 2,
			Where: query.QueryWhere{
				OpCode: query.PREDICATE,
				Predicate: query.QueryPredicate{
					OpCode:   query.STR_EQ,
					Literals: []string{"Hi"},
					Keys:     []crdt.EntryName{"Entry A"},
				},
			},
		},
	}

	selector := makeNamespaceTreeSelect(mock)
	failQuery.Visit(selector)
	resp := selector.RunQuery()

	if resp.Msg != "error" {
		t.Error("Expected Msg error but received", resp.Msg)
	}

	if resp.Err == nil {
		t.Error("Expected response Err")
	}
}

func TestRunQuerySelectInvalid(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mock := NewMockNamespaceTree(ctrl)

	invalidQueries := []*query.Query{
		// Basically wrong.
		&query.Query{},
		&query.Query{OpCode: query.JOIN},
		// No limit
		&query.Query{
			Select: query.QuerySelect{
				Where: query.QueryWhere{
					OpCode: query.PREDICATE,
					Predicate: query.QueryPredicate{
						OpCode:   query.STR_EQ,
						Literals: []string{"Hi"},
						Keys:     []crdt.EntryName{"Entry A"},
					},
				},
			},
		},
		// No where OpCode
		&query.Query{
			Select: query.QuerySelect{
				Limit: 1,
				Where: query.QueryWhere{
					Predicate: query.QueryPredicate{
						OpCode:   query.STR_EQ,
						Literals: []string{"Hi"},
						Keys:     []crdt.EntryName{"Entry A"},
					},
				},
			},
		},
		// No predicate OpCode
		&query.Query{
			Select: query.QuerySelect{
				Limit: 1,
				Where: query.QueryWhere{
					OpCode: query.PREDICATE,
					Predicate: query.QueryPredicate{
						Literals: []string{"Hi"},
						Keys:     []crdt.EntryName{"Entry A"},
					},
				},
			},
		},
	}

	for _, q := range invalidQueries {
		selector := makeNamespaceTreeSelect(mock)
		q.Visit(selector)
		resp := selector.RunQuery()

		if resp.Msg != "error" {
			t.Error("Expected Msg error but received", resp.Msg)
		}

		if resp.Err == nil {
			t.Error("Expected response Err")
		}
	}
}

func rowsA() []crdt.Row {
	return []crdt.Row{
		crdt.MakeRow(map[crdt.EntryName]crdt.Entry{
			// TODO use user concepts to match only the Hi.
			"Entry A": crdt.MakeEntry([]crdt.Point{crdt.UnsignedPoint("Hi"), crdt.UnsignedPoint("Hello")}),
		}),
	}
}

func rowsB() []crdt.Row {
	return []crdt.Row{
		crdt.MakeRow(map[crdt.EntryName]crdt.Entry{
			"Entry B": crdt.MakeEntry([]crdt.Point{crdt.UnsignedPoint("Hi"), crdt.UnsignedPoint("Hello World")}),
		}),
	}
}

func rowsC() []crdt.Row {
	return []crdt.Row{
		crdt.MakeRow(map[crdt.EntryName]crdt.Entry{
			"Entry B": crdt.MakeEntry([]crdt.Point{crdt.UnsignedPoint("Hi"), crdt.UnsignedPoint("Hello Dude")}),
		}),
	}
}

func rowsD() []crdt.Row {
	return []crdt.Row{
		crdt.MakeRow(map[crdt.EntryName]crdt.Entry{
			"Entry C": crdt.MakeEntry([]crdt.Point{crdt.UnsignedPoint("Apple")}),
			"Entry D": crdt.MakeEntry([]crdt.Point{crdt.UnsignedPoint("Orange")}),
		}),
	}
}

func rowsE() []crdt.Row {
	return []crdt.Row{
		crdt.MakeRow(map[crdt.EntryName]crdt.Entry{
			"Entry E": crdt.MakeEntry([]crdt.Point{crdt.UnsignedPoint("Bus")}),
		}),
		crdt.MakeRow(map[crdt.EntryName]crdt.Entry{
			"Entry E": crdt.MakeEntry([]crdt.Point{crdt.UnsignedPoint("Train")}),
		}),
	}
}

func rowsF() []crdt.Row {
	return []crdt.Row{
		crdt.MakeRow(map[crdt.EntryName]crdt.Entry{
			"Entry F": crdt.MakeEntry([]crdt.Point{crdt.UnsignedPoint("This row"), crdt.UnsignedPoint("rocks")}),
		}),
	}
}

func rowsG() []crdt.Row {
	return []crdt.Row{
		crdt.MakeRow(map[crdt.EntryName]crdt.Entry{
			"Entry Q": crdt.MakeEntry([]crdt.Point{crdt.UnsignedPoint("Hi"), crdt.UnsignedPoint("Folks")}),
		}),
		crdt.MakeRow(map[crdt.EntryName]crdt.Entry{
			"Entry R": crdt.MakeEntry([]crdt.Point{crdt.UnsignedPoint("Wowzer")}),
		}),
		crdt.MakeRow(map[crdt.EntryName]crdt.Entry{
			"Entry S": crdt.MakeEntry([]crdt.Point{crdt.UnsignedPoint("Trumpet")}),
		}),
	}
}

func rowsH() []crdt.Row {
	unsignedRows := []crdt.Row{
		crdt.MakeRow(map[crdt.EntryName]crdt.Entry{
			"Entry WF": crdt.MakeEntry([]crdt.Point{crdt.UnsignedPoint("Super duper"), crdt.UnsignedPoint("yes!")}),
		}),
		crdt.MakeRow(map[crdt.EntryName]crdt.Entry{
			"Entry WP": crdt.MakeEntry([]crdt.Point{crdt.UnsignedPoint("Lolprets")}),
		}),
		crdt.MakeRow(map[crdt.EntryName]crdt.Entry{
			"Entry WG": crdt.MakeEntry([]crdt.Point{crdt.UnsignedPoint("Wunderbarrier")}),
		}),
	}

	hRows := make([]crdt.Row, len(unsignedRows))

	keys := []crypto.PrivateKey{__SELECT_PRIVATE_KEY}
	for i, row := range unsignedRows {
		signedRow := crdt.EmptyRow()
		row.ForeachEntry(func(entryName crdt.EntryName, entry crdt.Entry) {
			unsignedPoints := entry.GetValues()

			signedPoints := make([]crdt.Point, len(unsignedPoints))

			for i, p := range unsignedPoints {
				signed, err := crdt.SignedPoint(p.Text, keys)

				if err != nil {
					panic(err)
				}

				signedPoints[i] = signed
			}

			signedEntry := crdt.MakeEntry(signedPoints)
			signedRow = signedRow.JoinEntry(entryName, signedEntry)
		})
		hRows[i] = signedRow
	}

	return hRows
}

// Non matching rows.
func rowsZ() []crdt.Row {
	return []crdt.Row{
		crdt.MakeRow(map[crdt.EntryName]crdt.Entry{
			"Entry A": crdt.MakeEntry([]crdt.Point{crdt.UnsignedPoint("No"), crdt.UnsignedPoint("Match")}),
		}),
		crdt.MakeRow(map[crdt.EntryName]crdt.Entry{
			"Entry C": crdt.MakeEntry([]crdt.Point{crdt.UnsignedPoint("No"), crdt.UnsignedPoint("Match"), crdt.UnsignedPoint("Here")}),
			"Entry D": crdt.MakeEntry([]crdt.Point{crdt.UnsignedPoint("Nada!")}),
		}),
		crdt.MakeRow(map[crdt.EntryName]crdt.Entry{
			"Entry E": crdt.MakeEntry([]crdt.Point{crdt.UnsignedPoint("Horse")}),
		}),
	}
}

func tableA() crdt.Table {
	return mktable("A", rowsA())
}

func tableB() crdt.Table {
	return mktable("B", rowsB())
}

func tableC() crdt.Table {
	return mktable("C", rowsC())
}

func tableD() crdt.Table {
	return mktable("D", rowsD())
}

func tableE() crdt.Table {
	return mktable("E", rowsE())
}

func tableF() crdt.Table {
	return mktable("F", rowsF())
}

func tableG() crdt.Table {
	return mktable("G", rowsG())
}

func tableH() crdt.Table {
	return mktable("G", rowsH())
}

func tableZ() crdt.Table {
	return mktable("Z", rowsZ())
}

func streamA() []crdt.NamespaceStreamEntry {
	return makeTableStream(MAIN_TABLE_KEY, tableA())
}

func streamB() []crdt.NamespaceStreamEntry {
	return makeTableStream(MAIN_TABLE_KEY, tableB())
}

func streamC() []crdt.NamespaceStreamEntry {
	return makeTableStream(MAIN_TABLE_KEY, tableC())
}

func streamD() []crdt.NamespaceStreamEntry {
	return makeTableStream(MAIN_TABLE_KEY, tableD())
}

func streamE() []crdt.NamespaceStreamEntry {
	return makeTableStream(MAIN_TABLE_KEY, tableE())
}

func streamF() []crdt.NamespaceStreamEntry {
	return makeTableStream(MAIN_TABLE_KEY, tableF())
}

func streamG() []crdt.NamespaceStreamEntry {
	return makeTableStream(ALT_TABLE_KEY, tableG())
}

func streamH() []crdt.NamespaceStreamEntry {
	return makeTableStream(ALT_TABLE_KEY, tableH())
}

func feedNamespace(ntr api.NamespaceTreeReader) {
	ntr.ReadNamespace(mkselectns())
}

func mkselectns() crdt.Namespace {
	namespace := crdt.EmptyNamespace()
	mainTables := []crdt.Table{
		tableA(),
		tableB(),
		tableC(),
		tableD(),
		tableE(),
		tableF(),
		tableZ(),
	}
	altTables := []crdt.Table{
		tableG(),
		tableH(),
	}

	tables := map[crdt.TableName][]crdt.Table{
		MAIN_TABLE_KEY: mainTables,
		ALT_TABLE_KEY:  altTables,
	}

	for tableKey, ts := range tables {
		for _, t := range ts {
			namespace = namespace.JoinTable(tableKey, t)
		}
	}

	return namespace
}

func mktable(name string, rows []crdt.Row) crdt.Table {
	table := crdt.EmptyTable()

	for i, r := range rows {
		rowKey := crdt.RowName(fmt.Sprintf("Row %v%v", name, i))
		table = table.JoinRow(rowKey, r)
	}

	return table
}

const MAIN_TABLE_KEY = crdt.TableName("The Table")
const ALT_TABLE_KEY = crdt.TableName("Another table")

func makeTableStream(name crdt.TableName, table crdt.Table) []crdt.NamespaceStreamEntry {
	stream, invalid := crdt.MakeTableStream(name, table)

	panicOnInvalidNamespace(invalid)

	return stream
}

func makeNamespaceTreeSelect(namespace api.NamespaceTree) *eval.NamespaceTreeSelect {
	keyStore := &crypto.KeyStore{}
	return eval.MakeNamespaceTreeSelect(namespace, keyStore)
}

func init() {
	var err error
	__SELECT_PRIVATE_KEY, __SELECT_PUBLIC_KEY, err = crypto.GenerateKey()

	if err != nil {
		panic(err)
	}
}

func joinNamespaceStream(stream []crdt.NamespaceStreamEntry) []crdt.NamespaceStreamEntry {
	joined, invalid, err := crdt.JoinStreamEntries(stream)

	if err != nil {
		panic(err)
	}

	panicOnInvalidNamespace(invalid)

	return joined
}

func panicOnInvalidNamespace(invalid []crdt.InvalidNamespaceEntry) {
	invalidCount := len(invalid)
	if invalidCount > 0 {
		panic(fmt.Sprintf("%v invalid entries", invalidCount))
	}
}

var __SELECT_PRIVATE_KEY crypto.PrivateKey
var __SELECT_PUBLIC_KEY crypto.PublicKey
