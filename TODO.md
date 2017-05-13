# TODO

1. Unit tests
3. Rethink indexing
4. Replication
5. libP2P client
6. Functional & integration tests

I want to do the replication bit before writing integration tests via docker because I think the API might change.  As a one man project, doing it in this order is less costly.

Indexing should be rethought to make replication easy as possible.  Currently we store data and indexes in the same pages on IPFS.  This is harder to replicate than simply an index table.  

Imagine another peer downloads the index and wants to check it has access to all the same data.

The client currently sends all requests via webservice.  However, replication and queries should also be handled by libp2p.  We shall keep this webservice though, to facilitate administration.

I am working on a reflection API that can dump data and info on the running system.  This will be very useful for debugging & integration testing, and might serve the basis of a metrics feature (although for a toy program I won't go that far.) I don't know if the reflection API should be available over P2P protocols.  Probably not!

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
