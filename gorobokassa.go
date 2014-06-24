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
	QUERY_OUT_SUMM    = "OutSum"
	QUERY_INV_ID      = "InvId"
	QUERY_CRC         = "SignatureValue"
	QUERY_DESCRIPTION = "Desc"
	QUERY_LOGIN       = "MrchLogin"
	ROBOKASSA_HOST    = "auth.robokassa.ru"
	ROBOKASSA_PATH    = "Merchant/Index.aspx"
	SCHEME            = "https"
	DELIMETER         = ":"
)

var (
	ErrIncorrectValue = errors.New("incorrect value")
)

type Client struct {
	login          string
	firstPassword  string
	secondPassword string
}

// формирование URL переадресации пользователя на оплату
func (client *Client) Url(invoice, value int, description string) (string, error) {
	return buildRedirectUrl(client.login, client.firstPassword, invoice, value, description)
}

// получение уведомления об исполнении операции (ResultURL)
func (client *Client) CheckResult(r *http.Request) bool {
	return verifyRequest(client.secondPassword, r)
}

// проверка параметров в скрипте завершения операции (SuccessURL)
func (client *Client) CheckSuccess(r *http.Request) bool {
	return verifyRequest(client.firstPassword, r)
}

func New(login, password1, password2 string) *Client {
	return &Client{login, password1, password2}
}

// join values with delimeter and return hex of md5
func CRC(v ...interface{}) string {
	s := make([]string, len(v))
	for key, value := range v {
		s[key] = fmt.Sprintf("%v", value)
	}
	h := md5.New()
	io.WriteString(h, strings.Join(s, DELIMETER))
	return fmt.Sprintf("%x", h.Sum(nil))
}

func buildRedirectUrl(login, password string, invoice, value int, description string) (string, error) {
	if value < 0 {
		return "", ErrIncorrectValue
	}

	q := url.URL{}
	q.Host = ROBOKASSA_HOST
	q.Scheme = SCHEME
	q.Path = ROBOKASSA_PATH

	params := url.Values{}
	params.Add(QUERY_LOGIN, login)
	params.Add(QUERY_OUT_SUMM, strconv.Itoa(value))
	params.Add(QUERY_INV_ID, strconv.Itoa(invoice))
	params.Add(QUERY_DESCRIPTION, description)
	params.Add(QUERY_CRC, CRC(login, value, invoice, password))

	q.RawQuery = params.Encode()

	return q.String(), nil
}

func verifyResult(password string, invoice, value int, crc string) bool {
	return strings.ToUpper(crc) == strings.ToUpper(CRC(value, invoice, password))
}

func verifyRequest(password string, r *http.Request) bool {
	q := r.URL.Query()
	value, err := strconv.Atoi(q.Get(QUERY_OUT_SUMM))
	if err != nil {
		log.Println(err)
		return false
	}
	invoice, err := strconv.Atoi(q.Get(QUERY_INV_ID))
	if err != nil {
		log.Println(err)
		return false
	}
	crc := q.Get(QUERY_CRC)
	return verifyResult(password, invoice, value, crc)
}
