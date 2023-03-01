module httpserver

go 1.19

require (
	github.com/kislerdm/diagramastext/server v0.0.1
)

replace (
    github.com/kislerdm/diagramastext/server v0.0.1 => ../..
)
