package util

var (
	// prefix of nodes at various levels
	DefaultLevelNames []string = []string{"server", "rack", "room", "zone", "region", "cloud"}

	// prefix of root node
	DefaultRootName string = "root"

	// default prefix of a node at a level
	DefaultLevelName string = "level"

	// Default weight
	DefaultWeight int = 1

	// Max weight
	MaxWeight int = 100

	// Min weight
	MinWeight int = 1
)
