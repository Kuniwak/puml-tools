How to Release
==============

Releases are published by the [Release workflow](../.github/workflows/release.yml),
which runs GoReleaser whenever a `x.y.z` tag is pushed. Do not create the GitHub
Release by hand — pushing the tag is what publishes it.

```console
$ git checkout master
$ git pull origin master

$ actionlint && ghalint run
$ goreleaser check
$ go test ./...

$ # Bump the version according to semantic versioning.
$ edit ./version/version.go

$ git add ./version/version.go
$ # Write the bumped version in place of x.x.x.
$ git commit -m "Ready to be vx.x.x"
$ git push

$ # Optional: check what GoReleaser will build, without publishing anything.
$ goreleaser release --clean --snapshot && open ./dist

$ git tag x.x.x
$ # Push only the tag you just created. `git push --tags` would push every
$ # local tag, including any left over from another clone.
$ git push origin x.x.x

$ # Wait for the Release workflow to succeed. Name the run explicitly, because
$ # the tag push also starts the Test workflow, and pass --exit-status so that a
$ # failed run stops the procedure instead of being reported as success.
$ gh run watch "$(gh run list --workflow Release --limit 1 --json databaseId -q '.[0].databaseId')" --exit-status
$ gh release view x.x.x -w
```

Next, update [Kuniwak/homebrew-tap](https://github.com/Kuniwak/homebrew-tap).

```console
$ cd path/to/homebrew-tap

$ # Update the checksum and the version embedded in the URL.
$ # If the release added, removed or renamed binaries, update the binary names listed in the formula too.
$ edit ./puml-tools.rb

$ git add ./puml-tools.rb
$ git commit -m "Update"
$ git push -u origin master
```

Finally, verify the published install path:

```console
$ brew install kuniwak/tap/puml-tools
```
