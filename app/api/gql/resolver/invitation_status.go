package resolver

type InvitationStatus string

const (
	Pending  InvitationStatus = "PENDING"
	Accepted InvitationStatus = "ACCEPTED"
	Declined InvitationStatus = "DECLINED"
	Revoked  InvitationStatus = "REVOKED"
	Expired  InvitationStatus = "EXPIRED"
)
