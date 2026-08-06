package shortlink

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	errPkg "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
)

func (c *ShortLinkController) GetByCode(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	ctx, segment := otelTracer.Start(ctx, "port/http/controller/v1/shortLink/GetByCode")
	defer segment.End()

	var err error
	code := chi.URLParam(r, "code")
	if code == "" {
		http.Redirect(w, r, c.config.ShortLinkRedirection.InvalidURL, http.StatusSeeOther)
		return
	}

	shortLink, err := c.shortLinkSvc.GetByCode(ctx, code)
	if err != nil {
		errType, _ := errPkg.ExtractError(err)
		if errType == response.HttpErrNotFound {
			http.Redirect(w, r, c.config.ShortLinkRedirection.InvalidURL, http.StatusSeeOther)
			return
		}

		// Tech Debt: should return to error page for internal error
		http.Redirect(w, r, c.config.ShortLinkRedirection.InvalidURL, http.StatusSeeOther)
		return
	}

	http.Redirect(w, r, shortLink.DestinationURL, http.StatusFound)
}
