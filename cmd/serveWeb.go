package cmd

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"sync/atomic"
	"time"

	"github.com/paper-indonesia/pdk/v2/logger"
	"github.com/paper-indonesia/pivot-backoffice/config"
	webhandler "github.com/paper-indonesia/pivot-backoffice/handler/web"
	dashboardhandler "github.com/paper-indonesia/pivot-backoffice/handler/web/dashboard"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/application"
	disbursementService "github.com/paper-indonesia/pivot-backoffice/internal/service/v1/disbursement"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(servewebcmd)
}

var servewebcmd = &cobra.Command{
	Use:   "serveWeb",
	Short: "Start Web server",
	Long:  `Start Back Office`,
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.Background()
		cfg, secret, err := config.LoadConfig(cfgFile, scrtFile)
		if err != nil {
			fmt.Printf("Unable to load configuration and secret: %v", err)
			panic(err)
		}

		app := application.NewApplication(cfg, secret, ctx)
		defer app.Recover()
		defer app.Closer()

		disbursementSvc := disbursementService.New(
			cfg, app.GetLogger(), app.GetMerchantRepository(), app.GetDisbursementRepository(), app.GetSnapCoreRepository(), app.GetBankAccountRepository(),
			disbursementService.WithStatusHistoriesRepository(app.GetStatusHistoriesRepository()),
		)
		defer disbursementSvc.WPRelease()

		dasboardHandler := dashboardhandler.NewDashboard(
			dashboardhandler.WithDisbursementService(disbursementSvc),
		)

		web := webhandler.NewWebServer("0.0.0.0:8080", app, dasboardHandler)

		ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
		defer cancel()

		exitCode := atomic.Int32{}

		stopped := make(chan struct{})
		go func() {
			defer close(stopped)

			app.GetLogger().Info(ctx, "starting server", logger.String("address", "0.0.0.0:8080"))

			if err := web.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
				app.GetLogger().Error(ctx, "server failed", logger.String("error", err.Error()))
				exitCode.Store(1)
				cancel()
			}
		}()

		<-ctx.Done()
		app.GetLogger().Info(ctx, "Shutting down gracefully...")

		if err := web.Stop(10 * time.Second); err != nil {
			app.GetLogger().Error(ctx, "Server failed to shutdown gracefully", logger.String("error", err.Error()))
			os.Exit(1)
		}

		<-stopped

		if code := exitCode.Load(); code != 0 {
			os.Exit(int(code))
		}
	},
}
