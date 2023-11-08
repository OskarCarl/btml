package tracker

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
)

type tracker struct {
	peers map[string]bool
}

func Serve(listenAddr string) {
	t := &tracker{peers: make(map[string]bool)}
	http.HandleFunc("/list", t.list)
	http.HandleFunc("/join", t.join)
	http.HandleFunc("/leave", t.leave)
	log.Fatal(http.ListenAndServe(listenAddr, nil))
}

func (t *tracker) list(w http.ResponseWriter, r *http.Request) {
	data, _ := json.Marshal(t.peers)
	w.Write(data)
}

func (t *tracker) join(w http.ResponseWriter, r *http.Request) {
	addr, err := getAddr(r)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Println("Error:", err.Error())
		return
	}
	t.peers[addr] = true
	w.WriteHeader(http.StatusOK)
	fmt.Printf("Added %s to the list of peers in the swarm\n", addr)
}

func (t *tracker) leave(w http.ResponseWriter, r *http.Request) {
	addr, err := getAddr(r)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Println("Error:", err.Error())
		return
	}
	delete(t.peers, addr)
	w.WriteHeader(http.StatusOK)
	fmt.Printf("Removed %s from the list of peers in the swarm\n", addr)
}

func getAddr(r *http.Request) (string, error) {
	type requestData struct {
		Addr string
	}

	body := make([]byte, 1024)
	n, err := r.Body.Read(body)
	if err != nil && !errors.Is(err, io.EOF) {
		return "", fmt.Errorf("unable to read request body data from %s: %s\n%w", r.RemoteAddr, r.RequestURI, err)
	}
	var data requestData
	err = json.Unmarshal(body[0:n], &data)
	if err != nil {
		return "", fmt.Errorf("unable to parse request body data from %s: %s\n%w", r.RemoteAddr, r.RequestURI, err)
	}
	return data.Addr, nil
}
