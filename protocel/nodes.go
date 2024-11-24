package protocel

type nodeKind = int32

const (
	CELFieldKind nodeKind = iota
	MessageKind  nodeKind = iota
	RepeatedKind nodeKind = iota
)

type Node interface {
	Kind() nodeKind
}

type nodeCEL string

func CEL(expr string) nodeCEL {
	return nodeCEL(expr)
}

func (nodeCEL) Kind() nodeKind {
	return CELFieldKind
}

type nodeMessage map[string]Node

func Message(fields map[string]Node) nodeMessage {
	return nodeMessage(fields)
}

func (nodeMessage) Kind() nodeKind {
	return MessageKind
}

type repeated []Node

func Repeated(nodes []Node) repeated {
	return repeated(nodes)
}

func (repeated) Kind() nodeKind {
	return RepeatedKind
}
