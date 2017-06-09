# TODO

1. Replication & index unit test.
2. Client console
3. Remaining unit tests
4. Integration tests
5. User concepts
6. Functional testing

Replication is going to be handled by simple polling of an IPNS name.

This name should point to an index.

The peer will then merge the indices and save it as its own index.

Indices should now be stable wrt Content Addressing but somehow (well, I was lazy) this is not unit tested.  Therefore ensuring we have proper unit tests for the index is a prerequisite for the replication task being completed.

Gavin has kindly offered to work on a terminal console for Godless.  I will pick out a terminal library, and fix up some minor issues before handing this over to him.

One of these issues concerns logging.  Currently we use the Golang standard logger, which outputs to stdout (I think) and does not support logging levels.  A client console should not log data to stdout, so I need to grab a better logging library where this stuff could be disabled or at least turned way down.

The plumbing command should also log to stderr, otherwise it would not be useful to incorporate into a script.  The plumbing command should also support protobuf binary output.

I will work on these client issues to enable Gavin's work, and then begin on replication.

# Testing plan

## Using integration testing via docker

7. IPFSPeer
8. Replication

# Functional tests

8. Run join query via plumbing
9. Run select query via plumbing
10. Run multiple replicated instances
