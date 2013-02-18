// Package user implements a simple user-management package.
package user

import (
	"crypto/sha1"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"
	"uuid"
)

var InvalidNicknameError = errors.New("Invalid nickname.")
var InvalidPasswordError = errors.New("Invalid password.")

var passwordSalt string
var userFilename string
var users []*User
var usersByNickname map[string]*User
var usersById map[string]*User

type User struct {
	Id string 	`json:"id"`
	Nickname string `json:"nickname"`
	Pwhash string `json:"pwhash"`
}

// Init initializes the user module.
func Init(salt string, filename string) {
	userFilename = filename
	passwordSalt = salt

	users = make([]*User, 0)
	usersByNickname = make(map[string]*User)
	usersById = make(map[string]*User)
}

// New creates a new user, with the given nickname and password.  It returns
// the user created, or an error if the nickname already exists.  The nickname
// is canonicalized, and the password is salted and hashed.
func New(nickname string, password string) (*User, error) {

	if passwordSalt == "" {
		return nil, errors.New("Must call Init before creating any user.")
	}

	id, err := uuid.GenUUID()
	if err != nil {
		return nil, err
	}

	u := &User {
		Id: id,
		Nickname: nickname,
		Pwhash: hashPassword(canonicalizePassword(password))}

	key := canonicalizeNickname(u.Nickname)
	if usersByNickname[key] != nil {
		return nil, errors.New("Duplicate nickname.")
	}

	users = append(users, u)
	usersByNickname[key] = u
	usersById[u.Id] = u

	return u, nil
}

// Login validates a user login and returns the User.
func Login(nickname, password string) (*User, error) {
	u := LookupByNickname(nickname)
	if u == nil {
		return nil, InvalidNicknameError
	}
	if u.Pwhash != hashPassword(canonicalizePassword(password)) {
		return nil, InvalidPasswordError
	}
	return u, nil
}

// LookupById returns the User with the given ID, or nil if none exists.
func LookupById(id string) *User {
	return usersById[id]
}

// LookupByNickname returns the User with the given nickname, or nil if none exists.
func LookupByNickname(nickname string) *User {
	return usersByNickname[canonicalizeNickname(nickname)]
}

// canonicalizeNickname strips non-alpha-numeric characters from
// the nickname and converts it to lower case.
func canonicalizeNickname(nickname string) string {
	re := regexp.MustCompile("[^A-Za-z0-9]*")
	s := re.ReplaceAllLiteralString(nickname, "")
	return strings.ToLower(s)
}

// canonicalizePassword strips whitespace from the password.
func canonicalizePassword(password string) string {
	re := regexp.MustCompile("[\\s]*")
	s := re.ReplaceAllLiteralString(password, "")
	return s
}

// hashPassword salts and hashes a password.
func hashPassword(password string) string {

	h := sha1.New()
	io.WriteString(h, passwordSalt)
	io.WriteString(h, password)

	return fmt.Sprintf("%x", h.Sum(nil))
}

// LoadUsers loads users saved in userFilename.
func LoadUsers() error {
	if userFilename == "" {
		return errors.New("Must call Init to set user filename first.")
	}
	var err error
	var f *os.File
	var fi os.FileInfo
	var b []byte

	fi, err = os.Stat(userFilename)
	if err != nil {
		return err
	}

	b = make([]byte, fi.Size())
	f, err = os.Open(userFilename)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = f.Read(b)
	if err != nil {
		return err
	}

	err = json.Unmarshal(b, &users)
	if err != nil {
		return err
	}

	usersByNickname = make(map[string] *User)
	usersById = make(map[string] *User)
	for _, u := range users {
		usersByNickname[canonicalizeNickname(u.Nickname)] = u
		usersById[u.Id] = u
	}

	return nil

}

// SaveUsers writes the users out to userFilename.
func SaveUsers() error {
	if userFilename == "" {
		return errors.New("Must call Init to set user filename first.")
	}
	if users == nil {
		return errors.New("Must load or initialize users first.")
	}
	var b []byte
	var f *os.File
	var err error

	b, err = json.Marshal(users)
	if err != nil {
		return err
	}
	f, err = os.Create(userFilename)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = f.Write(b)
	if err != nil {
		return err
	}
	return nil
}
