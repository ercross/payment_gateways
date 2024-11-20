package gateways

import (
	"errors"
	"github.com/ercross/payment_gateways/db"
)

var (
	ErrUnknownPaymentGateway = errors.New("unknown payment gateway")

	// ErrPaymentGatewayNotResponding wraps any error returned from PaymentGateway.CheckAvailability
	ErrPaymentGatewayNotResponding = errors.New("payment gateway not responding")
)

// PaymentGateway represents a payment gateway
type PaymentGateway interface {

	// Name must be a unique name corresponding with the PaymentGateway name as saved in DB
	Name() string
	GenerateDepositCheckoutSessionData(trx db.Transaction, callbackUrl string) (sessionData any, err error)
	RegisterWithdrawal(trx db.Transaction, callbackUrl, receivingAccount string) error

	// CheckAvailability sends a liveness probe to this PaymentGateway
	CheckAvailability() error
}

func GlobalDefault() PaymentGateway {
	return new(Stripe)
}

func PaymentGatewayFromName(name string) (PaymentGateway, error) {
	switch name {
	case "stripe":
		return new(Stripe), nil
	case "paypal":
		return new(PayPal), nil
	default:
		return nil, ErrUnknownPaymentGateway
	}
}
