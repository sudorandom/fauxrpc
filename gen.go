package fauxrpc

const MaxNestedDepth = 20

type state struct {
	Depth int
}

// Increment to depth and reset layer-specific values (like IsKey)
func (st state) Inc() state {
	st.Depth++
	return st
}
