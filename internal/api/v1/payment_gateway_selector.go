package v1

import (
	"errors"
	"fmt"
	"github.com/ercross/payment_gateways/db"
	"github.com/ercross/payment_gateways/internal/logger"
	"github.com/ercross/payment_gateways/internal/payment_gateways"
)

func selectPaymentGateway(repo db.Repository, userCountryID int, log *logger.Logger) (gatewayImpl gateways.PaymentGateway, err error) {

	priorityGateways, err := repo.GetGatewayPriorities(userCountryID)
	if err != nil {
		return gatewayImpl, fmt.Errorf("error getting gateway priorities: %w", err)
	}

	if priorityGateways == nil || len(priorityGateways) == 0 {

		// select fallback gateway
		gatewayImpl = gateways.GlobalDefault()
		if err = gatewayImpl.CheckAvailability(); err != nil {
			return gatewayImpl, fmt.Errorf("error checking default gateway availability: %w", err)
		}

		return gatewayImpl, nil

	} else {

		// select from available gateways by priority
		for _, pg := range priorityGateways {
			gatewayImpl, err = gateways.PaymentGatewayFromName(pg.Gateway.Name)
			if err != nil {
				log.Warn(err.Error(), logger.NewField("payment-gateway", pg.Gateway.Name))
				continue
			}
			err = gatewayImpl.CheckAvailability()
			if err != nil {
				log.Warn(err.Error(), logger.NewField("payment-gateway", pg.Gateway.Name))
				continue
			}
			return gatewayImpl, nil
		}
	}

	return nil, errors.New("no payment gateway available")
}
