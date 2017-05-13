# TODO

1. Protocol buffers
2. Replication
3. Any additional mocking tests (I think a couple might be untested)
4. Write generator of testing data.
5. libP2P client
6. Functional & integration tests
7. User concepts (crypto!)

We now have an index that associates table names with IPFS content.  When we adopt user concepts, we shall have a 2-dimensional index via crypto keys too (so that users can easily find their data).  Regarding indexing, I want to keep things simple for now.  Need to get a working system first before we start optimising.

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
