package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/dgryski/go-topk"
)

func main() {

	f := flag.String("f", "", "file to read")

	flag.Parse()

	var r io.Reader

	if *f == "" {
		r = os.Stdin
	} else {
		var err error
		r, err = os.Open(*f)
		if err != nil {
			log.Fatal(err)
		}
	}

	tk := topk.New(500)
	sc := bufio.NewScanner(r)

	items := 0

	for sc.Scan() {
		items++
		tk.Insert(sc.Text())
	}

	if err := sc.Err(); err != nil {
		log.Fatal(err)
	}

	for _, v := range tk.Keys() {
		fmt.Println(v.Key, v.Count)
	}
}
