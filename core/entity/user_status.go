package entity

type UserStatus string

const (
	UserStatusInitialized  UserStatus = "Initialized"
	UserStatusConnected    UserStatus = "Connected"
	UserStatusDisConnected UserStatus = "DisConnected"
)
