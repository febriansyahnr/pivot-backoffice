package httputil

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"path/filepath"
	"sync"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

var (
	once   sync.Once
	client *http.Client
)

var bufferPool = &sync.Pool{
	New: func() any {
		return new(bytes.Buffer)
	},
}

var otelTracer = otel.Tracer("HttpRequest")

// CustomTransport wraps the standard HTTP transport to add request path as tags
type CustomTransport struct {
	Transport http.RoundTripper
	Tracer    trace.Tracer
}

func (t *CustomTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	// Start a new span for the request
	ctx, span := t.Tracer.Start(req.Context(), req.Method, trace.WithAttributes(
		attribute.Key("http.url.path").String(req.URL.Path),
	))
	defer span.End()

	// Set the new context with the span
	req = req.WithContext(ctx)

	// Make the actual HTTP request using the original transport
	return t.Transport.RoundTrip(req)
}

func RequestHitAPI(ctx context.Context, method, uri string, data interface{}, headers map[string]string) ([]byte, int, error) {
	ctx, segment := otelTracer.Start(ctx, "pkg/util/http/RequestHitAPI")
	defer segment.End()

	once.Do(func() {
		client = &http.Client{}
		client.Transport = &CustomTransport{
			Transport: otelhttp.NewTransport(http.DefaultTransport),
			Tracer:    otel.Tracer("http-client"),
		}
	})

	buff := bufferPool.Get().(*bytes.Buffer)
	if data != nil {
		// Check if the data is already []byte, if so, write it directly to the buffer
		if byteData, ok := data.([]byte); ok {
			buff.Write(byteData)
		} else {
			// Otherwise, use JSON encoding for Go struct
			if err := json.NewEncoder(buff).Encode(data); err != nil {
				return nil, 0, err
			}
		}
	}
	defer func() {
		buff.Reset()
		bufferPool.Put(buff)
	}()

	request, err := http.NewRequestWithContext(ctx, method, uri, buff)
	if err != nil {
		return nil, 0, err
	}
	for k, v := range headers {
		request.Header.Add(k, v)
	}
	if val := request.Header.Get(constant.HeaderContentType); val == "" {
		request.Header.Set(constant.HeaderContentType, "application/json")
	}

	response, err := client.Do(request)
	if err != nil {
		return nil, 0, err
	}
	defer response.Body.Close()

	responseBody, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, 0, err
	}

	entryPoint, ok := ctx.Value(constant.CtxEntryPoint).(string)
	if ok && entryPoint == constant.SnapCoreProcessor {
		return responseBody, response.StatusCode, nil
	}

	if response.StatusCode >= 400 {
		errRes := map[string]interface{}{}
		if err = json.Unmarshal(responseBody, &errRes); err != nil {
			return []byte(fmt.Sprintf(`{"status":%d, "body":"%s"}`, response.StatusCode, string(responseBody))), response.StatusCode, nil
		}
	}
	return responseBody, response.StatusCode, nil
}

// RequestHitAPIWithFormData is a function designed specifically for handling form data with file uploads.
func RequestHitAPIWithFormData(ctx context.Context, method, uri string, formData map[string]string, files map[string]*multipart.FileHeader, headers map[string]string) ([]byte, int, error) {
	ctx, segment := otelTracer.Start(ctx, "pkg/util/http/RequestHitAPIWithFiles")
	defer segment.End()

	once.Do(func() {
		client = &http.Client{}
		client.Transport = &CustomTransport{
			Transport: otelhttp.NewTransport(http.DefaultTransport),
			Tracer:    otel.Tracer("http-client"),
		}
	})

	// Create multipart form data
	buff, contentType, err := createMultipartForm(formData, files)
	if err != nil {
		return nil, 0, err
	}
	defer func() {
		buff.Reset()
		bufferPool.Put(buff)
	}()

	request, err := http.NewRequestWithContext(ctx, method, uri, buff)
	if err != nil {
		return nil, 0, err
	}
	for k, v := range headers {
		request.Header.Add(k, v)
	}
	request.Header.Set("Content-type", contentType)

	response, err := client.Do(request)
	if err != nil {
		return nil, 0, err
	}
	defer response.Body.Close()

	responseBody, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, 0, err
	}

	if response.StatusCode >= 400 {
		errRes := map[string]interface{}{}
		if err = json.Unmarshal(responseBody, &errRes); err != nil {
			return []byte(fmt.Sprintf(`{"status":%d, "body":"%s"}`, response.StatusCode, string(responseBody))), response.StatusCode, nil
		}
	}
	return responseBody, response.StatusCode, nil
}

// createMultipartForm is a helper function that builds the multipart form data body.
func createMultipartForm(formData map[string]string, files map[string]*multipart.FileHeader) (*bytes.Buffer, string, error) {
	buff := bufferPool.Get().(*bytes.Buffer)
	writer := multipart.NewWriter(buff)

	// Add form fields
	for key, val := range formData {
		if err := writer.WriteField(key, val); err != nil {
			return nil, "", err
		}
	}

	// Add files
	for key, fileHeader := range files {
		// Open the file associated with the FileHeader
		file, err := fileHeader.Open()
		if err != nil {
			return nil, "", err
		}
		defer file.Close()

		// Determine the MIME type of the file
		fileExt := filepath.Ext(fileHeader.Filename)
		mimeType := mime.TypeByExtension(fileExt)
		if mimeType == "" {
			// Default to application/octet-stream if unable to determine MIME type
			mimeType = "application/octet-stream"
		}

		// Create a form file part with the correct Content-Type header
		partHeaders := textproto.MIMEHeader{}
		partHeaders.Set("Content-Disposition", fmt.Sprintf(`form-data; name="%s"; filename="%s"`, key, fileHeader.Filename))
		partHeaders.Set("Content-Type", mimeType)

		part, err := writer.CreatePart(partHeaders)
		if err != nil {
			return nil, "", err
		}
		if _, err := io.Copy(part, file); err != nil {
			return nil, "", err
		}
	}

	if err := writer.Close(); err != nil {
		return nil, "", err
	}

	return buff, writer.FormDataContentType(), nil
}
