# Godless

A toy peer-to-peer database for the future internet.

What does that mean?

Godless is a way of sharing structured data between computers without using servers or the cloud.

Check [the wiki](https://github.com/johnny-morrice/godless/wiki) for a tutorial.

## Try it!

Run IPFS:

```
$ ipfs daemon --enable-pubsub-experiment
```

Run server:

```
$ godless store server --early --topics=godless
```

Any other godless server that uses the same topic will replicate all data.

The `--early` flag indicates that the server should fail if it can find no running IPFS daemon.

Run plumbing query command:

```
$ godless query plumbing --query 'join books rows (@key=book50, authorName="EL James", publisher="Don'\''t wanna know")'
2017/06/07 21:07:35 DEBUG HTTP POST to http://localhost:8085/api/query
message: "ok"
error: ""
type: 1
queryResponse: <
  namespace: <
  >
>

$ godless query plumbing --query 'select books where str_eq(@key, "book50") limit 10'     
2017/06/07 21:07:46 DEBUG HTTP POST to http://localhost:8085/api/query
message: "ok"
error: ""
type: 1
queryResponse: <
  namespace: <
    entries: <
      table: "books"
      row: "book50"
      entry: "authorName"
      points: "EL James"
    >
    entries: <
      table: "books"
      row: "book50"
      entry: "publisher"
      points: "Don't wanna know"
    >
  >
>
```

## How does it work?

Data is stored in a [CRDT](https://en.wikipedia.org/wiki/Conflict-free_replicated_data_type) namespace suitable for sharing between peers.  This is indexed using an another datastructure which itself is a CRDT.  Indexes and namespaces are arranged into a canonical order and saved to the [Interplanetary File System](http://ipfs.io/).  

The datastructures can be reassembled at another peer by looking up the IPFS hash.  Index hashes are shared over PubSub.

Crucially, data is signed using strong cryptography.  You can specify a key in your queries to sign (in joins) or verify (in selects).  This is crucial to maintaining data consistency in the face of arbitrary joins by other net users :).

## Installing

Godless is currently in alpha stage for Linux only.

### For everyone

Check out the [releases page](https://github.com/johnny-morrice/godless/releases).

Download signatures from [teoma.org.uk](https://teoma.org.uk/godless.html).

### For Golang programmers

```
go get -u github.com/johnny-morrice/godless/godless
```
