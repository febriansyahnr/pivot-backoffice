package credential

import "time"

type ClientSecret struct {
	Secret        string    `db:"secret"`
	SecretVersion uint      `db:"secret_version"`
	UpdatedAt     time.Time `db:"updated_at"`
}

func (c *ClientSecret) ToResponse() *ClientSecretResp {
	return &ClientSecretResp{
		Secret:     c.Secret,
		LastUpdate: c.UpdatedAt,
		Time:       time.Now().UTC().Unix(),
	}
}
