How to Test
===========

Running the Tests
-----------------

```console
$ go test ./...
```


Using the Built Binaries
------------------------

In most cases `go run ./tools/<id>/main.go` is enough.
If you need a built binary, use GoReleaser to build it under `./bin` and run it from there.

```console
$ goreleaser build --clean --snapshot --single-target --id <id> -o ./bin
$ ls ./bin
<id>
```
