package payments

import (
	"os"
	"strconv"
	"time"

	"strings"

	creditcard "github.com/durango/go-credit-card"
	paypalsdk "github.com/ssherriff/PayPal-Go-SDK"
	"github.com/unirep/ur-local-web/app/config"
	"github.com/unirep/ur-local-web/app/models/user"
)

var appConfig = config.Config()

//CreditCard credit card details
type CreditCard struct {
	Number     string
	ExpiryDate string
	CVV        string
	Name       string
}

//CreditCardToken saved credit card reference
type CreditCardToken struct {
	ID   string
	Mask string
}

//CreditCardPaymentError error response
type CreditCardPaymentError struct {
	ErrorCode    string
	ErrorMessage string
}

// Error method implementation for ErrorResponse struct
func (r *CreditCardPaymentError) Error() string {
	return r.ErrorMessage
}

//StoreCreditCard save a credit card for later use to a tokenised system
func StoreCreditCard(card *CreditCard) (*CreditCardToken, error) {

	expiry := strings.Split(card.ExpiryDate, "/")
	number := strings.Replace(card.Number, " ", "", -1)
	expiryMonth := strings.TrimSpace(expiry[0])
	expiryYear := strings.TrimSpace(expiry[1])

	cardToValidate := creditcard.Card{Number: number, Cvv: card.CVV, Month: expiryMonth, Year: expiryYear}
	// pass bool to allow test credit cards, so if PaymentSandbox is true, allow for test credit cards
	err := cardToValidate.Validate(appConfig.PaymentSandbox)
	if err != nil {
		return nil, err
	}
	// retrieves details about card, allows us to use type in request to Paypal
	err = cardToValidate.Method()
	if err != nil {
		return nil, err
	}

	// current using the Paypal Vault for saving credit cards
	c, err := createPaypalClient()
	if err != nil {
		return nil, err
	}

	// Store CC
	storeYear := cardToValidate.Year
	if len(cardToValidate.Year) == 2 {
		storeYear = "20" + cardToValidate.Year
	}
	retCC, err := c.StoreCreditCard(paypalsdk.CreditCard{
		Number:      cardToValidate.Number,
		Type:        cardToValidate.Company.Short,
		ExpireMonth: cardToValidate.Month,
		ExpireYear:  storeYear,
		CVV2:        card.CVV,
		//FirstName:   "Foo",
		//LastName:    "Bar",
	})
	if err != nil {
		return nil, err
	}

	token := &CreditCardToken{
		ID:   retCC.ID,
		Mask: MaskCreditCard(card.Number),
	}
	return token, nil
}

//DeleteCreditCard save a credit card for later use to a tokenised system
func DeleteCreditCard(ID string) error {
	c, err := createPaypalClient()
	if err != nil {
		return err
	}

	err = c.DeleteCreditCard(ID)
	if err != nil {
		return err
	}

	return nil
}

//MakeCreditCardPayment make a credit card payment through paypal using stored credit card ID
func MakeCreditCardPayment(paymentRequest *user.Payment) error {
	c, err := createPaypalClient()
	if err != nil {
		return err
	}

	decAmount := float64(paymentRequest.Amount)

	decAmountStr := strconv.FormatFloat(decAmount, 'f', 2, 32)

	p := paypalsdk.Payment{
		Intent: "sale",
		Payer: &paypalsdk.Payer{
			PaymentMethod: "credit_card",
			FundingInstruments: []paypalsdk.FundingInstrument{paypalsdk.FundingInstrument{
				CreditCardToken: &paypalsdk.CreditCardToken{
					CreditCardID: paymentRequest.Booking.User.CreditCardID,
				},
			}},
		},
		Transactions: []paypalsdk.Transaction{paypalsdk.Transaction{
			Amount: &paypalsdk.Amount{
				Currency: "AUD",
				Total:    decAmountStr,
			},
			Description: "UR Local service payment",
		}},
	}

	response, err := c.CreatePayment(p)
	if err != nil {
		errorResponse := err.(*paypalsdk.ErrorResponse)
		ccError := &CreditCardPaymentError{
			ErrorCode:    errorResponse.Name,
			ErrorMessage: errorResponse.Error(),
		}
		return ccError
	}

	currDate := time.Now().UTC()
	paymentRequest.PaymentDate = currDate
	paymentRequest.ConfirmedDate = currDate
	paymentRequest.PaymentMethod = user.PaymentMethodCreditCard
	paymentRequest.PaymentStatus = response.State
	paymentRequest.TransactionID = response.ID
	paymentRequest.AcctDisplay = paymentRequest.Booking.User.CreditCardMask
	paymentRequest.Status = user.PaymentRequestStatusPaid

	return nil
}

//****** ONLY HELPER FUNCTIONS BELOW ******///

func createPaypalClient() (*paypalsdk.Client, error) {
	// Create a client instance
	apiBase := paypalsdk.APIBaseLive
	if appConfig.PaymentSandbox {
		apiBase = paypalsdk.APIBaseSandBox
	}
	c, err := paypalsdk.NewClient(appConfig.PayPalAPI.ClientID, appConfig.PayPalAPI.Secret, apiBase)
	if err != nil {
		return nil, err
	}
	c.SetLog(os.Stdout) // Set log to terminal stdout

	_, err = c.GetAccessToken()
	if err != nil {
		return nil, err
	}
	return c, err
}

//MaskCreditCard mask the credit card number for display
func MaskCreditCard(creditCardNumber string) string {
	first2 := creditCardNumber[0:2]
	last4 := creditCardNumber[len(creditCardNumber)-4:]
	mask := first2 + "xx xxxx xxxx " + last4
	return mask
}
