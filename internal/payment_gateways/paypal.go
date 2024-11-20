package gateways

import (
	"github.com/ercross/payment_gateways/db"
)

type PayPal struct {
}

func (p *PayPal) Name() string {
	return "paypal"
}

func (p *PayPal) CheckAvailability() error {
	return nil
}

func (p *PayPal) GenerateDepositCheckoutSessionData(trx db.Transaction, callbackUrl string) (sessionData any, err error) {
	return nil, nil
}

func (p *PayPal) RegisterWithdrawal(trx db.Transaction, callbackUrl, receivingAccount string) error {
	return nil
}
