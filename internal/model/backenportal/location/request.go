package location

type LocationReq struct {
	Name       string `validate:"required,oneof=provinces cities districts"`
	ProvinceId string `validate:"required_if=Name cities,omitempty,numeric"`
	CityId     string `validate:"required_if=Name districts,omitempty,numeric"`
}
