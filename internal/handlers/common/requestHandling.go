package handlers

import (
	"errors"
	"io"
	"net/http"
)

const recordError = "No valid credential record sent in request body."

func ExtractBody(request *http.Request) ([]byte, error) {
	if request.Body == nil {
		return nil, errors.New("no body")
	}
	data, err := io.ReadAll(request.Body)
	if data != nil && err == nil {

		return data, nil
	}
	return nil, err
}
