package main

import (
	"errors"
	"fmt"
	"maps"
	"net"
	"regexp"
	"slices"
	"strings"
)

type StatusLine struct {
	Method   string
	Endpoint string
	HttpV    string
}

type Headers struct {
	Headers map[string]string
	Length  int
}

type Body struct {
	Body   string
	Length int
}

type Response struct {
	RespHeaders    map[string]string
	RespBody       string
	RespStatusCode uint16
	HttpV          string
	Message        string
}

func createResponseString(response Response) string {
	responseStr := ""
	responseStr += fmt.Sprintf("%v %v %v\r\n",
		response.HttpV, response.RespStatusCode, response.Message)
	for header := range maps.Keys(response.RespHeaders) {
		responseStr += fmt.Sprintf("%v: %v\r\n", header, response.RespHeaders[header])
	}
	responseStr += "\r\n"
	responseStr += response.RespBody
	return responseStr
}

func readRequest(conn net.Conn) string {
	buffer := make([]byte, REQ_BYTE_SIZE)
	_, err := conn.Read(buffer)
	if err != nil {
		fmt.Printf("error reading connection: %s\n", err.Error())
		return ""
	}
	return string(buffer)
}

func getStatusLine(request string) (StatusLine, error) {
	lines := strings.Split(request, "\r\n")
	if len(lines) == 0 {
		return StatusLine{}, errors.New("expected three sections 'Status line', 'Headers' & 'Response body'")
	}

	parts := strings.Split(lines[0], " ")
	if len(parts) < 3 {
		return StatusLine{}, errors.New("malformed 'Status line'")
	}

	return StatusLine{
		Method:   parts[0],
		Endpoint: parts[1],
		HttpV:    parts[2],
	}, nil
}

func getHeaders(request string) Headers {
	headerStartIdx := strings.Index(request, "\r\n")
	headerEndIdx := strings.Index(request, "\r\n\r\n")
	if headerStartIdx == -1 || headerEndIdx == -1 {
		return Headers{}
	}

	headerSection := request[headerStartIdx:headerEndIdx]
	headers := strings.Split(headerSection, "\r\n")
	if len(headers) == 0 {
		return Headers{}
	}

	result := Headers{
		Headers: make(map[string]string),
		Length:  len(headers),
	}

	for _, header := range headers {
		var sepIdx = -1
		for idx, chr := range []byte(header) {
			if chr == ':' {
				sepIdx = idx
				break
			}
		}
		if sepIdx == -1 {
			continue
		}
		headerKey := strings.TrimSpace(header[0:sepIdx])
		headerValue := strings.TrimSpace(header[sepIdx+1:])
		result.Headers[headerKey] = headerValue
	}
	return result
}

func routeMatches(expr string, endpoint string) bool {
	res, err := regexp.Match(expr, []byte(endpoint))
	if err != nil {
		fmt.Printf("Error while matching paths, Error %v\n", err)
		return false
	}
	return res
}

func getAcceptedEncodings(value string) []string {
	reqEncodings := strings.Split(value, ",")
	supportedEncodings := make([]string, 0)

	for _, encoding := range reqEncodings {
		if slices.Contains(SUPPORTED_COMPRESSION_ALGOS, strings.TrimSpace(encoding)) {
			supportedEncodings = append(supportedEncodings, encoding)
		}
	}
	return supportedEncodings
}
