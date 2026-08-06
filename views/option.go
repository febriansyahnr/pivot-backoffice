package views

import "github.com/a-h/templ"

type SwapOption string

const (
	SwapOuterHTML SwapOption = "outerHTML"
	SwapInnerHTML SwapOption = "innerHTML"
)

type HXOption struct {
	Method    string
	URL       string
	Trigger   string
	Target    string
	Indicator string
	Swap      SwapOption
	Headers   map[string]string
	Include   string
}

func (h HXOption) ToAttributes() templ.Attributes {
	attrs := make(templ.Attributes)
	switch h.Method {
	case "GET":
		attrs["hx-get"] = h.URL
	case "POST":
		attrs["hx-post"] = h.URL
	case "PUT":
		attrs["hx-put"] = h.URL
	case "DELETE":
		attrs["hx-delete"] = h.URL
	case "PATCH":
		attrs["hx-patch"] = h.URL
	}

	if h.Trigger != "" {
		attrs["hx-trigger"] = h.Trigger
	}
	if h.Target != "" {
		attrs["hx-target"] = h.Target
	}
	if h.Indicator != "" {
		attrs["hx-indicator"] = h.Indicator
	}
	if h.Swap != "" {
		attrs["hx-swap"] = string(h.Swap)
	}
	if len(h.Headers) > 0 {
		attrs["hx-headers"] = h.Headers
	}
	if h.Include != "" {
		attrs["hx-include"] = h.Include
	}
	return attrs
}
