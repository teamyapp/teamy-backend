package entity

type InvitationStatus int

const (
	Pending InvitationStatus = iota
	Accepted
	Declined
	Revoked
	Expired
)
