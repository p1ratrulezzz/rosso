package net

import (
   "bufio"
   "io"
   "net/http"
   "net/textproto"
   "net/url"
   "strconv"
   "strings"
)

func ReadRequest(src io.Reader) (*http.Request, error) {
   var req http.Request
   text := textproto.NewReader(bufio.NewReader(src))
   // .Method
   sMethodPath, err := text.ReadLine()
   if err != nil {
      return nil, err
   }
   // GET /fdfe/details?doc=com.instagram.android HTTP/1.1
   methodPath := strings.Fields(sMethodPath)
   if len(methodPath) != 3 {
      return nil, textproto.ProtocolError(sMethodPath)
   }
   req.Method = methodPath[0]
   // .URL
   addr, err := url.Parse(methodPath[1])
   if err != nil {
      return nil, err
   }
   req.URL = addr
   // .URL.Host
   head, err := text.ReadMIMEHeader()
   if err != nil {
      return nil, err
   }
   req.URL.Host = head.Get("Host")
   // .Header
   req.Header = http.Header(head)
   // .ContentLength
   sLength := head.Get("Content-Length")
   if sLength != "" {
      length, err := strconv.ParseInt(sLength, 10, 64)
      if err != nil {
         return nil, err
      }
      req.ContentLength = length
   }
   // .Body
   req.Body = io.NopCloser(text.R)
   return &req, nil
}

// text/plain encoding algorithm
// html.spec.whatwg.org/multipage/form-control-infrastructure.html
func ReadQuery(src io.Reader) url.Values {
   vals := make(url.Values)
   buf := bufio.NewScanner(src)
   for buf.Scan() {
      kv := strings.SplitN(buf.Text(), "=", 2)
      if len(kv) == 2 {
         vals.Add(kv[0], kv[1])
      }
   }
   return vals
}
