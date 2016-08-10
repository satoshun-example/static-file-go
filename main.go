package main

import (
	"bytes"
	"compress/gzip"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"mime"
	"os"
	"path/filepath"
)

type file struct {
	name string
	data string
	mime string
}

func processDir(c chan [1]string, dir string) {
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		log.Fatal(err)
	}
	for _, file := range files {
		d := filepath.Join(dir, file.Name())
		if file.IsDir() {
			processDir(c, d)
		} else {
			c <- [1]string{d}
		}
	}
}

func main() {
	flag.Parse()

	c := make(chan [1]string, 4)
	go func() {
		for _, arg := range flag.Args() {
			processDir(c, arg)
		}
		close(c)
	}()

	fc := make(chan file, 4)
	go func() {
		var b bytes.Buffer
		var b2 bytes.Buffer
		for d := range c {
			f, err := os.Open(d[0])
			if err != nil {
				log.Fatal(err)
			}
			if _, err := b.ReadFrom(f); err != nil {
				log.Fatal(err)
			}
			f.Close()
			writer, err := gzip.NewWriterLevel(&b2, gzip.BestCompression)
			if err != nil {
				log.Fatal(err)
			}
			if _, err := writer.Write(b.Bytes()); err != nil {
				log.Fatal(err)
			}
			writer.Close()
			fc <- file{
				name: d[0],
				data: b2.String(),
				mime: mime.TypeByExtension(filepath.Ext(d[0])),
			}
		}
		close(fc)
	}()

	for f := range fc {
		fmt.Println(f)
	}
}
