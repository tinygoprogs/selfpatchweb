package main

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"github.com/kr/binarydist"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
)

func versionSupply(version string, done chan bool) func(w http.ResponseWriter, r *http.Request) {
	f := func(w http.ResponseWriter, r *http.Request) {
		fmt.Printf("[+] version req from: %s\n", r.RemoteAddr)
		fmt.Fprintf(w, "%s", version)
		done <- true
	}
	return f
}

func createDiff() ([]byte, error) {
	tmp := "/tmp/newthingy"
	cli_src := "cli/self-patch.go"
	cmd_str := fmt.Sprintf("GOPATH=/usr/share/gocode/ go build -o %s %s", tmp, cli_src)
	println("[+] executing ", cmd_str)
	cmd := exec.Command("/bin/bash", "-c", cmd_str)
	o, e := cmd.Output()
	if e != nil {
		fmt.Printf("error, cmd returned: %s\n->%#v\n", o, e)
	}

	fnew, err := ioutil.ReadFile(tmp)
	if err != nil {
		return []byte{}, errors.New("open tmp")
	}
	fold, err := ioutil.ReadFile("./self-patch")
	if err != nil {
		return []byte{}, errors.New("open self-patch")
	}

	if bytes.Equal(fold, fnew) {
		println("nothing to diff!")
		return []byte{}, nil
	}

	patch := bytes.Buffer{}
	diff_err := binarydist.Diff(
		bytes.NewReader(fold),
		bytes.NewReader(fnew),
		bufio.NewWriter(&patch),
	)
	fmt.Printf("diff [len=%d, err=%#v]\n", patch.Len(), diff_err)
	return patch.Bytes(), diff_err
}
func binaryPatchSupply(done chan bool) func(w http.ResponseWriter, r *http.Request) {

	diff, err := createDiff()
	if err != nil {
		done <- true
		return func(w http.ResponseWriter, r *http.Request) {
			r.Body.Close()
		}
	}
	return func(w http.ResponseWriter, r *http.Request) {
		n, err := w.Write(diff)
		if err != nil {
			println("error: ", err.Error())
			return
		}
		defer r.Body.Close()
		if n != len(diff) {
			println("send incomplete diff, well..")
			return
		}
		done <- true
	}
}

func pushNewVersion(version string) {
	srv := http.Server{
		Addr: "127.0.0.1:1234",
	}

	versionUpdated := make(chan bool)
	binaryUpdated := make(chan bool)
	http.HandleFunc("/update/version", versionSupply(version, versionUpdated))
	http.HandleFunc("/update/binary", binaryPatchSupply(binaryUpdated))

	go func() {
		srv.ListenAndServe()
	}()

	for true {
		if <-versionUpdated && <-binaryUpdated {
			os.Exit(0)
		} else {
			println("cannot happen..")
			os.Exit(1)
		}
	}
}

func main() {
	if len(os.Args) != 2 {
		os.Exit(1)
	}
	pushNewVersion(os.Args[1])
}
