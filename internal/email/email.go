package email

import (
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"net"
	"net/mail"
	"net/smtp"
	"os"
	"regexp"
	"strings"
	"sync"
	"time"
)

const (
	ForceDisconnectAfter = time.Second * 15
	SmtpPort             = 25
)

var (
	EmailRegexp     = regexp.MustCompile(`(?m)^(((((((((\s? +)?(\(((\s? +)?(([!-'*-[\]-~]*)|(\\([ -~]|\s))))*(\s? +)?\)))(\s? +)?)|(\s? +))?([A-Za-z0-9!#-'*+\/=?^_\x60{|}~-])+((((\s? +)?(\(((\s? +)?(([!-'*-[\]-~]*)|(\\([ -~]|\s))))*(\s? +)?\)))(\s? +)?)|(\s? +))?)|(((((\s? +)?(\(((\s? +)?(([!-'*-[\]-~]*)|(\\([ -~]|\s))))*(\s? +)?\)))(\s? +)?)|(\s? +))?"((\s? +)?(([!#-[\]-~])|(\\([ -~]|\s))))*(\s? +)?"))?)?(((((\s? +)?(\(((\s? +)?(([!-'*-[\]-~]*)|(\\([ -~]|\s))))*(\s? +)?\)))(\s? +)?)|(\s? +))?<(((((((\s? +)?(\(((\s? +)?(([!-'*-[\]-~]*)|(\\([ -~]|\s))))*(\s? +)?\)))(\s? +)?)|(\s? +))?(([A-Za-z0-9!#-'*+\/=?^_\x60{|}~-])+(\.([A-Za-z0-9!#-'*+\/=?^_\x60{|}~-])+)*)((((\s? +)?(\(((\s? +)?(([!-'*-[\]-~]*)|(\\([ -~]|\s))))*(\s? +)?\)))(\s? +)?)|(\s? +))?)|(((((\s? +)?(\(((\s? +)?(([!-'*-[\]-~]*)|(\\([ -~]|\s))))*(\s? +)?\)))(\s? +)?)|(\s? +))?"((\s? +)?(([!#-[\]-~])|(\\([ -~]|\s))))*(\s? +)?"))@((((((\s? +)?(\(((\s? +)?(([!-'*-[\]-~]*)|(\\([ -~]|\s))))*(\s? +)?\)))(\s? +)?)|(\s? +))?(([A-Za-z0-9!#-'*+\/=?^_\x60{|}~-])+(\.([A-Za-z0-9!#-'*+\/=?^_\x60{|}~-])+)*)((((\s? +)?(\(((\s? +)?(([!-'*-[\]-~]*)|(\\([ -~]|\s))))*(\s? +)?\)))(\s? +)?)|(\s? +))?)|(((((\s? +)?(\(((\s? +)?(([!-'*-[\]-~]*)|(\\([ -~]|\s))))*(\s? +)?\)))(\s? +)?)|(\s? +))?\[((\s? +)?([!-Z^-~]))*(\s? +)?\]((((\s? +)?(\(((\s? +)?(([!-'*-[\]-~]*)|(\\([ -~]|\s))))*(\s? +)?\)))(\s? +)?)|(\s? +))?)))>((((\s? +)?(\(((\s? +)?(([!-'*-[\]-~]*)|(\\([ -~]|\s))))*(\s? +)?\)))(\s? +)?)|(\s? +))?))|(((((((\s? +)?(\(((\s? +)?(([!-'*-[\]-~]*)|(\\([ -~]|\s))))*(\s? +)?\)))(\s? +)?)|(\s? +))?(([A-Za-z0-9!#-'*+\/=?^_\x60{|}~-])+(\.([A-Za-z0-9!#-'*+\/=?^_\x60{|}~-])+)*)((((\s? +)?(\(((\s? +)?(([!-'*-[\]-~]*)|(\\([ -~]|\s))))*(\s? +)?\)))(\s? +)?)|(\s? +))?)|(((((\s? +)?(\(((\s? +)?(([!-'*-[\]-~]*)|(\\([ -~]|\s))))*(\s? +)?\)))(\s? +)?)|(\s? +))?"((\s? +)?(([!#-[\]-~])|(\\([ -~]|\s))))*(\s? +)?"))@((((((\s? +)?(\(((\s? +)?(([!-'*-[\]-~]*)|(\\([ -~]|\s))))*(\s? +)?\)))(\s? +)?)|(\s? +))?(([A-Za-z0-9!#-'*+\/=?^_\x60{|}~-])+(\.([A-Za-z0-9!#-'*+\/=?^_\x60{|}~-])+)*)((((\s? +)?(\(((\s? +)?(([!-'*-[\]-~]*)|(\\([ -~]|\s))))*(\s? +)?\)))(\s? +)?)|(\s? +))?)|(((((\s? +)?(\(((\s? +)?(([!-'*-[\]-~]*)|(\\([ -~]|\s))))*(\s? +)?\)))(\s? +)?)|(\s? +))?\[((\s? +)?([!-Z^-~]))*(\s? +)?\]((((\s? +)?(\(((\s? +)?(([!-'*-[\]-~]*)|(\\([ -~]|\s))))*(\s? +)?\)))(\s? +)?)|(\s? +))?))))$`)
	CommonUsernames = []string{"contact", "team", "marketing", "info", "infos", "information", "informations", "commercial", "rh", "hr", "recrutement", "support", "admin", "webmaster", "feedback", "help", "sales", "billing", "hello", "career", "careers"}
)

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
		return err
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
		return nil, err
	}

	return client, nil
}

func getMX(emailOrHost string) ([]string, error) {
	mx, err := net.LookupMX(emailOrHost)
	if err != nil {
		return nil, err
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
		return err
	}

	err = client.Mail(email)
	if err != nil {
		return err
	}

	err = client.Rcpt(email)
	if err != nil {
		return err
	}

	// w, err := client.Data()
	// if err != nil {
	// 	return err
	// }

	// ID := uuid.New()
	// message := []byte(
	// 	fmt.Sprintf(("From: Your Name <your_email@example.com>\r\n" +
	// 		"To: Recipient <%s>\r\n" +
	// 		"Subject: A subject\r\n" +
	// 		"Message-ID: <%s>\r\n" +
	// 		"\r\n"), email, ID),
	// )

	// _, err = w.Write(message)
	// if err != nil {
	// 	return err
	// }

	// w.Close()

	return nil
}

func GetValidEmailsFromCSVIntoNewCSV(inputPath string, outputPath string) {
	type EmailWithLine struct {
		Email      string
		LineNumber int
	}

	file, err := os.Open(inputPath)
	if err != nil {
		fmt.Println("Error opening file:", err)
		return
	}
	defer file.Close()

	reader := csv.NewReader(file)
	reader.Comma = ';' // Assuming each email is separated by a semicolon.

	// Buffered channel to store valid emails.
	validEmailsChan := make(chan EmailWithLine, 100)

	// Channel to signal when all processing is done.
	doneChan := make(chan struct{})

	// Slice to hold valid emails after validation.
	var validEmails []string

	// Mutex to protect access to validEmails slice.
	var mu sync.Mutex
	var wg sync.WaitGroup

	// Start a worker pool to validate emails.
	for i := 0; i < 16; i++ { // Worker pool size can be adjusted.
		wg.Add(1)
		go func() {
			defer wg.Done()
			for ewl := range validEmailsChan {
				if err := ValidateEmailAddress(ewl.Email); err == nil || strings.Contains(err.Error(), "553 ") || strings.Contains(err.Error(), "Spamhaus") || strings.Contains(err.Error(), "block") || strings.Contains(err.Error(), "Block") || strings.Contains(err.Error(), "spam") || strings.Contains(err.Error(), "Spam") || strings.Contains(err.Error(), "DNS") || strings.Contains(err.Error(), "dns") || strings.Contains(err.Error(), "abus") {
					if !strings.Contains(err.Error(), "DNS response contained records which contain invalid names") && !strings.Contains(err.Error(), "501") {
						mu.Lock()
						log.Printf("Valid email on line %d: %s.\n", ewl.LineNumber, ewl.Email)
						validEmails = append(validEmails, ewl.Email)
						mu.Unlock()
					}
				} else if ewl.Email != "EMAIL" {
					log.Printf("Invalid email on line %d: %s. Error: %s\n", ewl.LineNumber, ewl.Email, err)
				}
			}
		}()
	}

	// Go routine to wait for worker pool to finish and close done channel.
	go func() {
		wg.Wait()
		close(doneChan)
	}()

	// Read from the CSV and send emails to be validated.
	lineNumber := 1 // Initialize line number counter.
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Printf("Error reading CSV on line %d: %v\n", lineNumber, err)
			lineNumber++
			continue
		}
		emails := strings.Split(record[0], ";")
		for _, email := range emails {
			email = strings.TrimSpace(email)
			if email != "" {
				validEmailsChan <- EmailWithLine{Email: email, LineNumber: lineNumber}
			}
		}
		lineNumber++
	}
	close(validEmailsChan)

	// Waiting for all emails to be processed.
	<-doneChan

	// Now write the valid emails to the output CSV file.
	outputFile, err := os.Create(outputPath)
	if err != nil {
		fmt.Println("Error creating an output file:", err)
		return
	}
	defer outputFile.Close()

	writer := csv.NewWriter(outputFile)
	// writer.Comma = ';'
	defer writer.Flush()

	writer.Write([]string{"EMAIL" + ";"})
	for _, validEmail := range validEmails {
		if err := writer.Write([]string{validEmail + ";"}); err != nil {
			log.Printf("Error writing to CSV: %v\n", err)
			return
		}
	}
}
