package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/tarantool/go-tarantool/v2"
)

var conn *tarantool.Connection

func main() {
	ctx := context.Background()
	host := os.Getenv("TARANTOOL_HOST")
	if host == "" {
		host = "127.0.0.1:3301"
	}

	dialer := tarantool.NetDialer{
		Address: host,
		User:    "guest",
	}
	opts := tarantool.Opts{
		Timeout: 5 * time.Second,
	}

	var err error
	conn, err = tarantool.Connect(ctx, dialer, opts)
	if err != nil {
		log.Fatalf("Cannot connect to Tarantool: %s", err)
	}

	http.HandleFunc("/kv/", kvHandler)

	fmt.Println("Server listening on :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}

func kvHandler(w http.ResponseWriter, r *http.Request) {
	key := r.URL.Path[len("/kv/"):]

	switch r.Method {
	case http.MethodPut:
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Cannot read body", http.StatusBadRequest)
			return
		}

		f := conn.Do(tarantool.NewReplaceRequest("kv").
			Context(r.Context()).
			Tuple([]interface{}{key, string(body)}))
		<-f.WaitChan()
		_, err = f.Get()
		if err != nil {
			http.Error(w, fmt.Sprintf("Tarantool error: %s", err), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusCreated)

	case http.MethodGet:
		req := tarantool.NewSelectRequest("kv").
			Context(r.Context()).
			Index("primary").
			Iterator(tarantool.IterEq).
			Limit(1).
			Key([]interface{}{key})
		f := conn.Do(req)
		<-f.WaitChan()
		v, err := f.Get()
		if err != nil {
			http.Error(w, fmt.Sprintf("Tarantool error: %s", err), http.StatusInternalServerError)
			return
		}
		if len(v) == 0 {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		value, ok := v[0].([]interface{})[1].(string)
		if !ok {
			http.Error(w, "Unexpected tuple format", http.StatusInternalServerError)
			return
		}
		_, err = w.Write([]byte(value))
		if err != nil {
			http.Error(w, fmt.Sprintf("Write error: %s", err), http.StatusInternalServerError)
		}

	case http.MethodDelete:
		f := conn.Do(
			tarantool.NewDeleteRequest("kv").
				Context(r.Context()).
				Index("primary").
				Key([]interface{}{key}),
		)
		<-f.WaitChan()
		_, err := f.Get()
		if err != nil {
			http.Error(w, fmt.Sprintf("Tarantool error: %s", err), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusNoContent)

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}
