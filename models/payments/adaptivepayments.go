package payments

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

const (
	AdaptivePaymentsActionTypePay        = "PAY"
	AdaptivePaymentsActionTypeCreate     = "CREATE"
	AdaptivePaymentsActionTypePayPrimary = "PAY_PRIMARY"

	AdaptivePaymentsFeesPayerSender          = "SENDER"
	AdaptivePaymentsFeesPayerPrimaryReceiver = "PRIMARYRECEIVER"
	AdaptivePaymentsFeesPayerEachReceiver    = "EACHRECEIVER"
	AdaptivePaymentsFeesPayerSecondaryOnly   = "SECONDARYONLY"
)

type APRequestEnvelope struct {
	ErrorLanguage string `json:"errorLanguage"`
}

type APResponseEnvelope struct {
	Ack           string    `json:"ack"`
	Build         string    `json:"build"`
	CorrelationID string    `json:"correlationId"`
	Timestamp     time.Time `json:"timestamp"`
}

type APPhoneNumber struct {
	CountryCode string `json:"countryCode"`
	PhoneNumber string `json:"phoneNumber"`
	Extension   string `json:"extension"`
}

type APReceiver struct {
	Amount    float64        `json:"amount"`
	Email     *string        `json:"email,omitempty"`
	Phone     *APPhoneNumber `json:"phoneNumber,omitempty"`
	AccountID *string        `json:"accountId,omitempty"`
	InvoiceID *string        `json:"invoiceId,omitempty"`
	Primary   bool           `json:"primary"`
}

type APReceiverList struct {
	Receiver []*APReceiver `json:"receiver"`
}

type APPaymentRequest struct {
	ActionType         string             `json:"actionType"`
	CancelURL          string             `json:"cancelUrl"`
	ReturnURL          string             `json:"returnUrl"`
	CurrencyCode       string             `json:"currencyCode"`
	FeesPayer          *string            `json:"feesPayer,omitempty"`
	IPNNotificationURL *string            `json:"ipnNotificationUrl,omitempty"`
	Memo               *string            `json:"memo,omitempty"`
	PayKeyDuration     *string            `json:"payKeyDuration,omitempty"`
	PIN                *string            `json:"pin,omitempty"`
	PreapprovalKey     *string            `json:"preapprovalKey,omitempty"`
	TrackingID         *string            `json:"trackingId,omitempty"`
	ReceiverList       *APReceiverList    `json:"receiverList"`
	RequestEnvelope    *APRequestEnvelope `json:"requestEnvelope"`
}

type APPaymentInfoList struct {
	PaymentInfo []*APPaymentInfo `json:"paymentInfo"`
}

type APPaymentInfo struct {
	PendingReason           string      `json:"pendingReason"`
	PendingRefund           bool        `json:"pendingRefund"`
	Receiver                *APReceiver `json:"receiver"`
	RefundedAmount          string      `json:"refundedAmount"`
	SenderTransactionID     string      `json:"senderTransactionId"`
	SenderTransactionStatus string      `json:"senderTransactionStatus"`
	TransactionID           string      `json:"transactionId"`
	TransactionStatus       string      `json:"transactionStatus"`
}

type APWarningData struct {
	WarningID int64
	Message   string
}

type APErrorData struct {
	ErrorID  string `json:"errorId"`
	Category string `json:"category"`
	Domain   string `json:"domain"`
	Message  string `json:"message"`
	Severity string `json:"severity"`
}

type APPaymentResponse struct {
	PayKey            string              `json:"payKey"`
	PayErrorList      string              `json:"payErrorList"`
	PaymentExecStatus string              `json:"payExecStatus"`
	PaymentInfoList   []*APPaymentInfo    `json:"paymentInfoList"`
	ResponseEnvelope  *APResponseEnvelope `json:"responseEnvelope"`
	WarningDataList   []*APWarningData    `json:"warningDataList"`
	Error             []*APErrorData      `json:"error"`
}

type APPaymentDetailsRequest struct {
	PayKey          string             `json:"payKey"`
	RequestEnvelope *APRequestEnvelope `json:"requestEnvelope"`
	TransactionID   *string            `json:"transactionId,omitempty"`
	TrackingID      *string            `json:"trackingId,omitempty"`
}

type APPaymentDetailsResponse struct {
	ActionType           string              `json:"actionType"`
	CancelURL            string              `json:"cancelUrl"`
	ReturnURL            string              `json:"returnUrl"`
	CurrencyCode         string              `json:"currencyCode"`
	FeesPayer            string              `json:"feesPayer,omitempty"`
	IPNNotificationURL   string              `json:"ipnNotificationUrl,omitempty"`
	Memo                 string              `json:"memo,omitempty"`
	PayKey               string              `json:"payKey"`
	PayKeyExpirationDate time.Time           `json:"payKeyExpirationDate"`
	PreapprovalKey       string              `json:"preapprovalKey,omitempty"`
	ResponseEnvelope     *APResponseEnvelope `json:"responseEnvelope"`
	Status               string              `json:"status"`
	TrackingID           string              `json:"trackingId"`
}

//AdaptivePaymentsPay Pay method for Paypal Adaptive Payments
func AdaptivePaymentsPay(paymentRequest *APPaymentRequest) (*APPaymentResponse, error) {

	var requestData []byte
	var err error

	retVal := &APPaymentResponse{}

	requestEnvelope := &APRequestEnvelope{
		ErrorLanguage: "en_US",
	}
	paymentRequest.RequestEnvelope = requestEnvelope

	if requestData, err = json.Marshal(paymentRequest); err != nil {
		log.Println("Error in rmodels/payments/adaptivepayments.go AdaptivePaymentsPay(paymentRequest *APPaymentRequest): Error marsheling json for paypal adaptive payments request: ", err.Error())
		return nil, err
	}
	log.Println("DEBUG: requestData: ", string(requestData))

	body, err := sendRequest("Pay", requestData)
	if err != nil {
		log.Println("Error in rmodels/payments/adaptivepayments.go AdaptivePaymentsPay(paymentRequest *APPaymentRequest): Error sending paypal adaptive payments request: ", err.Error())
		return nil, err
	}

	err = json.Unmarshal(body, retVal)
	if err != nil {
		log.Println("Error in rmodels/payments/adaptivepayments.go AdaptivePaymentsPay(paymentRequest *APPaymentRequest): Error unmarshelling returned paypal adaptive payments response: ", err.Error())
		return nil, err
	}

	if retVal.ResponseEnvelope.Ack == "Failure" {
		return retVal, errors.New("Errors returned from Paypal request")
	}

	return retVal, nil
}

//AdaptivePaymentsGetPaymentDetails PaymentDetails method for Paypal Adaptive Payments
func AdaptivePaymentsGetPaymentDetails(detailsRequest *APPaymentDetailsRequest) (*APPaymentDetailsResponse, error) {

	var requestData []byte
	var err error

	retVal := &APPaymentDetailsResponse{}

	requestEnvelope := &APRequestEnvelope{
		ErrorLanguage: "en_US",
	}
	detailsRequest.RequestEnvelope = requestEnvelope

	if requestData, err = json.Marshal(detailsRequest); err != nil {
		log.Println("Error in rmodels/payments/adaptivepayments.go AdaptivePaymentsGetPaymentDetails(detailsRequest *APPaymentDetailsRequest): Error marsheling json for paypal adaptive payments detail request: ", err.Error())
		return nil, err
	}
	log.Println("requestData: ", string(requestData))

	body, err := sendRequest("PaymentDetails", requestData)
	if err != nil {
		log.Println("Error in rmodels/payments/adaptivepayments.go AdaptivePaymentsGetPaymentDetails(detailsRequest *APPaymentDetailsRequest): Error sending paypal adaptive payments details request: ", err.Error())
		return nil, err
	}

	err = json.Unmarshal(body, retVal)
	if err != nil {
		log.Println("Error in rmodels/payments/adaptivepayments.go AdaptivePaymentsGetPaymentDetails(detailsRequest *APPaymentDetailsRequest): Error unmarshelling returned paypal adaptive payments response: ", err.Error())
		return nil, err
	}

	if retVal.ResponseEnvelope.Ack == "Failure" {
		return retVal, errors.New("Errors returned from Paypal request")
	}

	return retVal, nil
}

func sendRequest(api string, requestData []byte) ([]byte, error) {
	paypalURL := appConfig.PayPalAPI.URL
	userID := appConfig.PayPalAPI.Username
	password := appConfig.PayPalAPI.Password
	signature := appConfig.PayPalAPI.Signature
	appID := appConfig.PayPalAPI.AppID

	req, err := http.NewRequest("POST", "https://"+paypalURL+"/"+api, bytes.NewBuffer(requestData))
	if err != nil {
		log.Println("Error in rmodels/payments/adaptivepayments.go endRequest(api string, requestData []byte): Error building paypal adaptive payments request: ", err.Error())
		return nil, err
	}

	req.ContentLength = int64(len(requestData))
	req.Header.Add("X-PAYPAL-SECURITY-USERID", userID)
	req.Header.Add("X-PAYPAL-SECURITY-PASSWORD", password)
	req.Header.Add("X-PAYPAL-SECURITY-SIGNATURE", signature)
	req.Header.Add("X-PAYPAL-REQUEST-DATA-FORMAT", "JSON")
	req.Header.Add("X-PAYPAL-RESPONSE-DATA-FORMAT", "JSON")
	req.Header.Add("X-PAYPAL-APPLICATION-ID", appID)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Println("Error in rmodels/payments/adaptivepayments.go endRequest(api string, requestData []byte): Error sending paypal adaptive payments request: ", err.Error())
		return nil, err
	}
	// Defer the closing of the body
	defer resp.Body.Close()

	log.Println("DEBUG: response Status:", resp.Status)
	log.Println("DEBUG: response Headers:", resp.Header)
	body, _ := ioutil.ReadAll(resp.Body)
	log.Println("DEBUG: response Body:", string(body))

	if resp.StatusCode == 500 {
		return nil, errors.New("Status 500 returned from server: " + string(body))
	}

	return body, nil
}
