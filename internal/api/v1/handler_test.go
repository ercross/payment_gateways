package v1

import (
	"bytes"
	"encoding/json"
	"github.com/ercross/payment_gateways/db"
	"github.com/ercross/payment_gateways/internal/api/v1/dto"
	"github.com/ercross/payment_gateways/internal/kafka"
	"github.com/ercross/payment_gateways/internal/logger"
	cache "github.com/ercross/payment_gateways/internal/redis"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestDeposit_Success(t *testing.T) {
	mockRepo := new(db.Mock)
	mockCache := new(cache.Mock)
	log, _ := logger.NewSilentLogger()
	publisher := &kafka.Mock{}

	depositRequest := dto.DepositRequest{
		UserID:   1,
		Amount:   100.0,
		Currency: "USD",
	}

	req, _ := http.NewRequest(http.MethodPost, "/deposit", bytes.NewReader(encodeJSON(depositRequest)))
	rr := httptest.NewRecorder()

	handler := handleDeposit(mockRepo, log, publisher, mockCache, "https://localhost:8080")
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
}

func TestWithdraw_InsufficientBalance(t *testing.T) {
	mockRepo := new(db.Mock)
	mockCache := new(cache.Mock)
	log, _ := logger.NewSilentLogger()
	publisher := &kafka.Mock{}

	withdrawRequest := dto.WithdrawalRequest{
		Amount:             200.0,
		UserID:             1,
		PaymentGatewayName: "Stripe",
		ReceivingAccount:   "odeyemi.t.e@gmail.com",
		AuthenticationCode: "123456",
	}

	req, _ := http.NewRequest(http.MethodPost, "/withdraw", bytes.NewReader(encodeJSON(withdrawRequest)))
	rr := httptest.NewRecorder()

	handler := initiateWithdrawal(mockRepo, log, publisher, mockCache, "https://localhost:8080")
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
	assert.Contains(t, rr.Body.String(), "You do not have sufficient balance")
}

func TestDepositCallback_Success(t *testing.T) {
	mockRepo := new(db.Mock)
	mockCache := new(cache.Mock)
	log, _ := logger.NewSilentLogger()
	publisher := &kafka.Mock{}

	callbackRequest := dto.TransactionStatusCallback{
		TransactionID: 1,
		Status:        "success",
	}

	req, _ := http.NewRequest(http.MethodPost, "/callback/deposit", bytes.NewReader(encodeJSON(callbackRequest)))
	rr := httptest.NewRecorder()

	handler := depositCallbackHandler(mockRepo, log, publisher, mockCache)
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
}

func TestWithdrawCallback_Failure(t *testing.T) {
	mockRepo := new(db.Mock)
	mockCache := new(cache.Mock)
	log, _ := logger.NewSilentLogger()
	publisher := &kafka.Mock{}

	callbackRequest := dto.TransactionStatusCallback{
		TransactionID: 2,
		Status:        "FAILED",
	}

	req, _ := http.NewRequest(http.MethodPost, "/callback/withdraw", bytes.NewReader(encodeJSON(callbackRequest)))
	rr := httptest.NewRecorder()

	handler := withdrawalCallbackHandler(mockRepo, log, publisher, mockCache)
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
}

func encodeJSON(data interface{}) []byte {
	b, _ := json.Marshal(data)
	return b
}

func TestDeposit_InvalidUserID(t *testing.T) {
	mockRepo := new(db.Mock)
	mockCache := new(cache.Mock)
	log, _ := logger.NewSilentLogger()
	publisher := &kafka.Mock{}

	depositRequest := dto.DepositRequest{
		UserID:   0, // Invalid user ID
		Amount:   100.0,
		Currency: "USD",
	}

	req, _ := http.NewRequest(http.MethodPost, "/deposit", bytes.NewReader(encodeJSON(depositRequest)))
	rr := httptest.NewRecorder()

	handler := handleDeposit(mockRepo, log, publisher, mockCache, "https://localhost:8080")
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
	assert.Contains(t, rr.Body.String(), "failed validation:")
}

func TestWithdraw_NegativeAmount(t *testing.T) {
	mockRepo := new(db.Mock)
	mockCache := new(cache.Mock)
	log, _ := logger.NewSilentLogger()
	publisher := &kafka.Mock{}

	withdrawRequest := dto.WithdrawalRequest{
		Amount:             -50.0, // Invalid amount
		UserID:             1,
		PaymentGatewayName: "Stripe",
		ReceivingAccount:   "",
		AuthenticationCode: "",
	}

	req, _ := http.NewRequest(http.MethodPost, "/withdraw", bytes.NewReader(encodeJSON(withdrawRequest)))
	rr := httptest.NewRecorder()

	handler := initiateWithdrawal(mockRepo, log, publisher, mockCache, "https://localhost:8080")
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
	assert.Contains(t, rr.Body.String(), "failed validation")
}
