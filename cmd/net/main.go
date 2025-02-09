package main

import (
   "flag"
   "github.com/89z/rosso/http"
   "os"
)

type flags struct {
   golang bool
   https bool
   name string
   output string
}

func main() {
   var f flags
   // f
   flag.StringVar(&f.name, "f", "", "input file")
   // g
   flag.BoolVar(&f.golang, "g", false, "request as Go code")
   // o
   flag.StringVar(&f.output, "o", "", "output file")
   // s
   flag.BoolVar(&f.https, "s", false, "HTTPS")
   flag.Parse()
   if f.name != "" {
      create, err := os.Create(f.output)
      if err != nil {
         create = os.Stdout
      }
      defer create.Close()
      open, err := os.Open(f.name)
      if err != nil {
         panic(err)
      }
      defer open.Close()
      req, err := http.Read_Request(open)
      if err != nil {
         panic(err)
      }
      if req.URL.Scheme == "" {
         if f.https {
            req.URL.Scheme = "https"
         } else {
            req.URL.Scheme = "http"
         }
      }
      if f.golang {
         err := Write_Request(req, create)
         if err != nil {
            panic(err)
         }
      } else {
         err := write(req, create)
         if err != nil {
            panic(err)
         }
      }
   } else {
      flag.Usage()
   }
}
