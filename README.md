# isodb - isolated db (experimental)

`isodb` could be considered as a simple `document` database build on top of a `git-like` data strcture.

Operations always take a `commit` pointer (a sha256 hash) with a `root` folder, then, document keys are mapped into sub-folders of this root until the `leaf file` is found.

At this point, the `leaf file` has an `edge` named `blob` which points to the content of that `document`.

The whole datastructure is immutable and allow for multiple writers to work on the same database. If their changes are different they will end-up with different `commits` (different sha256 hash). This will trigger a conflict resolution (not yet implemented).

The database also allows for any `sha256` reference to have a human readable name. Updating this `ref` is atomic and has `cas` semantics.

## Why?

Because I like to work on systems where the nodes might spend hours,days,weeks or even months, without having a chance to sync their writes. Allowing applications to have some guarantees even in such scenarios is very uself (think of a farmer without 3G or satellite connection). They should be able to have sane way to keep information about their systems while they are disconnected.

Eventually `isodb` will have a turing interpreter with its code being stored as another document. So in this scenario conflict resolution could run a pre-defined script which given the same inputs will generate the same output regardless of where they are executed.

On top of that, commits might be `signed` so not only commits are immutable but can be verified without having to sync with the person who created it in the first place.

## Why not?

The database must not be the fastest one and if you don't have latency issues between your process and your datastorage and using locks to protect the documents/rows is feasible, then there is not reason to use `isodb`.

Just answer the question: Should your system being able to accept writes even when the host is completely disconnected from others? If yes, then maybe `isodb` is for you.

## How it works?

Check `repo_test.go` to have an idea of how to use it directly (API needs polishing, so don't judge)

## Prior art

- CouchDB
- Couchbase
- Noms
- Git
- Fossil

CouchDB and Couchbase aren't specifically desigend to support the days/weeks/months of disconnected work.
