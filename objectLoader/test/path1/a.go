package path1

type A struct {
	name string
}

func (a *A) GetName() string {
	return a.name
}
