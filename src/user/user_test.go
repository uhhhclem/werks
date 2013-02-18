package user

import (
	"testing"
)

func TestNew(t *testing.T) {
	var u *User
	var err error

	u, err = New("nickname", "password")
	if err != nil {
		t.Errorf("%s", err)
	}
	t.Logf("ID: %s, Nickname: %s, Pwhash: %s", u.Id, u.Nickname, u.Pwhash)
	t.Fail()
}
