package main

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/chi/v5"
	"go.etcd.io/etcd/clientv3"
)

var (
	key                = "KEY"
	unset              = "UNSET"
	set                = "SET"
	initialVersion int = 2
)

func main() {

	cli, err := clientv3.New(clientv3.Config{
		Endpoints: []string{"http://etcd:2379"},
	})
	if err != nil {
		log.Fatal("unable to set init value")
		return
	}
	defer cli.Close()
	cli.Put(context.TODO(), key, unset)

	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Post("/notify", NotifyHandler)
	log.Fatal(http.ListenAndServe(":8080", r))

}

func NotifyHandler(w http.ResponseWriter, r *http.Request) {
	// Connect to etcd
	cli, err := clientv3.New(clientv3.Config{
		Endpoints: []string{"http://etcd:2379"},
	})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		log.Fatal(err)
		return
	}
	defer cli.Close()
	resp, err := cli.Get(context.Background(), key)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		log.Fatal(err)
		return
	}
	if len(resp.Kvs) > 0 {
		kv := resp.Kvs[0]
		if kv == nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			log.Fatal(err)
			return
		}

		if bytes.Equal(kv.Value, []byte(set)) {
			w.WriteHeader(http.StatusCreated)
			w.Write([]byte("locked"))
			return
		}
	}
	fmt.Println(resp.Kvs)
	cmp := clientv3.Compare(clientv3.Value(key), "=", "UNSET")

	txnResp, err := cli.Txn(context.Background()).
		If(cmp).
		Then(clientv3.OpPut(key, "SET")).
		Commit()

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		log.Fatal(err)
		return
	}

	time.AfterFunc(10*time.Second, func() {
		cli, err := clientv3.New(clientv3.Config{
			Endpoints: []string{"http://etcd:2379"},
		})
		if err != nil {
			log.Fatal("unable to set init value")
			return
		}
		cmp := clientv3.Compare(clientv3.Value(key), "=", set)
		_, err = cli.Txn(context.Background()).
			If(cmp).
			Then(clientv3.OpPut(key, unset)).
			Commit()

		if err != nil {

			log.Fatal(err)
			return
		}
	})

	log.Default().Println("Bibin's logs", txnResp)
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte("submitted"))
	log.Default().Println("Bibin's handler is done!!")

}
