package prizekey

import (
	"net/url"
	"strconv"
)

const (
	EmailLockPrefix     = "game:v2:prize:email:"
	ParticipationPrefix = "game:v2:prize:participation:"
	ActivationAPIURL    = "https://game.go-furry.com/api/v2/game/prizes/participation/activation"
)

func EmailLockKey(id int64, email string) string {
	return EmailLockPrefix + strconv.FormatInt(id, 10) + ":" + email
}

func EmailLockKeyString(id string, email string) string {
	return EmailLockPrefix + id + ":" + email
}

func ParticipationKey(id int64, key string) string {
	return ParticipationPrefix + strconv.FormatInt(id, 10) + ":" + key
}

func ParticipationKeyString(id string, key string) string {
	return ParticipationPrefix + id + ":" + key
}

func ActivationLink(id int64, key string) string {
	u, err := url.Parse(ActivationAPIURL)
	if err != nil {
		return ActivationAPIURL
	}
	q := u.Query()
	q.Set("id", strconv.FormatInt(id, 10))
	q.Set("key", key)
	u.RawQuery = q.Encode()
	return u.String()
}
