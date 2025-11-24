package controller

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"net/http"
	"net/mail"
	"time"

	"gopkg.in/gomail.v2"

	"github.com/BorisIosifov/money-transfers-api/model"
)

// swagger:model
type AuthenticationInfo struct {
	Email    string
	Password string
}

// swagger:route POST /auth Auth PostAuth
//
// Authenticate user
//
//	Parameters:
//	  + name: AuthenticationInfo
//	    in: body
//	    description: Login and password
//	    required: true
//	    type: AuthenticationInfo
//	Responses:
//	  default: errorResult
//	  200: Token
//	  400: errorResult
//	  405: errorResult
//	  500: errorResult
func (ctrl Controller) Auth(w http.ResponseWriter, r *http.Request) {
	var input AuthenticationInfo
	err := fillObjectFromForm(r, &input)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		ctrl.PrintError(w, r, fmt.Errorf("Internal server error: %s", err))
		return
	}

	user, err := model.GetUserByEmailAndPassword(ctrl.DB, input.Email, input.Password)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			w.WriteHeader(http.StatusNotFound)
			ctrl.PrintError(w, r, fmt.Errorf("Wrong login or password"))
			return
		}
		ctrl.PrintError(w, r, err)
		return
	}

	resJSON, err := json.Marshal(user)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		ctrl.PrintError(w, r, fmt.Errorf("Internal server error: %s", err))
		return
	}
	fmt.Fprintln(w, string(resJSON))
}

// swagger:route GET /auth/send_code Auth SendEmailCode
//
// Send email code
//
//	Parameters:
//	  + name: Email
//	    in: query
//	    description: Email
//	    required: true
//	    type: string
//	Responses:
//	  default: errorResult
//	  200: successResult
//	  400: errorResult
//	  500: errorResult
func (ctrl Controller) SendEmailCode(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		ctrl.PrintError(w, r, fmt.Errorf("Internal server error: %s", err))
		return
	}

	params := []GetParamDesc{
		{name: "Email", paramType: "string", required: true},
	}
	validatedParams, paramsOK := ctrl.validateParams(w, r, params)
	if !paramsOK {
		return
	}

	email := validatedParams["Email"].(string)

	_, err = mail.ParseAddress(email)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		ctrl.PrintError(w, r, fmt.Errorf("Email is wrong"))
		return
	}

	lastEmailCode, err := model.GetLastEmailCode(ctrl.DB, email, "registration")
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		w.WriteHeader(http.StatusInternalServerError)
		ctrl.PrintError(w, r, fmt.Errorf("Internal server error: %s", err))
		return
	} else if err == nil && time.Now().Before(lastEmailCode.CTimeObj.Add(time.Minute)) {
		w.WriteHeader(http.StatusMethodNotAllowed)
		ctrl.PrintError(w, r, fmt.Errorf("Last code was sent less then minute ago"))
		return
	}

	code := fmt.Sprintf("%04d", rand.Intn(10000))
	emailCode := model.EmailCode{Email: email, Code: code, Type: "registration"}

	// Starting a transaction
	tx := ctrl.DB.MustBegin()
	defer tx.Rollback()

	err = emailCode.Create(tx)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		ctrl.PrintError(w, r, fmt.Errorf("Internal server error: %s", err))
		return
	}

	// Finishing the transaction
	err = tx.Commit()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		ctrl.PrintError(w, r, fmt.Errorf("Internal server error: %s", err))
		return
	}

	// Send email
	text := fmt.Sprintf("Code: %s", code)
	m := gomail.NewMessage()
	m.SetAddressHeader("From", "noreply@shekelrubl.co.il", "Shekel Rubl")
	m.SetHeader("To", email)
	// m.SetHeader("Bcc", "")
	// m.SetHeader("Reply-To", "")
	m.SetHeader("Subject", "Schekel Rubl Code")
	m.SetBody("text/html", text)
	//m.Attach("")

	d := gomail.NewDialer(ctrl.Config.SMTPhost, ctrl.Config.SMTPport, ctrl.Config.SMTPlogin, ctrl.Config.SMTPpassword)

	// Send the email
	err = d.DialAndSend(m)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		ctrl.PrintError(w, r, fmt.Errorf("Internal server error: %s", err))
		return
	}

	fmt.Fprintln(w, "{\"status\": \"ok\"}")
}

func (ctrl Controller) CheckCodeCommon(w http.ResponseWriter, r *http.Request,
	email, code, codeType string) (status int, err error) {

	emailCode, err := model.GetLastEmailCode(ctrl.DB, email, codeType)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return http.StatusInternalServerError, fmt.Errorf("Internal server error: %s", err)
	} else if err != nil && errors.Is(err, sql.ErrNoRows) {
		return http.StatusMethodNotAllowed, fmt.Errorf("Code unexists")
	} else if emailCode.Attempts > 5 {
		return http.StatusMethodNotAllowed, fmt.Errorf("Number of attempts exceeded")
	} else if time.Now().After(emailCode.CTimeObj.Add(5 * time.Minute)) {
		return http.StatusMethodNotAllowed, fmt.Errorf("Code is older then 5 minutes")
	} else if emailCode.Code != code {
		tx := ctrl.DB.MustBegin()
		defer tx.Rollback()
		err = emailCode.IncreaseAttempts(tx)
		if err != nil {
			return http.StatusInternalServerError, fmt.Errorf("Internal server error: %s", err)
		}
		_ = tx.Commit()

		return http.StatusMethodNotAllowed, fmt.Errorf("Code is wrong")
	}

	return http.StatusOK, nil
}

// swagger:route GET /auth/check_code Auth CheckCode
//
// Check sms code
//
//	Parameters:
//	  + name: Email
//	    in: query
//	    description: Email
//	    required: true
//	    type: string
//	  + name: Code
//	    in: query
//	    description: Code
//	    required: true
//	    type: string
//	Responses:
//	  default: errorResult
//	  200: successResult
//	  400: errorResult
//	  500: errorResult
func (ctrl Controller) CheckCode(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		ctrl.PrintError(w, r, fmt.Errorf("Internal server error: %s", err))
		return
	}

	params := []GetParamDesc{
		{name: "Email", paramType: "string", required: true},
		{name: "Code", paramType: "string", required: true},
	}
	validatedParams, paramsOK := ctrl.validateParams(w, r, params)
	if !paramsOK {
		return
	}

	status, err := ctrl.CheckCodeCommon(w, r,
		validatedParams["Email"].(string), validatedParams["Code"].(string), "registration")
	if err != nil {
		w.WriteHeader(status)
		ctrl.PrintError(w, r, err)
		return
	}

	fmt.Fprintln(w, "{\"status\": \"ok\"}")
}

// swagger:route GET /auth/check_user Auth CheckUser
//
// Check existing user
//
//	Parameters:
//	  + name: Email
//	    in: query
//	    description: Email
//	    required: true
//	    type: string
//	Responses:
//	  default: errorResult
//	  200: successResult
//	  400: errorResult
//	  500: errorResult
func (ctrl Controller) CheckUser(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		ctrl.PrintError(w, r, fmt.Errorf("Internal server error: %s", err))
		return
	}

	params := []GetParamDesc{
		{name: "Email", paramType: "string", required: true},
	}
	validatedParams, paramsOK := ctrl.validateParams(w, r, params)
	if !paramsOK {
		return
	}

	userExists, err := model.DoesUserExist(ctrl.DB, validatedParams["Email"].(string))
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		ctrl.PrintError(w, r, fmt.Errorf("Internal server error: %s", err))
		return
	}
	if userExists {
		w.WriteHeader(http.StatusMethodNotAllowed)
		ctrl.PrintError(w, r, fmt.Errorf("User already exists"))
		return
	}

	fmt.Fprintln(w, "{\"status\": \"ok\"}")
}

// swagger:model
type RegistrationInfo struct {
	Email    string
	Name     string
	Password string
	Code     string
}

// swagger:route POST /auth/register Auth Register
//
// Authenticate user
//
//	Parameters:
//	  + name: RegistrationInfo
//	    in: body
//	    description: Registration Info
//	    required: true
//	    type: RegistrationInfo
//	Responses:
//	  default: errorResult
//	  200: Token
//	  400: errorResult
//	  405: errorResult
//	  500: errorResult
func (ctrl Controller) Register(w http.ResponseWriter, r *http.Request) {
	var input RegistrationInfo
	err := fillObjectFromForm(r, &input)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		ctrl.PrintError(w, r, fmt.Errorf("Internal server error: %s", err))
		return
	}

	// Check code
	status, err := ctrl.CheckCodeCommon(w, r, input.Email, input.Code, "registration")
	if err != nil {
		w.WriteHeader(status)
		ctrl.PrintError(w, r, err)
		return
	}

	// Check existing user
	userExists, err := model.DoesUserExist(ctrl.DB, input.Email)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		ctrl.PrintError(w, r, fmt.Errorf("Internal server error: %s", err))
		return
	}
	if userExists {
		w.WriteHeader(http.StatusMethodNotAllowed)
		ctrl.PrintError(w, r, fmt.Errorf("User already exists"))
		return
	}

	// Starting a transaction
	tx := ctrl.DB.MustBegin()
	defer tx.Rollback()

	// Create user
	user := model.User{
		Email:             input.Email,
		PasswordUncrypted: input.Password,
		Name:              input.Name,
	}

	user, err = user.Create(tx)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		ctrl.PrintError(w, r, fmt.Errorf("Internal server error: %s", err))
		return
	}

	// Login user
	session := r.Context().Value("Session").(model.Session)
	session.UserID.Valid = true
	session.UserID.Int32 = int32(user.UserID)
	err = session.Update(tx)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		ctrl.PrintError(w, r, fmt.Errorf("Internal server error: %s", err))
		return
	}

	// Finishing the transaction
	err = tx.Commit()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		ctrl.PrintError(w, r, fmt.Errorf("Internal server error: %s", err))
		return
	}

	fmt.Fprintln(w, "{\"status\": \"ok\"}")
}

// swagger:route GET /auth/send_recovery_code Auth SendEmailCode
//
// Send email code
//
//	Parameters:
//	  + name: Email
//	    in: query
//	    description: Email
//	    required: true
//	    type: string
//	Responses:
//	  default: errorResult
//	  200: successResult
//	  400: errorResult
//	  500: errorResult
func (ctrl Controller) SendRecoveryCode(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		ctrl.PrintError(w, r, fmt.Errorf("Internal server error: %s", err))
		return
	}

	params := []GetParamDesc{
		{name: "Email", paramType: "string", required: true},
	}
	validatedParams, paramsOK := ctrl.validateParams(w, r, params)
	if !paramsOK {
		return
	}

	email := validatedParams["Email"].(string)

	userExists, err := model.DoesUserExist(ctrl.DB, email)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		ctrl.PrintError(w, r, fmt.Errorf("Internal server error: %s", err))
		return
	}
	if !userExists {
		w.WriteHeader(http.StatusMethodNotAllowed)
		ctrl.PrintError(w, r, fmt.Errorf("User doesn't exist"))
		return
	}

	lastEmailCode, err := model.GetLastEmailCode(ctrl.DB, email, "recovery")
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		w.WriteHeader(http.StatusInternalServerError)
		ctrl.PrintError(w, r, fmt.Errorf("Internal server error: %s", err))
		return
	} else if err == nil && time.Now().Before(lastEmailCode.CTimeObj.Add(time.Minute)) {
		w.WriteHeader(http.StatusMethodNotAllowed)
		ctrl.PrintError(w, r, fmt.Errorf("Last code was sent less then minute ago"))
		return
	}

	code := fmt.Sprintf("%08d%08d", rand.Intn(100000000), rand.Intn(100000000))
	emailCode := model.EmailCode{Email: email, Code: code, Type: "recovery"}

	// Starting a transaction
	tx := ctrl.DB.MustBegin()
	defer tx.Rollback()

	err = emailCode.Create(tx)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		ctrl.PrintError(w, r, fmt.Errorf("Internal server error: %s", err))
		return
	}

	// Finishing the transaction
	err = tx.Commit()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		ctrl.PrintError(w, r, fmt.Errorf("Internal server error: %s", err))
		return
	}

	// Send email
	text := fmt.Sprintf("Code: %s", code)
	m := gomail.NewMessage()
	m.SetAddressHeader("From", "noreply@shekelrubl.co.il", "Shekel Rubl")
	m.SetHeader("To", email)
	// m.SetHeader("Bcc", "")
	// m.SetHeader("Reply-To", "")
	m.SetHeader("Subject", "Schekel Rubl Code")
	m.SetBody("text/html", text)
	//m.Attach("")

	d := gomail.NewDialer(ctrl.Config.SMTPhost, ctrl.Config.SMTPport, ctrl.Config.SMTPlogin, ctrl.Config.SMTPpassword)

	// Send the email
	err = d.DialAndSend(m)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		ctrl.PrintError(w, r, fmt.Errorf("Internal server error: %s", err))
		return
	}

	fmt.Fprintln(w, "{\"status\": \"ok\"}")
}

// swagger:route GET /auth/check_recovery_code Auth CheckCode
//
// Check sms code
//
//	Parameters:
//	  + name: Email
//	    in: query
//	    description: Email
//	    required: true
//	    type: string
//	  + name: Code
//	    in: query
//	    description: Code
//	    required: true
//	    type: string
//	Responses:
//	  default: errorResult
//	  200: successResult
//	  400: errorResult
//	  500: errorResult
func (ctrl Controller) CheckRecoveryCode(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		ctrl.PrintError(w, r, fmt.Errorf("Internal server error: %s", err))
		return
	}

	params := []GetParamDesc{
		{name: "Email", paramType: "string", required: true},
		{name: "Code", paramType: "string", required: true},
	}
	validatedParams, paramsOK := ctrl.validateParams(w, r, params)
	if !paramsOK {
		return
	}

	status, err := ctrl.CheckCodeCommon(w, r,
		validatedParams["Email"].(string), validatedParams["Code"].(string), "recovery")
	if err != nil {
		w.WriteHeader(status)
		ctrl.PrintError(w, r, err)
		return
	}

	fmt.Fprintln(w, "{\"status\": \"ok\"}")
}

// swagger:model
type Recovery struct {
	Email    string
	Code     string
	Password string
}

// swagger:route PUT /auth/change_password_by_code Auth ChangePasswordByCode
//
// Authenticate user
//
//	Parameters:
//	  + name: Recovery
//	    in: body
//	    description: Token and password
//	    required: true
//	    type: Recovery
//	Responses:
//	  default: errorResult
//	  200: successResult
//	  400: errorResult
//	  405: errorResult
//	  500: errorResult
func (ctrl Controller) ChangePasswordByCode(w http.ResponseWriter, r *http.Request) {
	var input Recovery
	err := fillObjectFromForm(r, &input)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		ctrl.PrintError(w, r, fmt.Errorf("Internal server error: %s", err))
		return
	}

	status, err := ctrl.CheckCodeCommon(w, r, input.Email, input.Code, "recovery")
	if err != nil {
		w.WriteHeader(status)
		ctrl.PrintError(w, r, err)
		return
	}

	// Starting a transaction
	tx := ctrl.DB.MustBegin()
	defer tx.Rollback()

	// Change password
	model.User{Email: input.Email, PasswordUncrypted: input.Password}.UpdatePassword(tx)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		ctrl.PrintError(w, r, fmt.Errorf("Internal server error: %s", err))
		return
	}

	// Finishing the transaction
	err = tx.Commit()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		ctrl.PrintError(w, r, fmt.Errorf("Internal server error: %s", err))
		return
	}

	fmt.Fprintln(w, "{\"status\": \"ok\"}")
}
