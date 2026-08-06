package orchestratorController

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/paper-indonesia/pivot-backoffice/pkg/util"

	"github.com/stretchr/testify/assert"
	ffclient "github.com/thomaspoignant/go-feature-flag"
	"github.com/thomaspoignant/go-feature-flag/retriever/fileretriever"
)

func TestMain(m *testing.M) {
	cwd, _ := os.Getwd()
	projectRoot, _ := util.FindProjectRoot(cwd, "backend-portal")
	targetPath := filepath.Join(projectRoot, "test", "consul", "backend-portal", "feature-flag.yaml")

	_ = ffclient.Init(ffclient.Config{
		Retriever:    &fileretriever.Retriever{Path: targetPath},
		DataExporter: ffclient.DataExporter{},
	})
	defer ffclient.Close()

	m.Run()
}

func TestSortFieldForGetList(t *testing.T) {
	tests := []struct {
		input  string
		ok     bool
		sorted string
	}{
		{
			input:  "-date",
			ok:     true,
			sorted: "t.updated_at DESC",
		},
		{
			input:  "date",
			ok:     true,
			sorted: "t.updated_at",
		},
		{
			input:  "-beneficiaryAccountName",
			ok:     true,
			sorted: "d.beneficiary_account_name DESC",
		},
		{
			input:  "beneficiaryAccountName",
			ok:     true,
			sorted: "d.beneficiary_account_name",
		},
		{
			input:  "-amount",
			ok:     true,
			sorted: "((-1*t.debit)+t.credit) DESC",
		},
		{
			input:  "amount",
			ok:     true,
			sorted: "((-1*t.debit)+t.credit)",
		},
		{
			input:  "id",
			ok:     false,
			sorted: "",
		},
	}
	for _, test := range tests {
		val, ok := sortFieldForGetList(test.input)

		assert.Equal(t, test.sorted, val)
		assert.Equal(t, test.ok, ok)
	}
}
