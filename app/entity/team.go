package entity

type Team struct {
	Entity
	Name          string
	LogoURL       string
	MemberUserIds []int
}
