package resolver

type InvitationStatus string

const (
	invitationStatusPending  InvitationStatus = "PENDING"
	invitationStatusAccepted InvitationStatus = "ACCEPTED"
	invitationStatusDeclined InvitationStatus = "DECLINED"
	invitationStatusRevoked  InvitationStatus = "REVOKED"
	invitationStatusExpired  InvitationStatus = "EXPIRED"
)
