# Calcifer

[![Go Reference](https://pkg.go.dev/badge/github.com/RadiopaperInc/calcifer.svg)](https://pkg.go.dev/github.com/RadiopaperInc/calcifer)

Calcifer is Radiopaper's ODM (Object-Document Mapping) library, written in Go, targeting Google Cloud Firestore.

Planned features include:

* Foreign-key relations with (optionally) cascading reads, writes, deletes.
* Transactional and asynchronous denormalization based on declarative struct tags. 
* Computed document properties.
* Model history bookkeeping and visualization.
* Smart retries for reads when Firestore is unavailable.
* Smart retries for writes when an idempotency key is provided.
