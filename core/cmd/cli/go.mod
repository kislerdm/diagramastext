module cli

go 1.19

require (
	github.com/kislerdm/diagramastext/core v0.0.2
	github.com/kislerdm/diagramastext/core/storage v0.0.1
)

require github.com/lib/pq v1.10.7 // indirect

replace (
	github.com/kislerdm/diagramastext/core v0.0.2 => ../../
	github.com/kislerdm/diagramastext/core/storage v0.0.1 => ../../storage
)
