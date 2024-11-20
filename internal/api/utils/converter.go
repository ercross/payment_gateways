package utils

import (
	"github.com/ercross/payment_gateways/db"
	"github.com/ercross/payment_gateways/internal/api/v1/dto"
)

func ConvertDepositRequestToTransaction(dp dto.DepositRequest) db.Transaction {
	return db.Transaction{
		Amount:   dp.Amount,
		Type:     "deposit",
		Status:   "pending",
		UserID:   dp.UserID,
		Currency: dp.Currency,
	}
}

func ConvertWithdrawalRequestToTransaction(wr dto.WithdrawalRequest) db.Transaction {
	return db.Transaction{
		Amount: wr.Amount,
		Type:   "withdrawal",
		Status: "pending",
		UserID: wr.UserID,
	}
}
