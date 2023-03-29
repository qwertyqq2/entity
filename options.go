package entity

import "time"

const (
	SendMessageMaxDelay  = 10 * time.Second
	SendMsgCutoff        = 150
	MaxMsgSize           = 50
	ResentInterval       = 1 * time.Second
	WaitSendInterval     = 4 * time.Second
	ReconnectInterval    = 1 * time.Second
	MaxWaitingConnection = 100 * time.Second
)
