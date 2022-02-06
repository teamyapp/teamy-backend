package entity

type InvitationStatus string

const InvitationStatusPending InvitationStatus = "PENDING"
const InvitationStatusAccepted InvitationStatus = "ACCEPTED"
const InvitationStatusInvoked InvitationStatus = "REVOKED"
const InvitationStatusDeclined InvitationStatus = "DECLINED"
const InvitationStatusExpired InvitationStatus = "EXPIRED"
