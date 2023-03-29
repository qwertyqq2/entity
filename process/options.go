package process

import (
	"time"

	opts "github.com/qwertyqq2/entity"
)

type Option func(*impl)

func ResentInterval(resentInterval time.Duration) Option {
	return func(i *impl) {
		i.resendInterval = resentInterval
	}
}

func WaitRespInterval(waitRespInterval time.Duration) Option {
	return func(i *impl) {
		i.waitRespInterval = waitRespInterval
	}
}

func ReconnectInterval(reconnectInterval time.Duration) Option {
	return func(i *impl) {
		i.reconnectInterval = reconnectInterval
	}
}

func MaxMsgSize(msgSize int) Option {
	return func(i *impl) {
		i.maxMsgSize = msgSize
	}
}

func MaxWaitingConnection(maxWaitingConnection time.Duration) Option {
	return func(i *impl) {
		i.maxWaitingConnection = maxWaitingConnection
	}
}

func SendMessageMaxDelay(sendMessageMaxDelay time.Duration) Option {
	return func(i *impl) {
		i.sendMessageMaxDelay = sendMessageMaxDelay
	}
}

func SendMsgCutoff(sendMsgCutoff int) Option {
	return func(i *impl) {
		i.sendMsgCutoff = sendMsgCutoff
	}
}

// is better
func DefaultOpts() []Option {
	return []Option{
		ResentInterval(opts.ResentInterval),
		WaitRespInterval(opts.WaitSendInterval),
		ReconnectInterval(opts.ReconnectInterval),
		MaxWaitingConnection(opts.MaxWaitingConnection),
		MaxMsgSize(opts.MaxMsgSize),
		SendMessageMaxDelay(opts.SendMessageMaxDelay),
		SendMsgCutoff(opts.SendMsgCutoff),
	}
}
