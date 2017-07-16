package functional_godless

import (
	"fmt"
	"sync"
	"testing"

	"github.com/johnny-morrice/godless/api"
	"github.com/johnny-morrice/godless/crdt"
	"github.com/johnny-morrice/godless/internal/testutil"
	"github.com/johnny-morrice/godless/query"
)

func RunRequestResultTests(t *testing.T, client api.Client, size int) {
	commands := makeJoins(size)
	queries := makeSelectReflects(size)

	commandWg := &sync.WaitGroup{}
	for _, c := range commands {
		commandWg.Add(1)
		command := c
		go func() {
			resp, err := client.Send(command)
			testutil.AssertNil(t, err)
			testutil.Assert(t, "Unexpected empty response", !resp.IsEmpty())
			commandWg.Done()
		}()
	}

	commandWg.Wait()

	queryWg := &sync.WaitGroup{}
	for _, q := range queries {
		queryWg.Add(1)
		query := q
		go func() {
			resp, err := client.Send(query)

			testutil.AssertNil(t, err)
			testutil.Assert(t, "Unexpected empty response", !resp.IsEmpty())
			hasResult := !resp.Namespace.IsEmpty()
			hasResult = hasResult || !resp.Index.IsEmpty()
			hasResult = hasResult || !crdt.IsNilPath(resp.Path)
			queryWg.Done()
		}()
	}

	queryWg.Wait()
}

func makeJoins(size int) []api.Request {
	joins := make([]api.Request, size)

	for i := 0; i < size; i++ {
		query := genJoinQuery(i)
		joins[i] = api.MakeQueryRequest(query)
	}

	return joins
}

func makeSelectReflects(size int) []api.Request {
	selects := make([]api.Request, size)

	for i := 0; i < size; i++ {
		query := genSelectQuery(i)
		selects[i] = api.MakeQueryRequest(query)
	}

	return selects
}

func genSelectQuery(seq int) *query.Query {
	// TODO use parametrised query.. once they exist :)
	queryText := fmt.Sprintf("select factory where str_eq(foreman, \"Foreman %d\")", seq)
	return forceCompile(queryText)
}

func genJoinQuery(seq int) *query.Query {
	// TODO use parametrised query.. once they exist :)
	queryText := fmt.Sprintf("join factory rows (@key=factory%d, foreman=\"Foreman %d\")", seq, seq)
	return forceCompile(queryText)
}

func forceCompile(source string) *query.Query {
	query, err := query.Compile(source)

	if err != nil {
		panic(err)
	}

	return query
}
