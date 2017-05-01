# Things to do

1. Unit tests
2. Namespace equality (get rid of hacky test eq functions)
3. Rethink indexing
4. Replication
5. Functional & integration tests

I want to do the replication bit before writing integration tests via docker because I think the API might change.  As a one man project, doing it in this order is less costly.

Indexing should be rethought to make replication easy as possible.  Currently we store data and indexes in the same pages on IPFS.  This is harder to replicate than simply an index table.  

Imagine another peer downloads the index and wants to check it has access to all the same data.

# Test priority order

## Using mocks and other unit tests

1. remoteNamespace
2. keyValueStore
3. NamespaceTreeSelect
4. NamespaceTreeJoin
5. WebService
6. Query (and parsing)
7. Namespace

Namespace is last.  Although it is crucial to data integrity, the program can basically do something without it working 100%.  If the other parts don't work, then there is literally nothing to execute.  Namespace is also functional, and so easier to reckon about mentally. I have confidence it is already more stable than some other parts.  Nevertheless Namespace certainly needs solid testing before I can claim any stability.

## Using integration testing via docker

7. IPFSPeer

# Functional tests

8. Run HTTP join query
9. Run HTTP select query
10. Run multiple replicated instances
