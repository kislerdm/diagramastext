module httpserver

go 1.19

require (
github.com/kislerdm/diagramastext/server/pkg/core v0.0.1
github.com/kislerdm/diagramastext/server/pkg/rendering/plantuml v0.0.1
github.com/kislerdm/diagramastext/server v0.0.1
)

replace (
	github.com/kislerdm/diagramastext/server/pkg/core v0.0.1 => ../../pkg/core
    github.com/kislerdm/diagramastext/server/pkg/rendering/plantuml v0.0.1 => ../../pkg/rendering/plantuml
    github.com/kislerdm/diagramastext/server v0.0.1 => ../..
)

