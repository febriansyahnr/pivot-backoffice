package location

import "time"

type Province struct {
	Id        uint16    `json:"id" db:"id"`
	Name      string    `json:"name" db:"name"`
	CreatedAt time.Time `json:"-" db:"created_at"`
}

type City struct {
	Id         uint16    `json:"id" db:"id"`
	ProvinceId uint16    `json:"-" db:"province_id"`
	Name       string    `json:"name" db:"name"`
	CreatedAt  time.Time `json:"-" db:"created_at"`
}

type District struct {
	Id        uint16    `json:"id" db:"id"`
	CityId    uint16    `json:"-" db:"city_id"`
	Name      string    `json:"name" db:"name"`
	CreatedAt time.Time `json:"-" db:"created_at"`
}
