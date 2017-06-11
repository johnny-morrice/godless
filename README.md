# Godless

A toy distributed database for the future internet.

Godless is a CRDT database, and query language, that uses [Interplanetary Filesystem](https://ipfs.io/) as a data store.

## Demo

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

## Installing

Godless is currently in alpha stage for Linux only.

### For everyone

Check out the [releases page](https://github.com/johnny-morrice/godless/releases).

### For Golang programmers

```
go get -u github.com/johnny-morrice/godless/godless
```
