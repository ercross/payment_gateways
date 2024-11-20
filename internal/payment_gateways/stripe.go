package gateways

import (
	"github.com/ercross/payment_gateways/db"
)

type Stripe struct {
}

func (s *Stripe) Name() string {
	return "stripe"
}

func (s *Stripe) CheckAvailability() error {
	return nil
}

func (s *Stripe) GenerateDepositCheckoutSessionData(trx db.Transaction, callbackUrl string) (sessionData any, err error) {
	return map[string]interface{}{
		"callback_url":   callbackUrl,
		"another_random": "value",
	}, nil
}

func (s *Stripe) RegisterWithdrawal(trx db.Transaction, callbackUrl, receivingAccount string) error {
	return nil
}
