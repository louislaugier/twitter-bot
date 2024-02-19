package email

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/mail"
	"net/smtp"
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"golang.org/x/net/proxy"
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
	// Configuration for your proxy.
	proxyAddr := "68.71.254.6:4145" // Replace with your proxy's IP and port.

	// Establishing a connection to your proxy.
	dialer, err := proxy.SOCKS5("tcp", proxyAddr, nil, proxy.Direct)
	if err != nil {
		return nil, fmt.Errorf("error connecting to proxy: %v", err)
	}

	// Get the MX records for the host.
	mxRecords, err := net.LookupMX(host)
	if err != nil || len(mxRecords) == 0 {
		return nil, fmt.Errorf("LookupMX failed: %v", err)
	}

	// Attempt to connect to the SMTP server via the proxy.
	var smtpClient *smtp.Client
	for _, mxRecord := range mxRecords {
		serverAddr := strings.TrimSuffix(mxRecord.Host, ".") + ":" + strconv.Itoa(SmtpPort)
		log.Printf("Trying to connect to MX record: %s\n", serverAddr)
		conn, err := dialer.Dial("tcp", serverAddr)
		if err != nil {
			log.Println("abc", serverAddr)
			for _, v := range mxRecords {
				log.Println("xyz", v.Host)
			}
			log.Printf("Unable to dial to SMTP server via proxy: %s, error: %v\n", serverAddr, err)
			continue
		}

		log.Println("Dial to SMTP server successful, attempting to create SMTP client...")
		smtpClient, err = smtp.NewClient(conn, mxRecord.Host)
		if err != nil {
			conn.Close()
			log.Printf("SMTP client initialization failed for server: %s, error: %v\n", serverAddr, err)
			continue
		}
		log.Println("SMTP client successfully created")
		break
	}

	if smtpClient == nil {
		return nil, fmt.Errorf("unable to connect to any MX hosts")
	}

	return smtpClient, nil
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

	err = client.Hello("mail.example.com")
	if err != nil {
		return err
	}

	// err = client.Mail(email)
	// if err != nil {
	// 	return err
	// }

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

// docker run -p 8080:8080 reacherhq/backend:latest
func ValidateEmailAddressFromAPI(email string) error {
	r, err := http.Post("http://localhost:8080/v0/check_email", "application/json", bytes.NewBuffer([]byte(fmt.Sprintf(`
		{
			"to_email": "%s",
    		"from_email": "hello@tweeter-id.com",
    		"hello_name": "tweeter-id.com"
		}
	`, email))))
	if err != nil {
		log.Println("Error requesting API:", err)
		return err
	}
	defer r.Body.Close()

	responseMap := map[string]interface{}{}
	err = json.NewDecoder(r.Body).Decode(&responseMap)
	if err != nil {
		log.Println("Error decoding API response:", err)
		return err
	}

	reachability, ok := responseMap["is_reachable"]
	if !ok {
		return errors.New("'is_reachable' field not present in API response")
	}

	if ok && (reachability == "safe" || reachability == "risky") {
		return nil
	}

	return fmt.Errorf("reachability %s", reachability)
}

// run plusieurs fois, check output count (different?)
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
				if err := ValidateEmailAddressFromAPI(ewl.Email); err == nil {
					mu.Lock()
					log.Printf("Valid email on line %d: %s.\n", ewl.LineNumber, ewl.Email)
					validEmails = append(validEmails, ewl.Email)
					mu.Unlock()
				} else if ewl.Email != "E-MAIL" {
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
