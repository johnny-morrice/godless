# Godless

A toy distributed database for the future internet.

Godless is a CRDT database, and query language, that uses [Interplanetary Filesystem](https://ipfs.io/) as a data store.

## Demo

Run IPFS:

```
$ ipfs daemon --offline
```

Run server:

```
$ godless serve
```

Run plumbing query command:

```
$ godless query --query "join books rows (@key=book50, authorName='EL James', publisher='Don\'t wanna know')"
2017/04/20 15:57:24 DEBUG HTTP POST to http://localhost:8085/api/query/run
{
	"Err": null,
	"Msg": "ok",
	"Rows": null,
	"QueryId": ""
}

$ godless query --query "select books where str_eq(@key, 'book50') limit 1"                                  
2017/04/20 15:34:53 DEBUG HTTP POST to http://localhost:8085/api/query/run
{
	"Err": null,
	"Msg": "ok",
	"Rows": [
		{
			"Entries": {
				"authorName": [
					"EL James"
				],
				"publisher": [
					"Don't wanna know"
				]
			}
		}
	],
	"QueryId": ""
}
```

## Installing

Godless is currently in alpha stage for Linux only.

### For everyone

Check out the [releases page](https://github.com/johnny-morrice/godless/releases).

### For Golang programmers

```
go get github.com/johnny-morrice/godless
```
