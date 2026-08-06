package pdf

import (
	"bytes"
	"context"
	"html/template"
	"io"
	"strings"

	"github.com/SebastiaanKlippert/go-wkhtmltopdf"
)

type IRequestPdf interface {
	GeneratePDF(ctx context.Context, templateFileName string, data interface{}) error
}

type RequestPdf struct {
	size        string
	orientation string
	output      io.Writer
}

type OptionFunc func(*RequestPdf)

// new request to pdf function
func NewRequestPdf(size, orientation string, opts ...OptionFunc) IRequestPdf {
	r := &RequestPdf{
		size:        size,
		orientation: orientation,
	}

	for _, opt := range opts {
		opt(r)
	}

	return r
}

func WithOutput(output io.Writer) OptionFunc {
	return func(pdf *RequestPdf) {
		pdf.output = output
	}
}

// generate pdf function
func (r *RequestPdf) GeneratePDF(ctx context.Context, templateFileName string, data interface{}) error {
	// Parsing template file
	t, err := template.ParseFiles(templateFileName)
	if err != nil {
		return err
	}
	buf := new(bytes.Buffer)
	if err = t.Execute(buf, data); err != nil {
		return err
	}

	// Create PDF instance
	pdfg, err := wkhtmltopdf.NewPDFGenerator()
	if err != nil {
		return err
	}

	pdfg.AddPage(wkhtmltopdf.NewPageReader(strings.NewReader(buf.String())))
	pdfg.PageSize.Set(r.size)
	pdfg.Orientation.Set(r.orientation)
	pdfg.Dpi.Set(300)

	if r.output != nil {
		pdfg.SetOutput(r.output)
	}

	if err = pdfg.CreateContext(ctx); err != nil {
		return err
	}

	return nil
}
