package secrets

import (
	"errors"
	"fmt"
	"github.com/tlopo-go/cookie-exporter/cmd"
	"regexp"
	"strings"
)

type Credentials struct {
	Service  string
	Account  string
	Password string
}

func Set(creds Credentials) error {
	if Exist(creds.Service) {
		if err := Delete(creds.Service); err != nil {
			return err
		}
	}

	command := fmt.Sprintf(
		`security add-generic-password -s "%s" -a "%s" -w "%s"`,
		creds.Service,
		creds.Account,
		creds.Password)

	c := cmd.New(cmd.Options{Command: command})
	c.Run()

	if c.Err != nil {
		return errors.New(fmt.Sprintf("failed to set credentials for %+v", creds))
	}

	return nil
}

func Get(service string) (Credentials, error) {
	var acct string
	command := fmt.Sprintf(`security find-generic-password -s "%s" -w`, service)
	c := cmd.New(cmd.Options{Command: command})
	c.Run()

	if c.Err != nil {
		return Credentials{}, errors.New(fmt.Sprintf("failed to get password for service %s", service))
	}

	passwd := strings.TrimSpace(c.Stdout)

	command = fmt.Sprintf(`security find-generic-password -s "%s"`, service)
	c = cmd.New(cmd.Options{Command: command})
	c.Run()

	if c.Err != nil {
		return Credentials{}, errors.New(fmt.Sprintf("failed to get password for service %s", service))
	}

	regex, _ := regexp.Compile("acct")

	for _, str := range strings.Split(c.Stdout, "\n") {
		if regex.MatchString(str) {
			str = strings.Join(strings.Split(str, "=")[1:], " ")
			acct = regexp.MustCompile(`^"|"`).ReplaceAllString(str, "")
		}
	}

	return Credentials{service, acct, passwd}, nil
}

func Exist(service string) bool {
	command := fmt.Sprintf(`security find-generic-password -s "%s" -w`, service)
	c := cmd.New(cmd.Options{Command: command})
	c.Run()
	return c.ExitStatus == 0
}

func Delete(service string) error {
	command := fmt.Sprintf(`security delete-generic-password -s "%s" > /dev/null 2>&1`, service)
	c := cmd.New(cmd.Options{Command: command})
	c.Run()

	if c.Err != nil {
		return errors.New(fmt.Sprintf("failed to delete secret for service %s", service))
	}

	return nil
}
