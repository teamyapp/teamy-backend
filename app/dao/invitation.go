package dao

type Invitation interface {
	FindInvitation(id uint64) (Invitation, error)
}
