package main

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"os"
	"path"
	"strconv"
	"strings"
)

func baseCont() Response {
	return Response{
		RespHeaders:    make(map[string]string),
		HttpV:          "HTTP/1.1",
		Message:        "OK",
		RespBody:       "",
		RespStatusCode: 200,
	}
}

func echoCont(request, reqPath string) Response {
	respBody := ""
	headers := getHeaders(request)
	respHeaders := make(map[string]string)
	echo := strings.TrimPrefix(reqPath, "/echo/")

	compressing := false
	respHeaders["Content-Type"] = "text/plain"
	if len(getAcceptedEncodings(headers.Headers["Accept-Encoding"])) > 0 {
		respHeaders["Content-Encoding"] = "gzip"
		compressing = true
	}
	if compressing {
		var buf bytes.Buffer

		// GZIP Code
		zw := gzip.NewWriter(&buf)
		zw.Write([]byte(echo))
		zw.Flush()
		zw.Close()

		respHeaders["Content-Length"] = fmt.Sprint(len(buf.String()))
		respBody = buf.String()
	} else {
		respHeaders["Content-Length"] = fmt.Sprint(len(echo))
		respBody = echo

		// We can also do it like this -> (1st way)
		// fmt.Fprintf(conn, "Content-Length: %v\r\n\r\n", len(buf.String()))
		// fmt.Fprintf(conn, "Content-Length: %v\r\n\r\n", len(echo))
		// fmt.Fprintf(conn, "%v", echo)

		// We can also do it like this -> (2nd way)
		// conn.Write([]byte("Content-Type: text/plain\r\n"))
		// conn.Write([]byte(fmt.Sprintf("Content-Length: %v\r\n\r\n", len(echo))))
		// conn.Write([]byte(fmt.Sprintf("%v", echo)))
	}
	return Response{
		RespHeaders:    respHeaders,
		HttpV:          "HTTP/1.1",
		Message:        "OK",
		RespBody:       respBody,
		RespStatusCode: 200,
	}
}

func userAgentCont(req string) Response {
	headers := getHeaders(req)
	respHeaders := make(map[string]string)
	respBody := ""

	respHeaders["Content-Type"] = "text/plain"
	respHeaders["Content-Length"] = fmt.Sprint(len(headers.Headers["User-Agent"]))
	respBody = headers.Headers["User-Agent"]

	return Response{
		RespHeaders:    respHeaders,
		HttpV:          "HTTP/1.1",
		Message:        "OK",
		RespBody:       respBody,
		RespStatusCode: 200,
	}
}

func filsCont(reqPath, dirName string) Response {
	msg := "OK"
	respBody := ""
	var httpStatusCode uint16 = 200
	respHeaders := make(map[string]string)

	fileName := strings.Replace(reqPath, "/files/", "", 1)
	file, err := os.Open(dirName + fileName)
	if err != nil {
		httpStatusCode = 404
		msg = "Not Found"
	} else {
		fileContent := make([]byte, 1024)
		fileSize, err := file.Read(fileContent)
		if err != nil {
			msg = "Not Found"
			httpStatusCode = 404
		} else {
			httpStatusCode = 200
			respHeaders["Content-Type"] = "application/octet-stream"
			respHeaders["Content-Length"] = fmt.Sprint(fileSize)
			respBody = string(fileContent)
		}
	}

	return Response{
		RespHeaders:    respHeaders,
		HttpV:          "HTTP/1.1",
		Message:        msg,
		RespBody:       respBody,
		RespStatusCode: httpStatusCode,
	}
}

func postFileCont(request, reqPath, dir string) Response {
	msg := "OK"
	var httpStatusCode uint16 = 200

	file := strings.TrimPrefix(reqPath, "/files/")
	filePath := path.Join(dir, file)

	headers := getHeaders(request)
	contentLength, err := strconv.Atoi(headers.Headers["Content-Length"])
	if err != nil {
		fmt.Println("Error while parsing header")
		return Response{
			RespHeaders:    make(map[string]string),
			HttpV:          "HTTP/1.1",
			Message:        "Invalid request",
			RespBody:       "",
			RespStatusCode: 400,
		}
	}

	idxStart := strings.Index(request, "\r\n\r\n")
	if idxStart == -1 {
		return Response{
			RespHeaders:    make(map[string]string),
			HttpV:          "HTTP/1.1",
			Message:        "Invalid request",
			RespBody:       "",
			RespStatusCode: 400,
		}
	}
	fileBody := request[idxStart+4:]
	if len(fileBody) > contentLength {
		fileBody = fileBody[:contentLength]
	}

	err = os.WriteFile(filePath, []byte(fileBody), 0644)
	if err != nil {
		fmt.Println("Error while creating file")
		httpStatusCode = 500
		msg = "Internal Server Error"
	} else {
		httpStatusCode = 201
		msg = "Created"
	}

	return Response{
		RespHeaders:    make(map[string]string),
		HttpV:          "HTTP/1.1",
		Message:        msg,
		RespBody:       "",
		RespStatusCode: httpStatusCode,
	}
}

func notFound() Response {
	return Response{
		RespHeaders:    make(map[string]string),
		HttpV:          "HTTP/1.1",
		Message:        "Not Found",
		RespBody:       "",
		RespStatusCode: 404,
	}
}
