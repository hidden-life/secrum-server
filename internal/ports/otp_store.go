package ports

import "context"

// OTPProvider defines the interface for delivering OTP codes to destintaion.
type OTPProvider interface {
	Deliver(context.Context, string, string) error
}

// OTPStore defines the interface for storing and verifying OTP challenges.
type OTPStore interface {
	SaveChallenge(context.Context, string, string) (string, error)
	VerifyAndConsume(context.Context, string, string) (string, bool, error)
}
