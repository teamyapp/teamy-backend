package feature

type Toggles struct {
	EnableAuthorization bool
	EnableCache         bool
}

func NewStaticToggles() Toggles {
	return Toggles{
		EnableAuthorization: false,
		EnableCache:         false,
	}
}
