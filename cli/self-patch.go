package main

import (
	"fmt"
	"github.com/inconshreveable/go-update"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

func maybeUpdate() string {
	url := "http://127.0.0.1:1234/"
	cli := &http.Client{}

	var version string

	r1, err := cli.Get(url + "update/version")
	if err != nil {
		//log.Printf("no version update: %s", err.Error())
	} else {
		defer r1.Body.Close()
		buf, err := ioutil.ReadAll(r1.Body)
		if err == nil {
			version = string(buf)
		}
	}

	r2, err := cli.Get(url + "update/binary")
	if err != nil {
		//log.Printf("no binary update: %s", err.Error())
	} else if r2.ContentLength == 0 {
		log.Printf("empty binary/update content")
	} else {
		defer r2.Body.Close()
		log.Printf("applying update..: %#v", r2.Header)
		err := update.Apply(r2.Body, update.Options{
			Patcher: update.NewBSDiffPatcher(),
		})
		if err != nil {
			log.Printf("didn't work, rolling back..")
			if rerr := update.RollbackError(err); rerr != nil {
				log.Printf("failed to rollback from bad update: %v", rerr)
			}
			log.Printf("binary update failed: %s", err.Error())
			return version
		}
		log.Printf("success!")
	}

	return version
}

func main() {
	version := "v0.0.0"
	for true {
		time.Sleep(time.Second * 1)
		newversion := maybeUpdate()
		if len(newversion) != 0 {
			version = newversion
			fmt.Printf("[+] runningAonAAersion: %s\n", version)
		}
	}
}
