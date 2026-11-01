package hookfacet

type State struct{}

func NewSessionState() *State {
	panic("niy")
}

func NewChildState(parent *State) *State {
	panic("niy")
}

func Get[T any](st *State) (T, bool) {
	panic("niy")
}

func Set[T any](st *State, value T) {
	panic("niy")
}
