package feature

type Toggles struct {
	EnableAuthorization bool
}

func NewStaticToggles() Toggles {
	return Toggles{
		EnableAuthorization: false,
	}
}
