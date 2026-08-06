package httputil

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetPaginationQuery(t *testing.T) {

	t.Run("Pagination Query", func(t *testing.T) {
		r, _ := http.NewRequest("GET", "http://localhost:8080?page=10&perPage=11", nil)
		page, perPage := GetPaginationQuery(r)
		assert.Equal(t, page, int64(10))
		assert.Equal(t, perPage, int64(11))
	})

}
