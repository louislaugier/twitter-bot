package email

import (
	"errors"
	"fmt"
	"net"
	"net/mail"
	"net/smtp"
	"regexp"
	"strings"
	"time"
)

const (
	ForceDisconnectAfter = time.Second * 15
	SmtpPort             = 25
)

var (
	ErrBadFormat        = errors.New("invalid format")
	ErrUnresolvableHost = errors.New("unresolvable host")
	EmailRegexp         = regexp.MustCompile(`(?m)^(((((((((\s? +)?(\(((\s? +)?(([!-'*-[\]-~]*)|(\\([ -~]|\s))))*(\s? +)?\)))(\s? +)?)|(\s? +))?([A-Za-z0-9!#-'*+\/=?^_\x60{|}~-])+((((\s? +)?(\(((\s? +)?(([!-'*-[\]-~]*)|(\\([ -~]|\s))))*(\s? +)?\)))(\s? +)?)|(\s? +))?)|(((((\s? +)?(\(((\s? +)?(([!-'*-[\]-~]*)|(\\([ -~]|\s))))*(\s? +)?\)))(\s? +)?)|(\s? +))?"((\s? +)?(([!#-[\]-~])|(\\([ -~]|\s))))*(\s? +)?"))?)?(((((\s? +)?(\(((\s? +)?(([!-'*-[\]-~]*)|(\\([ -~]|\s))))*(\s? +)?\)))(\s? +)?)|(\s? +))?<(((((((\s? +)?(\(((\s? +)?(([!-'*-[\]-~]*)|(\\([ -~]|\s))))*(\s? +)?\)))(\s? +)?)|(\s? +))?(([A-Za-z0-9!#-'*+\/=?^_\x60{|}~-])+(\.([A-Za-z0-9!#-'*+\/=?^_\x60{|}~-])+)*)((((\s? +)?(\(((\s? +)?(([!-'*-[\]-~]*)|(\\([ -~]|\s))))*(\s? +)?\)))(\s? +)?)|(\s? +))?)|(((((\s? +)?(\(((\s? +)?(([!-'*-[\]-~]*)|(\\([ -~]|\s))))*(\s? +)?\)))(\s? +)?)|(\s? +))?"((\s? +)?(([!#-[\]-~])|(\\([ -~]|\s))))*(\s? +)?"))@((((((\s? +)?(\(((\s? +)?(([!-'*-[\]-~]*)|(\\([ -~]|\s))))*(\s? +)?\)))(\s? +)?)|(\s? +))?(([A-Za-z0-9!#-'*+\/=?^_\x60{|}~-])+(\.([A-Za-z0-9!#-'*+\/=?^_\x60{|}~-])+)*)((((\s? +)?(\(((\s? +)?(([!-'*-[\]-~]*)|(\\([ -~]|\s))))*(\s? +)?\)))(\s? +)?)|(\s? +))?)|(((((\s? +)?(\(((\s? +)?(([!-'*-[\]-~]*)|(\\([ -~]|\s))))*(\s? +)?\)))(\s? +)?)|(\s? +))?\[((\s? +)?([!-Z^-~]))*(\s? +)?\]((((\s? +)?(\(((\s? +)?(([!-'*-[\]-~]*)|(\\([ -~]|\s))))*(\s? +)?\)))(\s? +)?)|(\s? +))?)))>((((\s? +)?(\(((\s? +)?(([!-'*-[\]-~]*)|(\\([ -~]|\s))))*(\s? +)?\)))(\s? +)?)|(\s? +))?))|(((((((\s? +)?(\(((\s? +)?(([!-'*-[\]-~]*)|(\\([ -~]|\s))))*(\s? +)?\)))(\s? +)?)|(\s? +))?(([A-Za-z0-9!#-'*+\/=?^_\x60{|}~-])+(\.([A-Za-z0-9!#-'*+\/=?^_\x60{|}~-])+)*)((((\s? +)?(\(((\s? +)?(([!-'*-[\]-~]*)|(\\([ -~]|\s))))*(\s? +)?\)))(\s? +)?)|(\s? +))?)|(((((\s? +)?(\(((\s? +)?(([!-'*-[\]-~]*)|(\\([ -~]|\s))))*(\s? +)?\)))(\s? +)?)|(\s? +))?"((\s? +)?(([!#-[\]-~])|(\\([ -~]|\s))))*(\s? +)?"))@((((((\s? +)?(\(((\s? +)?(([!-'*-[\]-~]*)|(\\([ -~]|\s))))*(\s? +)?\)))(\s? +)?)|(\s? +))?(([A-Za-z0-9!#-'*+\/=?^_\x60{|}~-])+(\.([A-Za-z0-9!#-'*+\/=?^_\x60{|}~-])+)*)((((\s? +)?(\(((\s? +)?(([!-'*-[\]-~]*)|(\\([ -~]|\s))))*(\s? +)?\)))(\s? +)?)|(\s? +))?)|(((((\s? +)?(\(((\s? +)?(([!-'*-[\]-~]*)|(\\([ -~]|\s))))*(\s? +)?\)))(\s? +)?)|(\s? +))?\[((\s? +)?([!-Z^-~]))*(\s? +)?\]((((\s? +)?(\(((\s? +)?(([!-'*-[\]-~]*)|(\\([ -~]|\s))))*(\s? +)?\)))(\s? +)?)|(\s? +))?))))$`)
	CommonUsernames     = []string{"contact", "team", "marketing", "info", "infos", "information", "informations", "commercial", "rh", "hr", "recrutement", "support", "admin", "webmaster", "feedback", "help", "sales", "billing", "hello", "career", "careers"}
)

type smtpError struct {
	Err error
}

func (e smtpError) Error() string {
	return e.Err.Error()
}

func newSmtpError(err error) smtpError {
	return smtpError{
		Err: err,
	}
}

func split(email string) (account, host string) {
	i := strings.LastIndexByte(email, '@')
	if i < 0 {
		return
	}
	account = email[:i]
	host = email[i+1:]
	return
}

func validateFormat(email string) error {
	_, err := mail.ParseAddress(email)
	// if err != nil || !EmailRegexp.MatchString(strings.ToLower(email)) || strings.Contains(email, "no-reply") || strings.Contains(email, "noreply") || strings.Contains(email, "no_reply") {
	if err != nil || !EmailRegexp.MatchString(strings.ToLower(email)) {
		return ErrBadFormat
	}

	return nil
}

func validateHost(host string) (*smtp.Client, error) {
	hosts, err := getMX(host)
	if err != nil {
		return nil, err
	}

	client, err := dialTimeout(fmt.Sprintf("%s:%d", hosts[0], SmtpPort), ForceDisconnectAfter)
	if err != nil {
		return nil, newSmtpError(err)
	}

	return client, nil
}

func getMX(emailOrHost string) ([]string, error) {
	mx, err := net.LookupMX(emailOrHost)
	if err != nil {
		return nil, ErrUnresolvableHost
	}
	var hosts []string
	for _, mxRecord := range mx {
		hosts = append(hosts, mxRecord.Host)
	}
	return hosts, nil
}

func dialTimeout(addr string, timeout time.Duration) (*smtp.Client, error) {
	conn, err := net.DialTimeout("tcp", addr, timeout)
	if err != nil {
		return nil, err
	}
	t := time.AfterFunc(timeout, func() { conn.Close() })
	defer t.Stop()
	host, _, _ := net.SplitHostPort(addr)
	return smtp.NewClient(conn, host)
}

func ValidateEmailAddress(email string) error {
	_, host := split(email)

	if err := validateFormat(email); err != nil {
		return err
	}

	client, err := validateHost(host)
	if err != nil {
		return err
	}
	defer client.Close()

	err = client.Hello(host)
	if err != nil {
		return newSmtpError(err)
	}

	err = client.Mail(email)
	if err != nil {
		return newSmtpError(err)
	}

	err = client.Rcpt(email)
	if err != nil {
		return newSmtpError(err)
	}

	// w, err := client.Data()
	// if err != nil {
	// 	return newSmtpError(err)
	// }

	// message := []byte(
	// 	"From: Your Name <your_email@example.com>\r\n" +
	// 		"To: Recipient <recipient@example.com>\r\n" +
	// 		"Subject: \r\n" +
	// 		"Message-ID: <unique-message-id@example.com>\r\n" +
	// 		"\r\n",
	// )

	// _, err = w.Write(message)
	// if err != nil {
	// 	return newSmtpError(err)
	// }

	// err = w.Close()
	// if err != nil {
	// 	return newSmtpError(err)
	// }

	return nil
}
