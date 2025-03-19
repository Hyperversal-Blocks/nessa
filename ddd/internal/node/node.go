package node

type node struct {
}

type Node interface{}

func New() Node {
	return &node{}
}
