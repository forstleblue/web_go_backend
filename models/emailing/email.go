package emailing

import (
	"bytes"
	"fmt"
	"log"

	"github.com/gocql/gocql"
	"github.com/unirep/ur-local-web/app/config"
	"github.com/unirep/ur-local-web/app/models/platform"
	"github.com/unirep/ur-local-web/app/models/user"

	"github.com/CloudyKit/jet"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ses"
	"github.com/unirep/ur-local-web/app/models/feedback"
	"github.com/unirep/ur-local-web/app/models/message"
	"github.com/unirep/ur-local-web/app/models/notification"
)

var emailConfig *config.EmailSettings

func Configure(config *config.EmailSettings) {
	emailConfig = config
}

// SendRegistrationEmail Registration Welcome Email
func SendRegistrationEmail(u *user.User) {
	fromEmail := emailConfig.FromEmail
	subject := "Welcome to ur-local.com"
	dashboardURL := config.FullBaseURL() + "/dashboard"

	vars := make(jet.VarMap)
	vars.Set("User", u)
	vars.Set("DashboardURL", dashboardURL)
	html, err := generateEmail("registration.html", vars)
	if err != nil {
		fmt.Println("failed prepare html template", err)
	}
	text, err := generateEmail("registration.txt", vars)
	if err != nil {
		fmt.Println("failed prepare text template", err)
	}

	err = sendEmail(u.Email, fromEmail, subject, text, html)

	if err != nil {
		log.Printf("Failed to send registration welcome email: %s\n", err.Error())
		// should probably mark the account with a notice that welcome email hasn't been sent
		// then, either resend later, or manually look at admin and resend
	}
}

// SendRegistrationViaPlatformEmail Registration via platform Welcome Email
func SendRegistrationViaPlatformEmail(u *user.User, p *user.Profile, plat *platform.Platform, urLocalURL string) {
	fromEmail := emailConfig.FromEmail
	subject := "Welcome to ur-local.com"
	dashboardURL := urLocalURL + "/dashboard"

	vars := make(jet.VarMap)
	vars.Set("User", u)
	vars.Set("Profile", p)
	vars.Set("Platform", plat)
	vars.Set("DashboardURL", dashboardURL)
	html, err := generateEmail("registrationviaplatform.html", vars)
	if err != nil {
		fmt.Println("failed prepare html template", err)
	}
	text, err := generateEmail("registrationviaplatform.txt", vars)
	if err != nil {
		fmt.Println("failed prepare text template", err)
	}

	err = sendEmail(u.Email, fromEmail, subject, text, html)

	if err != nil {
		log.Printf("Failed to send registration welcome email: %s\n", err.Error())
		// should probably mark the account with a notice that welcome email hasn't been sent
		// then, either resend later, or manually look at admin and resend
	}
}

//SendNewPlatformProfileEmail Platform Welcome Email
func SendNewPlatformProfileEmail(p *user.Profile, plat *platform.Platform, urLocalURL string) {
	fromEmail := emailConfig.FromEmail
	subject := "New profile added"
	dashboardURL := urLocalURL + "/dashboard"

	vars := make(jet.VarMap)
	vars.Set("User", p.User)
	vars.Set("Profile", p)
	vars.Set("Platform", plat)
	vars.Set("DashboardURL", dashboardURL)
	html, err := generateEmail("new_platform_profile.html", vars)
	if err != nil {
		fmt.Println("failed prepare html template", err)
	}
	text, err := generateEmail("new_platform_profile.txt", vars)
	if err != nil {
		fmt.Println("failed prepare text template", err)
	}

	err = sendEmail(p.User.Email, fromEmail, subject, text, html)

	if err != nil {
		log.Printf("Failed to send new platform profile email: %s\n", err.Error())
	}
}

// SendNewPasswordResetEmail is invoked when reset password
func SendNewPasswordResetEmail(uReceiver *user.User, token gocql.UUID) {
	fromEmail := emailConfig.FromEmail
	subject := "UR-Local password reset was requested"
	dashboardURL := config.FullBaseURL() + "/dashboard#s1"
	resetPasswordURL := config.FullBaseURL() + "/reset-password?token=" + token.String()
	emailImageURL := config.FullBaseURL() + "/images/logo-email.png"
	findPersonURL := config.FullBaseURL() + "/find-person"

	vars := make(jet.VarMap)

	vars.Set("ReceiveUser", uReceiver)
	vars.Set("ResetPasswordUrl", resetPasswordURL)
	vars.Set("DashboardURL", dashboardURL)
	vars.Set("ImageURL", emailImageURL)
	vars.Set("FindPersonURL", findPersonURL)

	html, err := generateEmail("email_reset_password.html", vars)
	if err != nil {
		fmt.Println("failed prepare html template", err)
	}
	text, err := generateEmail("email_reset_password.txt", vars)
	if err != nil {
		fmt.Println("failed prepare text template", err)
	}

	err = sendEmail(uReceiver.Email, fromEmail, subject, text, html)

	if err != nil {
		log.Printf("Failed to send New Password email: %s\n", err.Error())
		// should probably mark the account with a notice that welcome email hasn't been sent
		// then, either resend later, or manually look at admin and resend
	}
}

// SendNewMessageEmail
func SendNewMessageEmail(uReceive *user.User, uSend *user.User, messageObject *message.Message) {
	fromEmail := emailConfig.FromEmail
	subject := "New UR-Local message!"
	dashboardURL := config.FullBaseURL() + "/dashboard#s1"
	emailImageURL := config.FullBaseURL() + "/images/logo-email.png"
	findPersonURL := config.FullBaseURL() + "/find-person"

	vars := make(jet.VarMap)
	vars.Set("ReceiveUser", uReceive)
	vars.Set("DashboardURL", dashboardURL)
	vars.Set("SendUser", uSend)
	vars.Set("Message", messageObject)
	vars.Set("ImageURL", emailImageURL)
	vars.Set("FindPersonURL", findPersonURL)
	html, err := generateEmail("email_message.html", vars)
	if err != nil {
		fmt.Println("failed prepare html template", err)
	}
	text, err := generateEmail("email_message.txt", vars)
	if err != nil {
		fmt.Println("failed prepare text template", err)
	}

	err = sendEmail(uReceive.Email, fromEmail, subject, text, html)

	if err != nil {
		log.Printf("Failed to send New Message email: %s\n", err.Error())
		// should probably mark the account with a notice that welcome email hasn't been sent
		// then, either resend later, or manually look at admin and resend
	}
}

//SendNewNotificationEmail sends notification email
func SendNewNotificationEmail(uReceive *user.User, uSend *user.User, notificationObject *notification.Notification, subjectName string, templateFileName string, templateFileNameText string) {
	fromEmail := emailConfig.FromEmail
	subject := subjectName
	baseURL := config.FullBaseURL()
	dashboardURL := baseURL + "/dashboard#s2"
	findPersonURL := baseURL + "/find-person"
	emailImageURL := baseURL + "/images/logo-email.png"

	var bookingInfo *user.Booking
	bookingInfo = &user.Booking{}
	var err error
	if notificationObject.NotificationType == notification.NotificationTypeBookingRequest ||
		notificationObject.NotificationType == notification.NotificationTypeBookingResponse ||
		notificationObject.NotificationType == notification.NotificationTypeLeaveFeedback ||
		notificationObject.NotificationType == notification.NotificationTypeFeedbackReceived {
		bookingInfo, err = user.GetBookingWithBookingId(notificationObject.EntityID)
		if err != nil {
			log.Println("Error in message/message.go SendNewNotificationEmail() fail to get booking information:", err)
		}
	} else if notificationObject.NotificationType == notification.NotificationTypePaymentRequest ||
		notificationObject.NotificationType == notification.NotificationTypePaymentRequestResponse ||
		notificationObject.NotificationType == notification.NotificationTypePaymentRequestMessage ||
		notificationObject.NotificationType == notification.NotificationTypePaid {
		pr, _ := user.GetPayment(notificationObject.EntityID)
		bookingInfo = &pr.Booking
		if err != nil {
			log.Println("Error in message/message.go SendNewNotificationEmail() fail to get booking information:", err)
		}
	}

	var fromDate, fromTime, message string
	var duration int
	var cost int32
	bookingHistory := &user.BookingHistory{}
	bookingHistory, err = user.GetBookingHistoryWithHistoryID(notificationObject.EntityHistoryID)
	if err != nil {
		log.Println("Error in message/message.go fail to get booking history information:", err)
		return
	}
	fromDate = bookingHistory.FromDate
	fromTime = bookingHistory.FromTime
	duration = bookingHistory.GetBookingDuration()
	message = bookingHistory.Message
	if bookingHistory.Fee.IntPart() == 0 {
		cost = int32(bookingHistory.TotalPrice.IntPart())
	} else {
		cost = int32(bookingHistory.Fee.IntPart())
	}
	profileServiceName := bookingInfo.Profile.GetProfileServiceName(bookingInfo.Profile.ServiceCategory)

	vars := make(jet.VarMap)
	vars.Set("ReceiveUser", uReceive)
	vars.Set("SendUser", uSend)
	vars.Set("ProfileServiceName", profileServiceName)
	vars.Set("DashboardURL", dashboardURL)
	vars.Set("BookingID", bookingInfo.BookingID)
	vars.Set("Date", fromDate)
	vars.Set("StartTime", fromTime)
	vars.Set("Duration", duration)
	vars.Set("Cost", cost)
	vars.Set("Comments", message)
	vars.Set("FindPersonURL", findPersonURL)
	vars.Set("ImageURL", emailImageURL)
	vars.Set("BaseURL", baseURL)
	vars.Set("BookingUUID", bookingInfo.BookingUUID)
	vars.Set("BookingHistoryUUID", bookingHistory.BookingHistoryUUID)
	vars.Set("IsDurationChanged", bookingHistory.DurationChanged())
	vars.Set("IsTotalCostChanged", bookingHistory.FeeChanged())
	IsDateChanged, IsStartTimeChanged := bookingHistory.DateAndStartTimeChanged()
	vars.Set("IsDateChanged", IsDateChanged)
	vars.Set("IsStartTimeChanged", IsStartTimeChanged)
	vars.Set("ServiceProviderName", bookingInfo.Profile.User.FullName())
	if notificationObject.NotificationType == notification.NotificationTypeFeedbackReceived && notificationObject.TwoFeedbackCompleted() == true {
		var feedbackInfo *feedback.Feedback
		if notificationObject.SenderID == bookingInfo.User.UserID {
			bookingInfo, _ := user.GetBookingWithBookingId(bookingInfo.BookingID)
			bookingSenderProfiles, _ := user.GetProfileByUserID(bookingInfo.User.UserID)
			bookingSenderProfile := bookingSenderProfiles[0]
			feedbackInfo, _ = feedback.GetFeedbackByBookingIDandCreatedProfileID(bookingInfo.BookingID, bookingSenderProfile.ProfileID)
		} else if notificationObject.SenderID == bookingInfo.Profile.User.UserID {
			feedbackInfo, _ = feedback.GetFeedbackByBookingIDandCreatedProfileID(bookingInfo.BookingID, bookingInfo.Profile.ProfileID)
		}

		vars.Set("Rating", feedbackInfo.Score/20)
		vars.Set("CommentTitle", feedbackInfo.Comment)
		var sdaString string
		for _, element := range feedbackInfo.SdaText {
			sdaString = sdaString + element + " "
		}

		vars.Set("SDA", sdaString)
		vars.Set("Comments", feedbackInfo.Description)
	}

	html, err := generateEmail(templateFileName, vars)

	if err != nil {
		fmt.Println("failed prepare html template", err)
	}
	text, err := generateEmail(templateFileNameText, vars)
	if err != nil {
		fmt.Println("failed prepare text template", err)
	}

	err = sendEmail(uReceive.Email, fromEmail, subject, text, html)

	if err != nil {
		log.Printf("Failed to send New Notification email: %s\n", err.Error())
		// should probably mark the account with a notice that welcome email hasn't been sent
		// then, either resend later, or manually look at admin and resend
	}
}

///*** HELPER METHODS ONLY BELOW THIS LINE **///
func generateEmail(templateFileName string, vars jet.VarMap) (string, error) {
	var View = jet.NewHTMLSet("./views/email")
	vw, err := View.GetTemplate(templateFileName)
	if err != nil {
		fmt.Println("failed to get template", err)
		return "", err
	}

	buf := bytes.Buffer{}
	if err = vw.Execute(&buf, vars, nil); err != nil {
		// error when executing template
		fmt.Println("failed to execute template", err)
		return "", err
	}

	body := buf.Bytes()

	return string(body), nil
}

// Internal function to send the email. Use specific email method to send email
func sendEmail(toEmail, fromEmail, subject, text, html string) error {

	sess := session.New(&aws.Config{
		Region:      aws.String(emailConfig.AWSRegion),
		Credentials: credentials.NewStaticCredentials(emailConfig.AWSAccessKeyID, emailConfig.AWSSecretAccessKey, ""),
	})

	//sess, err := session.NewSession()
	// if err != nil {
	// 	fmt.Println("failed to create session,", err)
	// 	return err
	// }

	svc := ses.New(sess)
	params := &ses.SendEmailInput{
		Destination: &ses.Destination{ // Required
			ToAddresses: []*string{
				aws.String(toEmail), // Required
			},
		},
		Message: &ses.Message{ // Required
			Body: &ses.Body{ // Required
				Html: &ses.Content{
					Data: aws.String(html), // Required
				},
				Text: &ses.Content{
					Data: aws.String(text), // Required
				},
			},
			Subject: &ses.Content{ // Required
				Data: aws.String(subject), // Required
			},
		},
		Source: aws.String(fromEmail), // Required
		ReplyToAddresses: []*string{
			aws.String(fromEmail), // Required
		},
	}
	//
	log.Println("sending email.......")
	resp, err := svc.SendEmail(params)

	if err != nil {
		// Print the error, cast err to awserr.Error to get the Code and
		// Message from an error.
		fmt.Println(err.Error())
		return err
	}
	log.Println("Done..............")
	// Pretty-print the response data.
	fmt.Println(resp)

	return nil
}
