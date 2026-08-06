package activityModel

import (
	"encoding/json"
	"fmt"
	"time"
)

const (
	COLLECTION_NAME = "pg_activities"
)

type Activity struct {
	ID          string          `json:"id" bson:"id" example:"b3b3b3b3-3b3b-3b3b-3b3b-3b3b3b3b3b3b"`
	MerchantID  string          `json:"merchantId" bson:"merchant_id" example:"b3b3b3b3-3b3b-3b3b-3b3b-3b3b3b3b3b3b"`
	UserID      *string         `json:"userId" bson:"user_id" example:"b3b3b3b3-3b3b-3b3b-3b3b-3b3b3b3b3b3b"`
	Tag         string          `json:"tag" bson:"tag" example:"account"`
	Activity    string          `json:"activity" bson:"activity" example:"User login"`
	ServiceName string          `json:"serviceName" bson:"service_name" example:"User login"`
	Parameter   *map[string]any `json:"parameter" bson:"parameter"`
	CreatedAt   time.Time       `json:"createdAt" bson:"created_at" example:"2024-01-01T00:00:00Z"`
	UpdatedAt   time.Time       `json:"updatedAt" bson:"updated_at" example:"2024-01-01T00:00:00Z"`
}

type ActivityDTO struct {
	ID          string    `json:"id" db:"id"`
	MerchantID  string    `json:"merchantId" db:"merchant_id"`
	UserID      *string   `json:"userId" db:"user_id"`
	Tag         string    `json:"tag" db:"tag"`
	Activity    string    `json:"activity" db:"activity"`
	ServiceName string    `json:"serviceName" db:"service_name"`
	Parameter   string    `json:"parameter" db:"parameter"`
	CreatedAt   time.Time `json:"createdAt" db:"created_at"`
	UpdatedAt   time.Time `json:"updatedAt" db:"updated_at"`
}

func (a *Activity) ToDTO() *ActivityDTO {
	parameterJson, _ := json.Marshal(a.Parameter)
	parameter := string(parameterJson)

	return &ActivityDTO{
		ID:          a.ID,
		MerchantID:  a.MerchantID,
		UserID:      a.UserID,
		Tag:         a.Tag,
		Activity:    a.Activity,
		ServiceName: a.ServiceName,
		Parameter:   parameter,
		CreatedAt:   a.CreatedAt,
		UpdatedAt:   a.UpdatedAt,
	}
}

func (a *Activity) FromDTO(dto *ActivityDTO) {
	a.ID = dto.ID
	a.MerchantID = dto.MerchantID
	a.UserID = dto.UserID
	a.Tag = dto.Tag
	a.Activity = dto.Activity
	a.ServiceName = dto.ServiceName
	a.CreatedAt = dto.CreatedAt
	a.UpdatedAt = dto.UpdatedAt

	var parameter map[string]interface{}
	err := json.Unmarshal([]byte(dto.Parameter), &parameter)
	if err != nil {
		fmt.Println("[ActivityModel@FromDTO] Error: ", err)
		return
	}
	a.Parameter = &parameter
}
