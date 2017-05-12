# TODO

1. Replication
2. Any additional mocking tests (I think a couple might be untested)
3. Write generator of testing data.
4. Functional & integration tests
5. User concepts (crypto!)

I want to do the replication bit before writing integration tests via docker because I think the API might change.  As a one man project, doing it in this order is less costly.

We now have an index that associates table names with IPFS content.  When we adopt user concepts, we shall have a 2-dimensional index via crypto keys too (so that users can easily find their data).  Regarding indexing, I want to keep things simple for now.  Need to get a working system first before we start optimising.

# Testing plan

## Using mocks and other unit tests

1. WebService
2. Query (and parsing)

## Using integration testing via docker

7. IPFSPeer

# Functional tests

8. Run HTTP join query
9. Run HTTP select query
10. Run multiple replicated instances
