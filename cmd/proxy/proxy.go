package main

import (
   "fmt"
   "github.com/89z/rosso/crypto"
   "io"
   "net"
   "net/http"
   "strconv"
)

func (s spy_conn) Read(buf []byte) (int, error) {
   num, err := s.Conn.Read(buf)
   if hello, err := crypto.Parse_TLS(buf[:num]); err == nil {
      ja3, err := crypto.Format_JA3(hello)
      if err == nil {
         fmt.Printf("%q\n", buf[:num])
         fmt.Print("\t", ja3, "\n")
      }
   }
   return num, err
}

func main() {
   http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
      err := root(w, r)
      if err != nil {
         fmt.Println(err)
      }
   })
   port := ":8080"
   fmt.Println(port)
   err := http.ListenAndServe(port, nil)
   if err != nil {
      panic(err)
   }
}

type spy_conn struct {
   net.Conn
}

func root(w http.ResponseWriter, r *http.Request) error {
   if r.Method == http.MethodConnect {
      hijacker, ok := w.(http.Hijacker)
      if ok {
         client_conn, _, err := hijacker.Hijack()
         if err != nil {
            return err
         }
         defer client_conn.Close()
         dst, err := net.Dial("tcp", r.URL.Host)
         if err != nil {
            return err
         }
         defer dst.Close()
         buf := []byte("HTTP/1.1 ")
         buf = strconv.AppendInt(buf, http.StatusOK, 10)
         buf = append(buf, "\n\n"...)
         if _, err := client_conn.Write(buf); err != nil {
            return err
         }
         src := spy_conn{client_conn}
         if _, err := io.Copy(dst, src); err != nil {
            return err
         }
      }
   }
   return nil
}
