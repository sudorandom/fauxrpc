package protocel

type nodeKind = int32

const (
	CELKind      nodeKind = iota
	MessageKind  nodeKind = iota
	RepeatedKind nodeKind = iota
	MapKind      nodeKind = iota
)

type Node interface {
	Kind() nodeKind
}

type nodeCEL string

func CEL(expr string) nodeCEL {
	return nodeCEL(expr)
}

func (nodeCEL) Kind() nodeKind {
	return CELKind
}

type nodeMessage map[string]Node

func Message(fields map[string]Node) nodeMessage {
	return nodeMessage(fields)
}

func (nodeMessage) Kind() nodeKind {
	return MessageKind
}

type nodeRepeated []Node

func Repeated(nodes []Node) nodeRepeated {
	return nodeRepeated(nodes)
}

func (nodeRepeated) Kind() nodeKind {
	return RepeatedKind
}

type nodeMap map[Node]Node

func Map(nodes map[Node]Node) nodeMap {
	return nodeMap(nodes)
}

func (nodeMap) Kind() nodeKind {
	return MapKind
}
