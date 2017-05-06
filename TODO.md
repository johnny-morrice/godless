# TODO

1. Unit tests
3. Rethink indexing
4. Replication
5. Functional & integration tests

I want to do the replication bit before writing integration tests via docker because I think the API might change.  As a one man project, doing it in this order is less costly.

Indexing should be rethought to make replication easy as possible.  Currently we store data and indexes in the same pages on IPFS.  This is harder to replicate than simply an index table.  

Imagine another peer downloads the index and wants to check it has access to all the same data.

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
