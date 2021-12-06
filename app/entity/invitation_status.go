package entity

type InvitationStatus int

const (
	InvitationStatusPending InvitationStatus = iota
	InvitationStatusAccepted
	InvitationStatusDeclined
	InvitationStatusRevoked
	InvitationStatusExpired
)
