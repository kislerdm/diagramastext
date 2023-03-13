package port

// C4ContainersGraph defines the containers and relations for C4 Container diagram's graph.
type C4ContainersGraph struct {
	Containers []*Container
	Rels       []*Rel
	Title      string
	Footer     string
	WithLegend bool
}

// Container C4 container definition.
type Container struct {
	ID          string
	Label       string
	Technology  string
	Description string
	System      string
	IsQueue     bool
	IsDatabase  bool
	IsUser      bool
	IsExternal  bool
}

// Rel containers relations.
type Rel struct {
	From        string
	To          string
	Technology  string
	Description string
}
