package request

import (
	"bufio"
	"bytes"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
)

var supportedHTTPMethods = map[string]struct{}{
	http.MethodConnect: {},
	http.MethodDelete:  {},
	http.MethodGet:     {},
	http.MethodHead:    {},
	http.MethodOptions: {},
	http.MethodPatch:   {},
	http.MethodPost:    {},
	http.MethodPut:     {},
	http.MethodTrace:   {},
}

type HTTPRequest struct {
	domain  string
	method  string
	port    string
	path    string
	version string
	rawData []byte
}

func (r *HTTPRequest) Domain() string {
	return r.domain
}

func (r *HTTPRequest) Method() string {
	return r.method
}

func (r *HTTPRequest) Port() string {
	return r.port
}

func (r *HTTPRequest) Path() string {
	return r.path
}

func (r *HTTPRequest) Version() string {
	return r.version
}

func (r *HTTPRequest) RawData() []byte {
	return r.rawData
}

func ParseHTTPRequest(reader io.Reader) (*HTTPRequest, error) {
	return processRequest(reader)
}

func (r *HTTPRequest) IsSupportedMethod() bool {
	_, exists := supportedHTTPMethods[r.method]
	return exists
}

func (r *HTTPRequest) IsMethodConnect() bool {
	return r.method == http.MethodConnect
}

// SanitizeRequest Функция для очистки заголовков и форматирования запроса
func (r *HTTPRequest) SanitizeRequest() {
	requestStr := string(r.rawData)
	headerEndIndex := strings.Index(requestStr, "\r\n\r\n")
	if headerEndIndex == -1 {
		return
	}

	headers := strings.Split(requestStr[:headerEndIndex], "\r\n")
	body := requestStr[headerEndIndex+4:]

	// Обновляем первую строку с методом, URI и версией протокола
	headers[0] = r.method + " " + r.path + " " + r.version

	var buffer bytes.Buffer
	crlf := "\r\n"

	for _, header := range headers {
		// Убираем заголовок Proxy-Connection
		if !strings.HasPrefix(header, "Proxy-Connection") {
			buffer.WriteString(header)
			buffer.WriteString(crlf)
		}
	}

	// Добавляем оставшуюся часть
	buffer.WriteString(crlf)
	buffer.WriteString(body)

	r.rawData = []byte(buffer.String())
}

// Парсинг HTTP-запроса
func processRequest(r io.Reader) (*HTTPRequest, error) {
	sb := strings.Builder{}
	tee := io.TeeReader(r, &sb)
	request, err := http.ReadRequest(bufio.NewReader(tee))
	if err != nil {
		return nil, err
	}

	p := &HTTPRequest{}
	p.rawData = []byte(sb.String())

	p.domain, p.port, err = net.SplitHostPort(request.Host)
	if err != nil {
		p.domain = request.Host
		p.port = ""
	}

	p.method = request.Method
	p.version = request.Proto
	p.path = request.URL.Path

	if request.URL.RawQuery != "" {
		p.path += "?" + request.URL.RawQuery
	}

	if request.URL.RawFragment != "" {
		p.path += "#" + request.URL.RawFragment
	}
	if p.path == "" {
		p.path = "/"
	}

	request.Body.Close()
	return p, nil
}

// Формирование URI с учетом query и fragment
func buildURI(url *url.URL) string {
	var uri string
	if url.Path != "" {
		uri = url.Path
	} else {
		uri = "/"
	}

	if url.RawQuery != "" {
		uri += "?" + url.RawQuery
	}

	if url.RawFragment != "" {
		uri += "#" + url.RawFragment
	}

	return uri
}
