package mag

// Change represents a change to be applied to a map or list.
type Change interface {
	Run() error
}
