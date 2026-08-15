package withdrawal_test

import (
	"testing"
	"time"

	. "github.com/paper-indonesia/pivot-backoffice/internal/model/withdrawal"

	"github.com/stretchr/testify/assert"
)

func TestHashFilterKey(t *testing.T) {
	request := &WithdrawalListRequest{
		MerchantId: "6ddea03f-1130-458c-99b1-bd48a7b22854",
		StartDate:  time.Date(2024, 10, 31, 17, 0, 0, 0, time.UTC),
		EndDate:    time.Date(2024, 11, 14, 16, 59, 59, 0, time.UTC),
	}

	assert.Equal(t, "939678511d01bfcd707294c7ec3282e8615e7dab44d5ec13875ffab28f74bdee", request.HashFilterKey(request.EndDate))
}
