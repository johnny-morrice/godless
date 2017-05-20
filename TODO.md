# TODO

1. Protocol buffers
2. More testing (Query parsing, IPFSPeer mocks, WebService mocks)
3. Replication
4. Write generator of functional testing data.
5. Functional & integration tests
6. User concepts (crypto!)
7. libP2P client
8. Client console (and web console?)

We now have an index that associates table names with IPFS content.  When we adopt user concepts, we shall have a 2-dimensional index via crypto keys too (so that users can easily find their data).  Regarding indexing, I want to keep things simple for now.  Need to get a working system first before we start optimising.

The client currently sends all requests via webservice.  However, replication and queries should also be handled by libp2p.  We shall keep this webservice though, to facilitate administration.

# Testing plan

## Using integration testing via docker

7. IPFSPeer

# Functional tests

8. Run HTTP join query
9. Run HTTP select query
10. Run multiple replicated instances
