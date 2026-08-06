package dashboardhandler

import (
	"net/http"

	"github.com/a-h/templ"
	"github.com/go-chi/chi/v5"
	webhandler "github.com/paper-indonesia/pivot-backoffice/handler/web"
	"github.com/paper-indonesia/pivot-backoffice/internal/service"
	"github.com/paper-indonesia/pivot-backoffice/views/dashboard"
	"github.com/paper-indonesia/pivot-backoffice/views/errorpages"
)

type DashboardComponent struct {
	Dashboard templ.Component
	NotFound  templ.Component
}

type Dashboard struct {
	Component       DashboardComponent
	disbursementSvc service.IDisbursementService
}

func NewDashboard(opts ...dashboardOption) *Dashboard {
	d := &Dashboard{
		Component: DashboardComponent{
			Dashboard: dashboard.Dashboard("payout"),
			NotFound:  errorpages.NotFound(),
		},
	}
	for _, opt := range opts {
		opt(d)
	}
	return d
}

func (s *Dashboard) Root() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			webhandler.WithBase(s.Component.NotFound, "Not found", "", templ.WithStatus(http.StatusNotFound)).ServeHTTP(w, r)
			return
		}
		http.Redirect(w, r, "/dashboard", http.StatusTemporaryRedirect)
	})
}

func (s *Dashboard) Dashboard() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tab := r.URL.Query().Get("tab")
		switch tab {
		case "payout":
			s.Component.Dashboard = dashboard.Dashboard("payout")
		case "payment":
			s.Component.Dashboard = dashboard.Dashboard("payment")
		default:
			s.Component.Dashboard = dashboard.Dashboard("payout")
		}
		webhandler.WithBase(s.Component.Dashboard, "Dashboard", "").ServeHTTP(w, r)
	})
}

func (s *Dashboard) Register(mux *chi.Mux) {
	mux.Get("/*", s.Root().ServeHTTP)
	mux.Get("/dashboard", s.Dashboard().ServeHTTP)
	mux.Post("/dashboard/search-payout", s.SearchPayout().ServeHTTP)
}
