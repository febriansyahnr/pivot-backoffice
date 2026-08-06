package advanceairepository

import (
	"encoding/json"
	"io"
)

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
