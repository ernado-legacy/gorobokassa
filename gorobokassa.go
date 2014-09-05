package gorobokassa

import (
	"crypto/md5"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

const (
	queryOutSumm     = "OutSum"
	queryInvID       = "InvId"
	queryCRC         = "SignatureValue"
	queryDescription = "Desc"
	queryLogin       = "MrchLogin"
	robokassaHost    = "auth.robokassa.ru"
	// robokassaHost = "test.robokassa.ru"
	robokassaPath = "Merchant/Index.aspx"
	robokassaPath = "Index.aspx"
	scheme        = "https"
	// scheme        = "http"
	delim = ":"
)

var (
	ErrBadRequest = errors.New("bad request")
)

// Client для генерации URL и проверки уведомлений
type Client struct {
	login          string
	firstPassword  string
	secondPassword string
}

// URL переадресации пользователя на оплату
func (client *Client) URL(invoice, value int, description string) string {
	return buildRedirectURL(client.login, client.firstPassword, invoice, value, description)
}

// CheckResult получение уведомления об исполнении операции (ResultURL)
func (client *Client) CheckResult(r *http.Request) bool {
	return verifyRequest(client.secondPassword, r)
}

func (client *Client) ResultInvoice(r *http.Request) (int, int, error) {
	return getInvoice(client.secondPassword, r)
}

// CheckSuccess проверка параметров в скрипте завершения операции (SuccessURL)
func (client *Client) CheckSuccess(r *http.Request) bool {
	return verifyRequest(client.firstPassword, r)
}

// New Client
func New(login, password1, password2 string) *Client {
	return &Client{login, password1, password2}
}

// CRC of joint values with delimeter
func CRC(v ...interface{}) string {
	s := make([]string, len(v))
	for key, value := range v {
		s[key] = fmt.Sprintf("%v", value)
	}
	h := md5.New()
	io.WriteString(h, strings.Join(s, delim))
	return fmt.Sprintf("%x", h.Sum(nil))
}

func buildRedirectURL(login, password string, invoice, value int, description string) string {
	q := url.URL{}
	q.Host = robokassaHost
	q.Scheme = scheme
	q.Path = robokassaPath

	params := url.Values{}
	params.Add(queryLogin, login)
	params.Add(queryOutSumm, strconv.Itoa(value))
	params.Add(queryInvID, strconv.Itoa(invoice))
	params.Add(queryDescription, description)
	params.Add(queryCRC, CRC(login, value, invoice, password))

	q.RawQuery = params.Encode()
	return q.String()
}

func verifyResult(password string, invoice, value int, crc string) bool {
	log.Println(crc, CRC(value, invoice, password), value, invoice, password)
	return strings.ToUpper(crc) == strings.ToUpper(CRC(value, invoice, password))
}

func getInvoice(password string, r *http.Request) (int, int, error) {
	q := r.URL.Query()
	value, err := strconv.Atoi(q.Get(queryOutSumm))
	if err != nil {
		log.Println(err)
		return 0, 0, ErrBadRequest
	}
	invoice, err := strconv.Atoi(q.Get(queryInvID))
	if err != nil {
		log.Println(err)
		return 0, 0, ErrBadRequest
	}
	crc := q.Get(queryCRC)
	if !verifyResult(password, invoice, value, crc) {
		log.Println("result not verified")
		return 0, 0, ErrBadRequest
	}
	return invoice, value, nil
}

func verifyRequest(password string, r *http.Request) bool {
	_, _, err := getInvoice(password, r)
	return err == nil
}
