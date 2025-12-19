package handlers

import (
	"log/slog"
	"net/http"
	"strconv"
)

func Handler1(w http.ResponseWriter, r *http.Request) {
	slog.Debug("Request accept", "Host:", r.Host, "Metohod:", r.Method, "Headers:", r.Header, "Body:", r.Body)
	n, err := w.Write([]byte("Hello World\n"))
	if err != nil {
		slog.Error("erroe in handler1", slog.Any("w.Write: ", err))
	}
	w.Header().Add("ContentLength", strconv.Itoa(n))
}
