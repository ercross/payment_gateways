package v1

import (
	"context"
	"errors"
	"github.com/ercross/payment_gateways/db"
	"github.com/ercross/payment_gateways/internal/api/utils"
	"github.com/ercross/payment_gateways/internal/api/v1/dto"
	"github.com/ercross/payment_gateways/internal/kafka"
	"github.com/ercross/payment_gateways/internal/logger"
	"github.com/ercross/payment_gateways/internal/payment_gateways"
	cache "github.com/ercross/payment_gateways/internal/redis"
	"github.com/ercross/payment_gateways/internal/services"
	"net/http"
	"strings"
	"time"
)

var (
	internalServerErrorMsg = "Server encountered an error and is unable to process your request. Please try again."
)

// DepositHandler handles deposit requests (feel free to update how user is passed to the request)
// Sample Request (POST /deposit):
//
//	{
//	    "amount": 100.00,
//	    "user_id": 1,
//	    "currency": "EUR"
//	}
func handleDeposit(
	repo db.Repository,
	log *logger.Logger,
	publisher kafka.EventPublisher,
	dstrCache cache.DistributedCache,
	baseURL string,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		dataFormat := utils.DetermineResponseContentDataType(r)

		// validate request
		var depositRequest dto.DepositRequest
		err := utils.DecodeRequest(r, &depositRequest)
		if err != nil {
			sendAPIResponse(w, r, http.StatusBadRequest, err.Error(), nil, dataFormat)
			return
		}

		if err = utils.ValidateDTO(depositRequest, utils.ContentDataTypeToTag[dataFormat]); err != nil {
			sendAPIResponse(w, r, http.StatusBadRequest, err.Error(), nil, dataFormat)
			return
		}
		trx := utils.ConvertDepositRequestToTransaction(depositRequest)

		// get user country
		var user db.User
		cacheKey := cache.ConstructUserIDKey(depositRequest.UserID)
		err = dstrCache.Get(r.Context(), cacheKey, &user)
		if err != nil && !errors.Is(err, cache.ErrKeyNotFound) {
			sendAPIResponse(w, r, http.StatusInternalServerError, internalServerErrorMsg, nil, dataFormat)
			log.Error("failed to get user", logger.ComponentRedis,
				logger.NewField("Error", err.Error()), logger.NewField("Request-ID", requestID(r)))
			return
		}
		if errors.Is(err, cache.ErrKeyNotFound) {
			user, err = repo.GetUserByID(depositRequest.UserID)
			if err != nil {
				sendAPIResponse(w, r, http.StatusInternalServerError, internalServerErrorMsg, nil, dataFormat)
			}
		}
		trx.CountryName = user.Country.Name

		// select payment gateway
		gatewayImpl, err := selectPaymentGateway(repo, user.Country.ID, log)
		if err != nil {
			sendAPIResponse(w, r, http.StatusInternalServerError, internalServerErrorMsg, nil, dataFormat)
			log.Error("failed to get user", logger.ComponentDatabase,
				logger.NewField("Error", err.Error()), logger.NewField("Request-ID", requestID(r)))
			return
		}

		// create trx
		trx.GatewayName = gatewayImpl.Name()

		lockKey := constructDepositLockKey(trx.Amount, trx.UserID, trx.Currency, trx.GatewayName)
		lock, err := dstrCache.AcquireLock(r.Context(), lockKey)
		if err != nil {
			sendAPIResponse(w, r, http.StatusConflict, "Transaction already being processed", nil, dataFormat)
			return
		}
		defer dstrCache.ReleaseLock(lock)

		trxID, err := repo.CreateTransaction(trx)
		if err != nil {
			sendAPIResponse(w, r, http.StatusInternalServerError, internalServerErrorMsg, nil, dataFormat)
			log.Error("failed to create transaction", logger.ComponentDatabase,
				logger.NewField("Error", err.Error()), logger.NewField("Request-ID", requestID(r)))
			return
		}
		trx.ID = trxID

		// prepare response
		sessionData, err := gatewayImpl.GenerateDepositCheckoutSessionData(trx, constructDepositCallbackUrl(baseURL, trx.ID))
		if err != nil {
			sendAPIResponse(w, r, http.StatusInternalServerError, internalServerErrorMsg, nil, dataFormat)
			log.Error("failed to generate deposit checkout session data", logger.NewField("Payment-Gateway", gatewayImpl.Name()),
				logger.NewField("Error", err.Error()), logger.NewField("Request-ID", requestID(r)))
			return
		}

		sendAPIResponse(w, r, http.StatusOK, "Success", sessionData, dataFormat)

		// cache transaction for faster retrieval
		cacheKey = cache.ConstructTransactionIDKey(trx.ID)
		err = dstrCache.Save(cacheKey, trx, time.Minute*5)
		if err != nil {
			log.Warn("failed to cache transaction", logger.ComponentRedis,
				logger.NewField("Error", err.Error()), logger.NewField("Request-ID", requestID(r)))
		}

		// log to kafka
		encryptedSessionData, err := services.MaskData(sessionData)
		if err != nil {
			log.Warn("error encrypting deposit session data", logger.NewField("Error", err.Error()), logger.NewField("Request-ID", requestID(r)))
			return
		}
		err = publisher.PublishTransaction(context.Background(), trx.ID, encryptedSessionData, dataFormat)
		if err != nil {
			log.Warn("error publishing deposit transaction", logger.ComponentKafka,
				logger.NewField("Error", err.Error()), logger.NewField("Request-ID", requestID(r)), logger.NewField("transaction-ID", trxID))
			return
		}
		log.Info("Deposit transaction event published to Kafka", logger.ComponentKafka, logger.NewField("Request-ID", requestID(r)), logger.NewField("transaction-ID", trxID))
	}
}

// WithdrawalHandler handles withdrawal requests (feel free to update how user is passed to the request)
// Sample Request (POST /deposit):
//
//	{
//	    "amount": 100.00,
//	    "user_id": 1,
//	}
func initiateWithdrawal(
	repo db.Repository,
	log *logger.Logger,
	publisher kafka.EventPublisher,
	dstrCache cache.DistributedCache,
	baseURL string,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		dataFormat := utils.DetermineResponseContentDataType(r)

		var withdrawalRequest dto.WithdrawalRequest
		err := utils.DecodeRequest(r, &withdrawalRequest)
		if err != nil {
			sendAPIResponse(w, r, http.StatusBadRequest, err.Error(), nil, dataFormat)
			return
		}

		if err = utils.ValidateDTO(withdrawalRequest, utils.ContentDataTypeToTag[dataFormat]); err != nil {
			sendAPIResponse(w, r, http.StatusBadRequest, err.Error(), nil, dataFormat)
			return
		}

		userAccount, err := repo.GetUserAccount(withdrawalRequest.UserID)
		if err != nil {
			sendAPIResponse(w, r, http.StatusInternalServerError, internalServerErrorMsg, nil, dataFormat)
			log.Error("failed to get user account", logger.ComponentDatabase, logger.NewField("Error", err.Error()),
				logger.NewField("Request-ID", requestID(r)), logger.NewField("Request-ID", withdrawalRequest.UserID))
			return
		}

		if userAccount.Balance < withdrawalRequest.Amount {
			sendAPIResponse(w, r, http.StatusBadRequest, "You do not have sufficient balance", nil, dataFormat)
			return
		}

		err = services.VerifyUserAuthenticatorCode(withdrawalRequest.UserID, withdrawalRequest.AuthenticationCode)
		if err != nil {
			sendAPIResponse(w, r, http.StatusUnauthorized, "Incorrect authentication code", nil, dataFormat)
			return
		}

		gatewayImpl, err := gateways.PaymentGatewayFromName(withdrawalRequest.PaymentGatewayName)
		if err != nil {
			sendAPIResponse(w, r, http.StatusUnprocessableEntity, "Unknown payment gateway", nil, dataFormat)
			return
		}
		trx := utils.ConvertWithdrawalRequestToTransaction(withdrawalRequest)

		trxID, err := repo.CreateTransaction(trx)
		if err != nil {
			sendAPIResponse(w, r, http.StatusInternalServerError, internalServerErrorMsg, nil, dataFormat)
			log.Error("failed to create transaction", logger.ComponentDatabase,
				logger.NewField("Error", err.Error()), logger.NewField("Request-ID", requestID(r)))
			return
		}
		trx.ID = trxID

		err = gatewayImpl.RegisterWithdrawal(trx, constructWithdrawalCallbackUrl(baseURL, trx.ID), withdrawalRequest.ReceivingAccount)
		if err != nil {
			sendAPIResponse(w, r, http.StatusInternalServerError, internalServerErrorMsg, nil, dataFormat)
			log.Error("failed to register withdrawal", logger.NewField("Payment-Gateway", gatewayImpl.Name()),
				logger.NewField("Error", err.Error()), logger.NewField("Request-ID", requestID(r)))
			return
		}

		sendAPIResponse(w, r, http.StatusOK, "Your withdrawal has been registered and will be processed shortly", trx, dataFormat)

		// cache transaction for faster retrieval
		cacheKey := cache.ConstructTransactionIDKey(trx.ID)
		err = dstrCache.Save(cacheKey, trx, time.Minute*5)
		if err != nil {
			log.Warn("failed to cache transaction", logger.ComponentRedis,
				logger.NewField("Error", err.Error()), logger.NewField("Request-ID", requestID(r)))
		}

		// log to kafka
		encryptedTrx, err := services.MaskData(trx)
		if err != nil {
			log.Warn("error encrypting withdrawal transaction", logger.NewField("Error", err.Error()), logger.NewField("Request-ID", requestID(r)))
			return
		}
		err = publisher.PublishTransaction(context.Background(), trx.ID, encryptedTrx, dataFormat)
		if err != nil {
			log.Warn("error publishing withdrawal transaction", logger.ComponentKafka,
				logger.NewField("Error", err.Error()), logger.NewField("Request-ID", requestID(r)), logger.NewField("transaction-ID", trxID))
			return
		}
		log.Info("Withdrawal initiation transaction event published to Kafka", logger.ComponentKafka, logger.NewField("Request-ID", requestID(r)), logger.NewField("transaction-ID", trxID))
	}
}

func depositCallbackHandler(repo db.Repository, log *logger.Logger, publisher kafka.EventPublisher, dstrCache cache.DistributedCache) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		dataFormat := utils.DetermineResponseContentDataType(r)

		// Parse request
		var callbackRequest dto.TransactionStatusCallback
		err := utils.DecodeRequest(r, &callbackRequest)
		if err != nil {
			sendAPIResponse(w, r, http.StatusBadRequest, err.Error(), nil, dataFormat)
			log.Error("failed to parse deposit callback request", logger.NewField("Error", err.Error()))
			return
		}

		if err := utils.ValidateDTO(callbackRequest, utils.ContentDataTypeToTag[dataFormat]); err != nil {
			sendAPIResponse(w, r, http.StatusBadRequest, err.Error(), nil, dataFormat)
			log.Error("invalid deposit callback request", logger.NewField("Error", err.Error()))
			return
		}

		var trx db.Transaction
		trx.ID = callbackRequest.TransactionID
		err = dstrCache.Get(r.Context(), cache.ConstructTransactionIDKey(trx.ID), &trx)
		if err != nil && !errors.Is(err, cache.ErrKeyNotFound) {
			log.Warn("failed to fetch transaction", logger.ComponentRedis, logger.NewField("Error", err.Error()))
		}
		if errors.Is(err, cache.ErrKeyNotFound) {
			trx, err = repo.GetTransactionByID(trx.ID)
			if err != nil {
				if errors.Is(err, db.ErrDataNotFound) {
					sendAPIResponse(w, r, http.StatusNotFound, "Transaction not found", nil, dataFormat)
					return
				}

				sendAPIResponse(w, r, http.StatusInternalServerError, internalServerErrorMsg, nil, dataFormat)
				log.Error("failed to fetch transaction", logger.NewField("Error", err.Error()))
				return
			}
		}

		// Update transaction status
		if err := repo.UpdateTransactionStatus(trx.ID, callbackRequest.Status); err != nil {
			sendAPIResponse(w, r, http.StatusInternalServerError, "", nil, dataFormat)
			log.Error("failed to update transaction status", logger.NewField("Error", err.Error()))
			return
		}

		// Additional actions based on status
		if strings.ToLower(callbackRequest.Status) == "success" {
			err = repo.UpdateUserBalance(trx.UserID, trx.Amount)
			if err != nil {
				log.Error("failed to update user balance", logger.NewField("Error", err.Error()))
				sendAPIResponse(w, r, http.StatusInternalServerError, internalServerErrorMsg, nil, dataFormat)
				return
			}
		}

		sendAPIResponse(w, r, http.StatusOK, "Deposit transaction status updated successfully", nil, dataFormat)

		cacheKey := cache.ConstructTransactionIDKey(trx.ID)
		_ = dstrCache.Delete(cacheKey)

		// log to kafka
		encryptedTrx, err := services.MaskData(trx)
		if err != nil {
			log.Warn("error encrypting deposit transaction update", logger.NewField("Error", err.Error()), logger.NewField("Request-ID", requestID(r)))
			return
		}
		err = publisher.PublishTransaction(context.Background(), trx.ID, encryptedTrx, dataFormat)
		if err != nil {
			log.Warn("error publishing deposit transaction status update", logger.ComponentKafka,
				logger.NewField("Error", err.Error()), logger.NewField("Request-ID", requestID(r)), logger.NewField("transaction-ID", trx.ID))
			return
		}
		log.Info("Deposit transaction status update event published to Kafka", logger.ComponentKafka, logger.NewField("Request-ID", requestID(r)), logger.NewField("transaction-ID", trx.ID))
	}
}

func withdrawalCallbackHandler(repo db.Repository, log *logger.Logger, publisher kafka.EventPublisher, dstrCache cache.DistributedCache) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		dataFormat := utils.DetermineResponseContentDataType(r)

		// Parse request
		var callbackRequest dto.TransactionStatusCallback
		err := utils.DecodeRequest(r, &callbackRequest)
		if err != nil {
			sendAPIResponse(w, r, http.StatusBadRequest, err.Error(), nil, dataFormat)
			log.Error("failed to parse deposit callback request", logger.NewField("Error", err.Error()))
			return
		}

		if err := utils.ValidateDTO(callbackRequest, utils.ContentDataTypeToTag[dataFormat]); err != nil {
			sendAPIResponse(w, r, http.StatusBadRequest, err.Error(), nil, dataFormat)
			log.Error("invalid deposit callback request", logger.NewField("Error", err.Error()))
			return
		}

		var trx db.Transaction
		trx.ID = callbackRequest.TransactionID
		err = dstrCache.Get(r.Context(), cache.ConstructTransactionIDKey(trx.ID), &trx)
		if err != nil && !errors.Is(err, cache.ErrKeyNotFound) {
			log.Warn("failed to fetch transaction", logger.ComponentRedis, logger.NewField("Error", err.Error()))
		}
		if errors.Is(err, cache.ErrKeyNotFound) {
			trx, err = repo.GetTransactionByID(trx.ID)
			if err != nil {
				sendAPIResponse(w, r, http.StatusNotFound, "Transaction not found", nil, dataFormat)
				log.Error("transaction not found", logger.NewField("TransactionID", callbackRequest.TransactionID))
				return
			}
		}

		// Update transaction status
		if err := repo.UpdateTransactionStatus(trx.ID, callbackRequest.Status); err != nil {
			sendAPIResponse(w, r, http.StatusInternalServerError, "", nil, dataFormat)
			log.Error("failed to update transaction status", logger.NewField("Error", err.Error()))
			return
		}

		// Additional actions based on status
		if strings.ToLower(callbackRequest.Status) == "failed" {
			err = repo.UpdateUserBalance(trx.UserID, trx.Amount)
			if err != nil {
				log.Error("failed to update user balance", logger.NewField("Error", err.Error()))
				sendAPIResponse(w, r, http.StatusInternalServerError, internalServerErrorMsg, nil, dataFormat)
				return
			}
		}

		sendAPIResponse(w, r, http.StatusOK, "Withdrawal transaction status updated successfully", nil, dataFormat)

		cacheKey := cache.ConstructTransactionIDKey(trx.ID)
		_ = dstrCache.Delete(cacheKey)

		// log to kafka
		encryptedTrx, err := services.MaskData(trx)
		if err != nil {
			log.Warn("error encrypting deposit transaction update", logger.NewField("Error", err.Error()), logger.NewField("Request-ID", requestID(r)))
			return
		}
		err = publisher.PublishTransaction(context.Background(), trx.ID, encryptedTrx, dataFormat)
		if err != nil {
			log.Warn("error publishing deposit transaction status update", logger.ComponentKafka,
				logger.NewField("Error", err.Error()), logger.NewField("Request-ID", requestID(r)), logger.NewField("transaction-ID", trx.ID))
			return
		}
		log.Info("Withdrawal transaction status update event published to Kafka", logger.ComponentKafka, logger.NewField("Request-ID", requestID(r)), logger.NewField("transaction-ID", trx.ID))

	}
}
