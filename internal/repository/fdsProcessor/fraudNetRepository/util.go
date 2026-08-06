package fraudnetrepository

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"

	fdscommon "github.com/paper-indonesia/pivot-backoffice/internal/model/fdsProcessor/fdsCommon"
	fraudnetmodel "github.com/paper-indonesia/pivot-backoffice/internal/model/fdsProcessor/fraudNet"
)

func CreateBasicAuth(username, password string) string {
	credentials := fmt.Sprintf("%s:%s", username, password)
	encoded := base64.StdEncoding.EncodeToString([]byte(credentials))

	return "Basic " + encoded
}

func ValidateHttpResponse[T any](resp io.ReadSeeker) (*T, error) {
	var response T

	decoder := json.NewDecoder(resp)

	err := decoder.Decode(&response)
	if err != nil {
		return &response, err
	}

	resp.Seek(0, io.SeekStart) // restart reading buffer from the beginning

	decoder.Decode(&response)

	return &response, nil
}

func CheckRequestMapToCommonResponse(fraudNetResponse *fraudnetmodel.MarketplaceCheckResponse) *fdscommon.CheckResponse {
	if fraudNetResponse == nil {
		return nil
	}

	tags := make([]fdscommon.CheckTags, 0, len(fraudNetResponse.Data.Tags))
	for _, tag := range fraudNetResponse.Data.Tags {
		tags = append(tags, fdscommon.CheckTags{
			ID:        tag.ID,
			Action:    tag.Action,
			Name:      tag.Name,
			Source:    tag.Source,
			Type:      tag.Type,
			State:     tag.State,
			Weight:    tag.Weight,
			RiskScore: tag.RiskScore,
			RiskGroup: tag.RiskGroup,
			Link:      tag.Link,
		})
	}

	return &fdscommon.CheckResponse{
		Success: fraudNetResponse.Success,
		Code:    fraudNetResponse.Code,
		Source:  fraudNetResponse.Source,
		Message: fraudNetResponse.Message,
		Data: fdscommon.CheckData{
			ID:        fraudNetResponse.Data.ID,
			Timer:     fraudNetResponse.Data.Timer,
			RiskScore: fraudNetResponse.Data.RiskScore,
			RiskGroup: fraudNetResponse.Data.RiskGroup,
			Link:      fraudNetResponse.Data.Link,
			Tags:      tags,
		},
	}
}

func UpdateRequestMapToCommonResponse(fraudNetResponse *fraudnetmodel.MarketplaceUpdateResponse) *fdscommon.UpdateResponse {
	if fraudNetResponse == nil {
		return nil
	}

	return &fdscommon.UpdateResponse{
		Success: fraudNetResponse.Success,
		Code:    fraudNetResponse.Code,
		Source:  fraudNetResponse.Source,
		Message: fraudNetResponse.Message,
		Data: fdscommon.UpdateData{
			ID:    fraudNetResponse.Data.ID,
			Link:  fraudNetResponse.Data.Link,
			Timer: fraudNetResponse.Data.Timer,
		},
	}
}
