# Memory image pattern

To increase performance, Godless shall use memory image pattern (outlined here)

# Why?

Memory image pattern is already a very good fit with how godless works.  Currently, we have one write goroutine, and many reader goroutines.  This solves the problem but is slow.  The question is understanding the solution space better, so Godless can become faster and so grow up into a real database.

# 12-factor

Although we will talk of a memory image, which refers to a resident memory image, and a fast local store, this will in fact be a provided by a plugable cache to reduce the number of assumptions made by godless regarding the user's architecture.

# How things are done now

## Writes

1. Client sends join query.
2. Godless writes namespace to IPFS.
3. Godless generates index for the join query.
4. Godless joins the new index with the old, and saves the index to IPFS.
5. Godless returns API response to client.

Step 4 is a bottleneck to all read processes.

## Reads

1. Client sends select query.
2. Godless reads index and discovers matching namespaces.
3. Godless queries namespaces for results.
4. Godless returns API response to client.

Step 2 will be blocked by Write step 4

# How we will do things in future

Our memory image is a list of Indices.  These will be represented as IPFS hashes which point to a serialized index.

## Writes

1. Client sends join query.
2. Godless writes namespace to IPFS.
3. Godless generates index for namespace and writes this to IPFS.
4. Godless persists the index to a memory image.
5. Godless returns API response to client

Steps 4 will lock the memory image

## Pulse

1. The pulse receives a tick
2. All indices in the memory image are joined to a new index and saved to IPFS
3. The memory image is overwritten with the composite index.
4. Wait till next tick

Step 3 will lock the memory image

## Reads

1. Client sends select query.
2. Godless reads memory image indices and discovers matching namespaces.
3. Godless queries namespaces for results.
4. Godless returns API response to client.

Step 2 will be blocked by write step 4 and pulse step 3

## Pubsub

Publication will advertise only the first index in the memory image, as this is the current composite.

Subscription will add a new index to the memory image, but the pulse will be responsible for joining this image to the others.
