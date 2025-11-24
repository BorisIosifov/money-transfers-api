// Package classification Money Transfer API.
//
// # Using in application
//
//	Schemes: https
//	Host:
//	BasePath: /api/
//	Version: 0.0.1
//
//	Consumes:
//	- application/json
//
//	Produces:
//	- application/json
//
//	Security:
//	- Bearer:
//
//	SecurityDefinitions:
//	Bearer:
//	     type: apiKey
//	     name: Authorization
//	     in: header
//
// swagger:meta
package controller

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"goji.io"
	"goji.io/pat"

	"github.com/BorisIosifov/money-transfers-api/model"
	"github.com/google/uuid"
)

type Controller struct {
	Config     model.Config
	DB         model.DBWrapper
	NeedToStop chan bool
}

func (ctrl Controller) NotFoundPage(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/json")
	w.WriteHeader(http.StatusNotFound)
	ctrl.PrintError(w, r, fmt.Errorf("Page not found"))
}

func (ctrl Controller) TestPage(w http.ResponseWriter, r *http.Request) {
	log.Print("Test page")
	fmt.Fprintln(w, "Test page")
}

func (ctrl Controller) Run() {
	mux := goji.NewMux()
	ctrl.handleFunc(mux, "GET", "/test", ctrl.TestPage)

	ctrl.handleFunc(mux, "GET", "/public/rates", ctrl.GetPublicRates)

	ctrl.handleFunc(mux, "POST", "/auth", ctrl.Auth)
	ctrl.handleFunc(mux, "GET", "/auth/send_code", ctrl.SendEmailCode)
	ctrl.handleFunc(mux, "GET", "/auth/check_code", ctrl.CheckCode)
	ctrl.handleFunc(mux, "GET", "/auth/check_user", ctrl.CheckUser)
	ctrl.handleFunc(mux, "POST", "/auth/register", ctrl.Register)
	ctrl.handleFunc(mux, "GET", "/auth/send_recovery_code", ctrl.SendRecoveryCode)
	ctrl.handleFunc(mux, "GET", "/auth/check_recovery_code", ctrl.CheckRecoveryCode)
	ctrl.handleFunc(mux, "PUT", "/auth/change_password_by_code", ctrl.ChangePasswordByCode)

	// ctrl.handleFunc(mux, "PUT", "/transport_area/:id", ctrl.PutTransportArea)

	mux.Use(func(h http.Handler) http.Handler {
		mw := func(w http.ResponseWriter, r *http.Request) {
			log.Printf("%s %s\n", r.Method, r.RequestURI)

			// user, tokenError := ctrl.DecodeAndValidateToken(w, r)

			if strings.HasSuffix(r.URL.Path, "/pdf") {
				w.Header().Set("Content-Type", "application/pdf")
			} else {
				w.Header().Set("Content-Type", "application/json")
			}

			ctrl.CORS(w, r)

			if r.Method != "OPTIONS" {
				session, err := ctrl.ManageSession(w, r)
				if err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					ctrl.PrintError(w, r, fmt.Errorf("Internal server Error: %s", err))
					return
				}
				ctx := context.WithValue(r.Context(), "Session", session)

				b, err := io.ReadAll(r.Body)
				if err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					ctrl.PrintError(w, r, fmt.Errorf("Internal server Error: %s", err))
					return
				}
				ctx = context.WithValue(ctx, "Body", b)

				h.ServeHTTP(w, r.WithContext(ctx))
			}
		}
		return http.HandlerFunc(mw)
	})

	srv := &http.Server{
		Handler: mux,
		Addr:    ":" + strconv.Itoa(ctrl.Config.Port),
	}

	go func() {
		<-ctrl.NeedToStop

		if err := srv.Shutdown(context.Background()); err != nil {
			log.Printf("HTTP server Shutdown: %v", err)
		}
	}()

	log.Fatal(srv.ListenAndServe())
}

func (ctrl Controller) Destroy() {
	ctrl.NeedToStop <- true
}

func (ctrl Controller) handleFunc(mux *goji.Mux, method string, pattern string, h func(http.ResponseWriter, *http.Request)) {
	var f *pat.Pattern
	switch method {
	case "GET":
		f = pat.Get(pattern)
	case "POST":
		f = pat.Post(pattern)
	case "PUT":
		f = pat.Put(pattern)
	case "DELETE":
		f = pat.Delete(pattern)
	case "PATCH":
		f = pat.Patch(pattern)
	}
	mux.HandleFunc(pat.Options(pattern), func(w http.ResponseWriter, r *http.Request) {})

	mux.HandleFunc(f, ctrl.timeoutHandler(h))
}

func (ctrl Controller) timeoutHandler(h func(http.ResponseWriter, *http.Request)) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		done := make(chan struct{})
		go func() {
			defer func() {
				if p := recover(); p != nil {
					w.WriteHeader(http.StatusInternalServerError)
					ctrl.PrintError(w, r, fmt.Errorf("Internal server error: %v", p))
				}
				close(done)
			}()

			h(w, r)
		}()

		select {
		case <-done:
			return
		case <-time.After(30 * time.Second):
			w.WriteHeader(http.StatusGatewayTimeout)
			ctrl.PrintError(w, r, fmt.Errorf("Operation timed out"))
			return
		}
	}
}

func (ctrl Controller) CORS(w http.ResponseWriter, r *http.Request) {
	env := os.Getenv("AGRO_ENV")
	if env == "test" {
		reqDump, err := httputil.DumpRequest(r, true)
		if err != nil {
			log.Printf("Error: %s\n", err)
		}

		log.Print(string(reqDump))
	}

	var origin = r.Header.Get("Origin")
	if origin == "" {
		origin = "*"
	}
	var headers = r.Header.Get("Access-Control-Request-Headers")
	if headers != "" {
		w.Header().Set("Access-Control-Allow-Headers", headers)
	}
	w.Header().Set("Access-Control-Allow-Origin", origin)
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Expose-Headers", "*")
}

// swagger:model
type successResult struct {
	Status string `json:"status"`
}

// swagger:model
type errorResult struct {
	Status string `json:"status"`
	Err    string `json:"error"`
}

func (ctrl Controller) PrintError(w http.ResponseWriter, r *http.Request, err error) {
	res := errorResult{Status: "error", Err: err.Error()}
	log.Print(err.Error())
	resJSON, _ := json.Marshal(res)
	fmt.Fprintln(w, string(resJSON))
	if strings.Contains(err.Error(), "Internal server error") ||
		strings.Contains(err.Error(), "Operation timed out") {
		go ctrl.SendErrorReport(r, err)
	}
}

func (ctrl Controller) SendErrorReport(r *http.Request, err error) {
	body := r.Context().Value("Body")
	var b []byte
	if body != nil {
		b = body.([]byte)
	}
	// SESSION
	sessionID := ""

	u, _ := url.Parse("https://api.telegram.org/bot6608246246:AAHZuBgwAEruYD4MIaZspuBWuUo5FNNyx8s/sendMessage")
	q := u.Query()
	q.Set("chat_id", ctrl.Config.TelegramChatID)
	q.Set("text", fmt.Sprintf("%s\n%s %s\n%s\n%s\n",
		err.Error(), r.Method, r.RequestURI, sessionID, string(b)))

	u.RawQuery = q.Encode()
	log.Print(u.String())
	resp, err := http.Get(u.String())
	if err != nil {
		return
	}
	defer resp.Body.Close()
	bodyBytes, _ := ioutil.ReadAll(resp.Body)

	log.Print(string(bodyBytes))
}

func fillObjectFromForm(r *http.Request, obj interface{}) (err error) {
	b := r.Context().Value("Body").([]byte)

	log.Print(string(b))
	dec := json.NewDecoder(bytes.NewReader(b))
	dec.DisallowUnknownFields()
	err = dec.Decode(obj)
	return err
}

type GetParamDesc struct {
	name         string
	paramType    string
	required     bool
	defaultValue interface{}
}

func (ctrl Controller) validateParams(w http.ResponseWriter, r *http.Request, params []GetParamDesc) (
	result map[string]interface{}, ok bool) {
	var errors string
	result = make(map[string]interface{})
	for _, param := range params {
		valuesArray, ok := r.Form[param.name]
		if param.required && !ok {
			errors += fmt.Sprintf("missing required parameter '%s'\n", param.name)

		} else if ok {
			switch param.paramType {
			case "int":
				n, err := strconv.Atoi(valuesArray[0])
				if err != nil {
					errors += fmt.Sprintf("%s is not a number: %s\n", param.name, err)
				}
				result[param.name] = n

			case "[]int":
				var ar []int
				for _, value := range valuesArray {
					n, err := strconv.Atoi(value)
					if err != nil {
						errors += fmt.Sprintf("%s is not a number: %s\n", param.name, err)
					}
					ar = append(ar, n)
				}
				result[param.name] = ar

			case "date":
				t, err := time.Parse("20060102", valuesArray[0])
				if err != nil {
					errors += fmt.Sprintf("%s is not a date: %s\n", param.name, err)
				}
				result[param.name] = t

			default:
				result[param.name] = valuesArray[0]
			}

		} else if !ok && param.defaultValue != nil {
			result[param.name] = param.defaultValue
		}
	}

	for key, _ := range r.Form {
		_, ok = result[key]
		if !ok {
			errors += fmt.Sprintf("unknown parameter '%s'\n", key)
		}
	}

	if len(errors) > 0 {
		w.WriteHeader(http.StatusBadRequest)
		ctrl.PrintError(w, r, fmt.Errorf("%v", errors))
		return nil, false
	}

	return result, true
}

func (ctrl Controller) ManageSession(w http.ResponseWriter, r *http.Request) (session model.Session, err error) {
	var needCreateSession bool

	sessionCookie, err := r.Cookie("SessionID")
	if err != nil {
		if errors.Is(err, http.ErrNoCookie) {
			needCreateSession = true
		} else {
			return session, err
		}
	} else {
		session, err = model.GetSession(ctrl.DB, sessionCookie.Value)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				needCreateSession = true
			} else {
				return session, err
			}
		}
	}

	if needCreateSession {
		sessionID := uuid.New().String()

		// Starting a transaction
		tx := ctrl.DB.MustBegin()
		defer tx.Rollback()

		session = model.Session{SessionID: sessionID, Data: "{}"}
		err := session.Create(tx)
		if err != nil {
			return session, err
		}

		err = tx.Commit()
		if err != nil {
			return session, err
		}

		cookie := http.Cookie{
			Name:     "SessionID",
			Value:    sessionID,
			Path:     "/",
			MaxAge:   356 * 24 * 3600,
			HttpOnly: true,
			Secure:   true,
			SameSite: http.SameSiteLaxMode,
		}

		http.SetCookie(w, &cookie)
	}

	return session, nil
}

// Dirty hack - unused types - just to make swagger to show correct types
// swagger:model
type NullString string

// swagger:model
type NullInt int

// swagger:model
type NullFloat float64

// swagger:model
type NullBool bool

// swagger:model
type IntArrayAsString []int
