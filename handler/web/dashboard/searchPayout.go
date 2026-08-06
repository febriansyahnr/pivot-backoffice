package dashboardhandler

import (
	"net/http"

	"github.com/a-h/templ"
	"github.com/paper-indonesia/pivot-backoffice/views/dashboard"
)

func (s *Dashboard) SearchPayout() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		payoutID := r.FormValue("payoutID")
		disbursement, err := s.disbursementSvc.FindByID(r.Context(), payoutID)
		if err != nil {
			templ.Handler(dashboard.ListData("Payout not found")).ServeHTTP(w, r)
			return
		}
		templ.
			Handler(dashboard.PayoutDetail(dashboard.PayoutDetailProps{
				Title:        "Payout " + disbursement.UUID,
				Disbursement: disbursement,
			})).
			ServeHTTP(w, r)
	})
}
