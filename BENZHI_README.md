# KilnGuard

KilnGuard is a local control simulator for cement kiln combustion, draft,
clinker cooling, bypass handling, and safety interlocks.

Run the service with `go run ./cmd/kilnguard`. It listens on
`127.0.0.1:21207` by default and stores its event journal below `data/`.
The operator pages are `/kiln`, `/burner`, `/cooler`, and `/incidents`.

The project vendors all Go dependencies. Offline validation can use
`go test -mod=vendor ./...`, `go build -mod=vendor ./...`, and
`go vet -mod=vendor ./...`.
