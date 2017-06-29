package payments

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/unirep/ur-local-web/app/config"
)

//PaypalResponse object representation of the response object returned from Paypal
type PaypalResponse struct {
	Timestamp     string
	CorrelationID string
	Ack           string
	Version       string
	Build         string
	AVSCode       string
	CVV2Match     string
	TransactionID string
	Errors        []PaypalError
}

//PaypalError representation of a Paypal error
type PaypalError struct {
	Code         string
	ShortMessage string
	LongMessage  string
	SeverityCode string
	ParamID      string
	ParamValue   string
}

//SetExpressCheckoutRequest fields for Paypal SetExpressCheckout request
type SetExpressCheckoutRequest struct {
	SuccessPath string
	CancelPath  string
	Total       float64
	Description string
}

//SetExpressCheckoutResponse fields in the response from SetExpressCheckout
type SetExpressCheckoutResponse struct {
	Timestamp     string
	CorrelationID string
	Ack           string
	Token         string
	RedirectURL   string
	Errors        []PaypalError
}

//GetExpressCheckoutDetailsResponse fields for Paypal GetExpressCheckoutDetails response
type GetExpressCheckoutDetailsResponse struct {
	Timestamp      string
	CorrelationID  string
	Ack            string
	Token          string
	CheckoutStatus string
	PayerID        string
	PayerStatus    string
	PaypalCountry  string
	Email          string
	Phone          string
	BusinessName   string
	Fname          string
	Lname          string
	ShipAddress    string
	ShipCity       string
	ShipState      string
	ShipZip        string
	ShipCountry    string
	Errors         []PaypalError
}

//DoExpressCheckoutPaymentResponse fields for Paypal DoExpressCheckoutPayment response
type DoExpressCheckoutPaymentResponse struct {
	Timestamp      string
	CorrelationID  string
	Ack            string
	TransactionID  string
	CheckoutStatus string
	Errors         []PaypalError
}

//ParsePaypalResponse method to parse the returned paypal response
func ParsePaypalResponse(response string) (PaypalResponse, error) {
	paypalResponse := PaypalResponse{}

	responseValues, err := url.ParseQuery(response)

	if err != nil {
		return paypalResponse, err
	}

	paypalResponse.Timestamp = responseValues.Get("TIMESTAMP")
	paypalResponse.CorrelationID = responseValues.Get("CORRELATIONID")
	paypalResponse.Ack = responseValues.Get("ACK")
	paypalResponse.Version = responseValues.Get("VERSION")
	paypalResponse.Build = responseValues.Get("BUILD")
	paypalResponse.AVSCode = responseValues.Get("AVSCODE")
	paypalResponse.CVV2Match = responseValues.Get("CVV2MATCH")
	paypalResponse.TransactionID = responseValues.Get("TRANSACTIONID")

	for k := range responseValues {
		switch {

		case strings.HasPrefix(k, "L_ERRORCODE"):
			index, _ := strconv.ParseInt(strings.Replace(k, "L_ERRORCODE", "", 1), 10, 32)
			GetPaypalErrorInstance(&paypalResponse, int(index)).Code = responseValues.Get(k)
		case strings.HasPrefix(k, "L_SHORTMESSAGE"):
			index, _ := strconv.ParseInt(strings.Replace(k, "L_SHORTMESSAGE", "", 1), 10, 32)
			GetPaypalErrorInstance(&paypalResponse, int(index)).ShortMessage = responseValues.Get(k)
		case strings.HasPrefix(k, "L_LONGMESSAGE"):
			index, _ := strconv.ParseInt(strings.Replace(k, "L_LONGMESSAGE", "", 1), 10, 32)
			GetPaypalErrorInstance(&paypalResponse, int(index)).LongMessage = responseValues.Get(k)
		case strings.HasPrefix(k, "L_SEVERITYCODE"):
			index, _ := strconv.ParseInt(strings.Replace(k, "L_SEVERITYCODE", "", 1), 10, 32)
			GetPaypalErrorInstance(&paypalResponse, int(index)).SeverityCode = responseValues.Get(k)
		case strings.HasPrefix(k, "L_ERRORPARAMID"):
			index, _ := strconv.ParseInt(strings.Replace(k, "L_ERRORPARAMID", "", 1), 10, 32)
			GetPaypalErrorInstance(&paypalResponse, int(index)).ParamID = responseValues.Get(k)
		case strings.HasPrefix(k, "L_ERRORPARAMVALUE"):
			index, _ := strconv.ParseInt(strings.Replace(k, "L_ERRORPARAMVALUE", "", 1), 10, 32)
			GetPaypalErrorInstance(&paypalResponse, int(index)).ParamValue = responseValues.Get(k)
		}
	}

	return paypalResponse, nil
}

//GetPaypalErrorInstance get specific error instance in the Paypal response
func GetPaypalErrorInstance(response *PaypalResponse, index int) *PaypalError {
	if len(response.Errors) < (index + 1) {
		response.Errors = append(response.Errors, PaypalError{})
	}
	return &response.Errors[index]
}

//SetExpressCheckout do the Paypal request of SetExpressCheckout
func SetExpressCheckout(request *SetExpressCheckoutRequest) (*SetExpressCheckoutResponse, error) {
	var signature, username, password, paypalURL, paypalRedirectURL, cancelURL, returnURL, description string

	retVal := &SetExpressCheckoutResponse{}

	signature = appConfig.PayPalAPI.Signature
	username = appConfig.PayPalAPI.Username
	password = appConfig.PayPalAPI.Password

	paypalURL = appConfig.PayPalAPI.URL
	paypalRedirectURL = appConfig.PayPalAPI.RedirectURL

	cancelURL = fmt.Sprintf("%s%s", config.FullBaseURL(), request.CancelPath)
	returnURL = fmt.Sprintf("%s%s", config.FullBaseURL(), request.SuccessPath)

	if len(request.Description) > 126 {
		description = request.Description[:126]
	} else {
		description = request.Description
	}

	transaction := fmt.Sprintf("METHOD=SetExpressCheckout&VERSION=95.0&SIGNATURE=%s&USER=%s&PWD=%s&PAYMENTREQUEST_0_AMT=%s&PAYMENTREQUEST_0_CURRENCYCODE=AUD&PAYMENTREQUEST_0_DESC=%s&CANCELURL=%s&RETURNURL=%s&PAYMENTREQUEST_0_PAYMENTACTION=Sale",
		signature, username, password, fmt.Sprintf("$%.2f", request.Total), url.QueryEscape(description), url.QueryEscape(cancelURL), url.QueryEscape(returnURL))

	log.Print("transaction string:", transaction)

	// Build the request
	req, err := http.NewRequest("POST", "https://"+paypalURL, bytes.NewBuffer([]byte(transaction)))

	if err != nil {
		log.Println("Error building paypal request: ", err.Error())
		return retVal, err
	}

	req.ContentLength = int64(len(transaction))

	// Send the request via a client
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Println("Error sending paypal request: ", err.Error())
		return retVal, err
	}
	// Defer the closing of the body
	defer resp.Body.Close()

	log.Println("response Status:", resp.Status)
	log.Println("response Headers:", resp.Header)
	body, _ := ioutil.ReadAll(resp.Body)
	response, _ := url.QueryUnescape(string(body))
	log.Println("response Body:", response)

	responseValues, err := url.ParseQuery(response)
	if err != nil {
		log.Println("Error parsing paypal express response: ", err.Error())
		return retVal, err
	}

	ack := responseValues.Get("ACK")
	token := responseValues.Get("TOKEN")

	retVal.Token = token
	retVal.Ack = ack
	retVal.CorrelationID = responseValues.Get("CORRELATIONID")
	retVal.Timestamp = responseValues.Get("TIMESTAMP")

	if ack == "Success" {
		paypalRedirectURLFull := fmt.Sprintf("https://%s?cmd=_express-checkout&token=%s", paypalRedirectURL, token)
		retVal.RedirectURL = paypalRedirectURLFull

		return retVal, nil
	}

	log.Println("Error returned from set express checkout")
	// need to set paypal errors

	return retVal, errors.New("Paypal error")

}

//GetExpressCheckoutDetails retrieve the details from the GetExpressCheckout request
func GetExpressCheckoutDetails(token string) (*GetExpressCheckoutDetailsResponse, error) {
	var signature, username, password, paypalURL string

	retVal := &GetExpressCheckoutDetailsResponse{}

	signature = appConfig.PayPalAPI.Signature
	username = appConfig.PayPalAPI.Username
	password = appConfig.PayPalAPI.Password

	paypalURL = appConfig.PayPalAPI.URL

	transaction := fmt.Sprintf("METHOD=GetExpressCheckoutDetails&VERSION=95.0&SIGNATURE=%s&USER=%s&PWD=%s&TOKEN=%s",
		signature, username, password, token)

	log.Print("transaction string:", transaction)

	// Build the request
	req, err := http.NewRequest("POST", "https://"+paypalURL, bytes.NewBuffer([]byte(transaction)))

	if err != nil {
		log.Println("Error building paypal request: ", err.Error())
		return retVal, err
	}

	req.ContentLength = int64(len(transaction))

	// Send the request via a client
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Println("Error sending paypal request: ", err.Error())
		return retVal, err
	}
	// Defer the closing of the body
	defer resp.Body.Close()

	log.Println("response Status:", resp.Status)
	log.Println("response Headers:", resp.Header)
	body, _ := ioutil.ReadAll(resp.Body)
	response, _ := url.QueryUnescape(string(body))
	log.Println("response Body:", response)

	responseValues, err := url.ParseQuery(response)
	if err != nil {
		log.Println("Error parsing paypal express response: ", err.Error())
		return retVal, err
	}

	ack := responseValues.Get("ACK")

	retVal.Ack = ack
	retVal.CorrelationID = responseValues.Get("CORRELATIONID")
	retVal.Timestamp = responseValues.Get("TIMESTAMP")

	if ack == "Success" {
		retVal.Token = responseValues.Get("TOKEN")
		retVal.CheckoutStatus = responseValues.Get("CHECKOUTSTATUS")
		retVal.Email = responseValues.Get("EMAIL")
		retVal.Phone = responseValues.Get("PHONENUM")
		retVal.PayerID = responseValues.Get("PAYERID")
		retVal.PayerStatus = responseValues.Get("PAYERSTATUS")
		retVal.BusinessName = responseValues.Get("BUSINESS")
		retVal.Fname = responseValues.Get("FIRSTNAME")
		retVal.Lname = responseValues.Get("LASTNAME")
		retVal.PaypalCountry = responseValues.Get("COUNTRYCODE")
		retVal.ShipAddress = responseValues.Get("PAYMENTREQUEST_0_SHIPTOSTREET")
		retVal.ShipCity = responseValues.Get("PAYMENTREQUEST_0_SHIPTOCITY")
		retVal.ShipState = responseValues.Get("PAYMENTREQUEST_0_SHIPTOSTATE")
		retVal.ShipZip = responseValues.Get("PAYMENTREQUEST_0_SHIPTOZIP")
		retVal.ShipCountry = responseValues.Get("PAYMENTREQUEST_0_SHIPTOCOUNTRYCODE")

		return retVal, err
	}

	log.Println("Error returned from get express checkout details")
	return retVal, errors.New("Paypal error")
}

//DoExpressCheckoutPayment make the request to Paypal for DoExpressCheckout
func DoExpressCheckoutPayment(token string, payerID string, total float64) (*DoExpressCheckoutPaymentResponse, error) {
	var signature, username, password, paypalURL string

	retVal := &DoExpressCheckoutPaymentResponse{}

	signature = appConfig.PayPalAPI.Signature
	username = appConfig.PayPalAPI.Username
	password = appConfig.PayPalAPI.Password

	paypalURL = appConfig.PayPalAPI.URL

	transaction := fmt.Sprintf("METHOD=DoExpressCheckoutPayment&VERSION=95.0&SIGNATURE=%s&USER=%s&PWD=%s&TOKEN=%s&PAYERID=%s&PAYMENTREQUEST_0_AMT=%s&PAYMENTREQUEST_0_CURRENCYCODE=AUD&PAYMENTREQUEST_0_PAYMENTACTION=Sale",
		signature, username, password, token, payerID, fmt.Sprintf("$%.2f", total))

	log.Print("transaction string:", transaction)

	// Build the request
	req, err := http.NewRequest("POST", "https://"+paypalURL, bytes.NewBuffer([]byte(transaction)))

	if err != nil {
		log.Println("Error building paypal request: ", err.Error())
		return retVal, err
	}

	req.ContentLength = int64(len(transaction))

	// Send the request via a client
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Println("Error sending paypal request: ", err.Error())
		return retVal, err
	}
	// Defer the closing of the body
	defer resp.Body.Close()

	log.Println("response Status:", resp.Status)
	log.Println("response Headers:", resp.Header)
	body, _ := ioutil.ReadAll(resp.Body)
	response, _ := url.QueryUnescape(string(body))
	log.Println("response Body:", response)

	responseValues, err := url.ParseQuery(response)
	if err != nil {
		log.Println("Error parsing paypal express response: ", err.Error())
		return retVal, err
	}

	ack := responseValues.Get("ACK")
	retVal.Ack = ack

	if ack == "Success" {
		retVal.CheckoutStatus = responseValues.Get("PAYMENTINFO_0_PAYMENTSTATUS")
		retVal.TransactionID = responseValues.Get("PAYMENTINFO_0_TRANSACTIONID")
		return retVal, err
	}

	return retVal, errors.New("Error submitting Paypal Payment - " + responseValues.Get("L_ERRORCODE0") + ": " + responseValues.Get("L_LONGMESSAGE0"))
}
