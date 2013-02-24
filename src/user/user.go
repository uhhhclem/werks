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

// Users is a persistable collection of users.
type Users struct {
  passwordSalt string
  userFilename string
  users []*User
  usersByNickname map[string]*User
  usersById map[string]*User
  usersByToken map[string]*User
}

// User is an individual user.
type User struct {
	Id string 	`json:"id"`
	Nickname string `json:"nickname"`
	Pwhash string `json:"pwhash"`
	Token string `json:"-"`
}

// Init creates a new Users structure.
func Init(salt string, filename string) *Users {
	u := new(Users)
	u.userFilename = filename
	u.passwordSalt = salt

	u.users = make([]*User, 0)
	u.usersByNickname = make(map[string]*User)
	u.usersById = make(map[string]*User)
	u.usersByToken = make(map[string]*User)

	return u
}

// Register creates a new user, with the given nickname and password.  It returns
// the user created, or an error if the nickname already exists.  The nickname
// is canonicalized, and the password is salted and hashed.
func (s *Users) Register(nickname string, password string) (*User, error) {

	if s.passwordSalt == "" {
		return nil, errors.New("No password salt found.")
	}

	id, err := uuid.GenUUID()
	if err != nil {
		return nil, err
	}


	u := &User {
		Id: id,
		Nickname: nickname,
		Pwhash: s.hashPassword(canonicalizePassword(password))}

	u.Token, err = uuid.GenUUID()
	if err != nil {
		return nil, err
	}

	key := canonicalizeNickname(u.Nickname)
	if s.usersByNickname[key] != nil {
		return nil, errors.New("Duplicate nickname.")
	}

	s.users = append(s.users, u)
	s.usersByNickname[key] = u
	s.usersById[u.Id] = u
	s.usersByToken[u.Token] = u

	return u, nil
}

// Login validates a user login and returns the User.  If a user is
// logged in, he gets assigned a Token.
func (s *Users) Login(nickname, password string) (*User, error) {
	u := s.LookupByNickname(nickname)
	if u == nil {
		return nil, InvalidNicknameError
	}
	if u.Pwhash != s.hashPassword(canonicalizePassword(password)) {
		return nil, InvalidPasswordError
	}

	// login succeeded, so assign token if it hasn't already been assigned.
	if u.Token == "" {
		var err error
		u.Token, err = uuid.GenUUID()
		if err != nil {
			return nil, err
		}
		s.usersByToken[u.Token] = u
	}
	return u, nil
}

// LookupById returns the User with the given ID, or nil if none exists.
func (s *Users) LookupById(id string) *User {
	return s.usersById[id]
}

// LookupByNickname returns the User with the given nickname, or nil if none exists.
func (s *Users) LookupByNickname(nickname string) *User {
	return s.usersByNickname[canonicalizeNickname(nickname)]
}

func (s *Users) LookupByToken(token string) *User {
	return s.usersByToken[token]
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
func (s *Users) hashPassword(password string) string {

	h := sha1.New()
	io.WriteString(h, s.passwordSalt)
	io.WriteString(h, password)

	return fmt.Sprintf("%x", h.Sum(nil))
}

// LoadUsers loads users saved in userFilename.
func (s *Users) LoadUsers() error {
	if s.userFilename == "" {
		return errors.New("Must set user filename first.")
	}
	var err error
	var f *os.File
	var fi os.FileInfo
	var b []byte

	fi, err = os.Stat(s.userFilename)
	if err != nil {
		return err
	}

	b = make([]byte, fi.Size())
	f, err = os.Open(s.userFilename)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = f.Read(b)
	if err != nil {
		return err
	}

	err = json.Unmarshal(b, &s.users)
	if err != nil {
		return err
	}

	s.usersByNickname = make(map[string] *User)
	s.usersById = make(map[string] *User)
	for _, u := range s.users {
		s.usersByNickname[canonicalizeNickname(u.Nickname)] = u
		s.usersById[u.Id] = u
	}

	return nil

}

// SaveUsers writes the users out to userFilename.
func (s *Users) SaveUsers() error {
	if s.userFilename == "" {
		return errors.New("Must set user filename first.")
	}
	if s.users == nil {
		return errors.New("Must load or initialize users first.")
	}
	var b []byte
	var f *os.File
	var err error

	b, err = json.Marshal(s.users)
	if err != nil {
		return err
	}
	f, err = os.Create(s.userFilename)
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
