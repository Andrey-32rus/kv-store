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

type Service struct {
	c *tarantool.Connection
}

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

	conn, err := tarantool.Connect(ctx, dialer, opts)
	if err != nil {
		log.Fatalf("Cannot connect to Tarantool: %s", err)
	}
	defer conn.Close()

	s := &Service{
		c: conn,
	}

	http.HandleFunc("/kv/", s.kvHandler)

	fmt.Println("Server listening on :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}

func (s *Service) kvHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	key := r.URL.Path[len("/kv/"):]

	switch r.Method {
	case http.MethodPut:
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Cannot read body", http.StatusBadRequest)
			return
		}
		s.PutHandler(ctx, key, body, w)

	case http.MethodGet:
		s.GetHandler(ctx, key, w)

	case http.MethodDelete:
		s.DeleteHandler(ctx, key, w)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *Service) PutHandler(ctx context.Context, key string, body []byte, w http.ResponseWriter) {
	f := s.c.Do(tarantool.NewReplaceRequest("kv").
		Context(ctx).
		Tuple([]interface{}{key, string(body)}))
	<-f.WaitChan()
	_, err := f.Get()
	if err != nil {
		http.Error(w, fmt.Sprintf("Tarantool error: %s", err), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func (s *Service) GetHandler(ctx context.Context, key string, w http.ResponseWriter) {
	req := tarantool.NewSelectRequest("kv").
		Context(ctx).
		Index("primary").
		Iterator(tarantool.IterEq).
		Limit(1).
		Key([]interface{}{key})
	f := s.c.Do(req)
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
}

func (s *Service) DeleteHandler(ctx context.Context, key string, w http.ResponseWriter) {
	f := s.c.Do(
		tarantool.NewDeleteRequest("kv").
			Context(ctx).
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
}
