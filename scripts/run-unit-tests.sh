set -o pipefail

TEST_PKGS=$(go list ./... | grep -v -E '/mocks|/pkg|/cmd|/config|/docs|/constant|/test|/scripts|/internal/model/proto')

go test -race -covermode=atomic -coverprofile=origin-coverage.out $TEST_PKGS

grep -v -E 'port/http/route\.go|port/cron/.*\.go|/type\.go' origin-coverage.out > coverage.out

rm ./origin-coverage.out
