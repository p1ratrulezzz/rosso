package hls

import (
   "bytes"
   "io"
   "strconv"
)

func merge(forms []Format) int {
   if len(forms) >= 1 {
      form := forms[len(forms)-1]
      if len(form) >= 1 {
         return len(forms)-1
      }
   }
   return -1
}

type Format map[string]string

func Decode(src io.Reader, dir string) ([]Format, error) {
   buf, err := io.ReadAll(src)
   if err != nil {
      return nil, err
   }
   return Unmarshal(buf, dir), nil
}

func Unmarshal(buf []byte, dir string) []Format {
   lines := bytes.FieldsFunc(buf, func(r rune) bool {
      return r == '\n'
   })
   var pass1 []Format
   for _, line := range lines {
      if line[0] == '#' {
         form := make(Format)
         pairs := reader{line}
         pairs.readBytes(':', '"')
         for {
            if pairs.buf == nil {
               break
            }
            var pair reader
            pair.buf = pairs.readBytes(',', '"')
            key := pair.readBytes('=', '"')
            if pair.buf != nil {
               val := string(pair.buf)
               unq, err := strconv.Unquote(val)
               if err == nil {
                  val = unq
               }
               form[string(key)] = val
            }
         }
         pass1 = append(pass1, form)
      } else {
         ind := merge(pass1)
         if ind >= 0 {
            pass1[ind]["URI"] = string(line)
         } else {
            form := make(Format)
            form["URI"] = string(line)
            pass1 = append(pass1, form)
         }
      }
   }
   var pass2 []Format
   uris := make(map[string]bool)
   for _, form := range pass1 {
      uri, ok := form["URI"]
      if ok && !uris[uri] {
         form["URI"] = dir + form["URI"]
         pass2 = append(pass2, form)
         uris[uri] = true
      }
   }
   return pass2
}

type reader struct {
   buf []byte
}

func (r *reader) readBytes(sep, enc byte) []byte {
   out := true
   for key, val := range r.buf {
      if val == enc {
         out = !out
      }
      if out && val == sep {
         buf := r.buf[:key]
         r.buf = r.buf[key+1:]
         return buf
      }
   }
   buf := r.buf
   r.buf = nil
   return buf
}
