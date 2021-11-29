package resolver

type User struct {
	Entity
}

func (u User) Name() string {
	panic("not implemented")
}

func (u User) ProfileURL() string {
	panic("not implemented")
}
