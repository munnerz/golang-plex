package plex

import (
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"net/http"
)

const (
	accountInfoURL = "https://plex.tv/users/account"
	authURL        = "https://plex.tv/users/sign_in.xml"
)

type User struct {
	ID       string `xml:"id,attr"`
	Username string `xml:"username,attr"`
	Email    string `xml:"email,attr"`
	Thumb    string `xml:"thumb"`
	Title    string `xml:"title"`
	Password string
	Token    string `xml:"authenticationToken"`
}

func (u *User) authenticate() error {
	req, err := http.NewRequest("POST", authURL, nil)

	if err != nil {
		return err
	}

	req.SetBasicAuth(u.Username, u.Password)

	addPlexHeaders(req)

	client := http.DefaultClient

	resp, err := client.Do(req)

	if err != nil {
		return err
	}

	if resp.StatusCode != 200 {
		return errors.New(http.StatusText(resp.StatusCode))
	}

	decoder := xml.NewDecoder(io.ReadCloser(resp.Body))

	err = decoder.Decode(u)

	if err != nil {
		return err
	}

	return nil
}

func (u *User) loadAccountDetails() error {
	req, err := http.NewRequest("GET", accountInfoURL, nil)

	if err != nil {
		return err
	}

	req.URL.RawQuery = fmt.Sprintf("X-Plex-Token=%s", u.Token)

	client := http.DefaultClient

	resp, err := client.Do(req)

	if err != nil {
		return err
	}

	if resp.StatusCode != 200 {
		return errors.New(http.StatusText(resp.StatusCode))
	}

	decoder := xml.NewDecoder(io.ReadCloser(resp.Body))

	err = decoder.Decode(u)

	if err != nil {
		return err
	}

	return nil
}

func (u *User) fetchUserDetails() error {
	if u.Token == "" {
		err := u.authenticate()

		if err != nil {
			return err
		}
	} else {
		err := u.loadAccountDetails()

		if err != nil {
			return err
		}
	}

	return nil
}

func NewUser(args ...func(*User) (*User, error)) (*User, error) {
	u := new(User)

	for _, f := range args {
		_, err := f(u)

		if err != nil {
			return u, err
		}
	}

	err := u.fetchUserDetails()

	if err != nil {
		return u, err
	}

	return u, nil
}

func SetUsername(username string) func(*User) (*User, error) {
	return func(u *User) (*User, error) {
		u.Username = username
		return u, nil
	}
}

func SetPassword(password string) func(*User) (*User, error) {
	return func(u *User) (*User, error) {
		u.Password = password
		return u, nil
	}
}

func SetToken(token string) func(*User) (*User, error) {
	return func(u *User) (*User, error) {
		u.Token = token
		return u, nil
	}
}
