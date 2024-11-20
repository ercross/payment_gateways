package dto

type DataFormat int8

const (

	// DataFormatJSON is the default data format
	DataFormatJSON DataFormat = iota
	DataFormatXML
)

// WithdrawalRequest is a standard request structure for the transactions
type WithdrawalRequest struct {
	Amount             float64 `json:"amount" xml:"amount" validate:"required,gt=0"`
	UserID             int     `json:"user_id" xml:"user_id" validate:"required"`
	PaymentGatewayName string  `json:"payment_gateway_name" xml:"payment_gateway_name" validate:"required"`
	ReceivingAccount   string  `json:"receiving_account_id" xml:"receiving_account_id" validate:"required"`
	AuthenticationCode string  `json:"authentication_code" xml:"authentication_code" validate:"required"`
}

type DepositRequest struct {
	Amount   float64 `json:"amount" xml:"amount" validate:"required,gt=0"`
	UserID   int     `json:"user_id" xml:"user_id" validate:"required"`
	Currency string  `json:"currency" xml:"currency" validate:"required"`
}

// APIResponse is a standard response structure for the APIs
type APIResponse struct {
	StatusCode int         `json:"status_code" xml:"status_code"`
	Message    string      `json:"message" xml:"message"`
	Data       interface{} `json:"data,omitempty" xml:"data,omitempty"`
}

type TransactionStatusCallback struct {
	TransactionID int    `json:"transaction_id" xml:"transaction_id" validate:"required"`
	Status        string `json:"status" xml:"status" validate:"required"`
}

func (w *WithdrawalRequest) IsDecodable() bool {
	return true
}

func (w *DepositRequest) IsDecodable() bool {
	return true
}

func (w *TransactionStatusCallback) IsDecodable() bool {
	return true
}
