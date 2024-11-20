package v1

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"github.com/ercross/payment_gateways/internal/api/v1/dto"
	"github.com/go-chi/chi/v5/middleware"
	"net/http"
)

// sendAPIResponse sends a response in JSON or XML format based on the Content-Type header.
func sendAPIResponse(w http.ResponseWriter, _ *http.Request, statusCode int, message string, data interface{}, dataType dto.DataFormat) {
	// Create the response object
	response := dto.APIResponse{
		StatusCode: statusCode,
		Message:    message,
		Data:       data,
	}

	switch dataType {
	case dto.DataFormatXML:
		w.Header().Set("Content-Type", "application/xml")
		w.WriteHeader(statusCode)
		if err := xml.NewEncoder(w).Encode(response); err != nil {
			http.Error(w, "Error encoding XML response", http.StatusInternalServerError)
		}
	default: // Default to JSON
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(statusCode)
		if err := json.NewEncoder(w).Encode(response); err != nil {
			http.Error(w, "Error encoding JSON response", http.StatusInternalServerError)
		}
	}
}

func constructDepositCallbackUrl(baseURL string, trxID int) string {
	return fmt.Sprintf("%s/%s/%d", baseURL, depositCallbackPath, trxID)
}

func constructWithdrawalCallbackUrl(baseURL string, trxID int) string {
	return fmt.Sprintf("%s/%s/%d", baseURL, withdrawalCallbackPath, trxID)
}

func constructDepositLockKey(amount float64, userID int, currency string, paymentGatewayName string) string {
	return fmt.Sprintf("deposit-trx_%f_%d_%s_%s", amount, userID, currency, paymentGatewayName)
}

func requestID(r *http.Request) string {
	return middleware.GetReqID(r.Context())
}
