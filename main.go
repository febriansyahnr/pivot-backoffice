package main

import (
	"github.com/paper-indonesia/pivot-backoffice/cmd"
	"time"
	_ "time/tzdata"
)

//	@title			Backend Portal API
//	@version		1.0
//	@description	To support payment gateway needs, we prepare technical documents on the backend-portal.
//	@description 	Where the backend-portal is intended to serve services to OpenAPI and MerchantPortal,
//	@description 	apart from that this service is integrated with Orchestrator, Core Processor
//	@description 	(SNAP Core Processor and Credit Card Processor)
//	@termsOfService	http://swagger.io/terms/

//	@contact.name	API Support
//	@contact.url	http://www.swagger.io/support
//	@contact.email	support@swagger.io

//	@license.name	Apache 2.0
//	@license.url	http://www.apache.org/licenses/LICENSE-2.0.html

// 	@securityDefinitions.apikey Bearer
// 	@in header
// 	@name Authorization
// 	@description Type "Bearer" followed by a space and JWT token.

// @host			localhost
// @Schemes 		https
// @BasePath		/
func main() {
	loc, err := time.LoadLocation("UTC")
	if err != nil {
		panic(err)
	}

	time.Local = loc

	_ = cmd.Execute()
}
