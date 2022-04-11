package hls

import (
   "crypto/aes"
   "crypto/cipher"
   "io"
   "net/url"
   "path"
   "strconv"
   "text/scanner"
   "unicode"
)

func scanLines(buf *scanner.Scanner) {
   buf.IsIdentRune = func(r rune, i int) bool {
      return r != '\r' && r != '\n'
   }
   buf.Whitespace = 1 << '\r' | 1 << '\n'
}

func scanWords(buf *scanner.Scanner) {
   buf.IsIdentRune = func(r rune, i int) bool {
      return r == '-' || r == '.' || unicode.IsLetter(r) || unicode.IsDigit(r)
   }
   buf.Whitespace = 1 << ' '
}

type Cipher struct {
   cipher.Block
   key []byte
}

func NewCipher(src io.Reader) (*Cipher, error) {
   key, err := io.ReadAll(src)
   if err != nil {
      return nil, err
   }
   block, err := aes.NewCipher(key)
   if err != nil {
      return nil, err
   }
   return &Cipher{block, key}, nil
}

func (c Cipher) Decrypt(info Information, src io.Reader) ([]byte, error) {
   buf, err := io.ReadAll(src)
   if err != nil {
      return nil, err
   }
   if info.IV == nil {
      info.IV = c.key
   }
   cipher.NewCBCDecrypter(c.Block, info.IV).CryptBlocks(buf, buf)
   if len(buf) >= 1 {
      pad := buf[len(buf)-1]
      if len(buf) >= int(pad) {
         buf = buf[:len(buf)-int(pad)]
      }
   }
   return buf, nil
}

func (s Segment) Ext() string {
   for _, info := range s.Info {
      ext := path.Ext(info.URI.Path)
      if ext != "" {
         return ext
      }
   }
   return ""
}

func NewSegment(addr *url.URL, body io.Reader) (*Segment, error) {
   var (
      buf scanner.Scanner
      err error
      info Information
      seg Segment
   )
   buf.Init(body)
   for {
      scanWords(&buf)
      if buf.Scan() == scanner.EOF {
         break
      }
      switch buf.TokenText() {
      case "EXTINF":
         scanLines(&buf)
         buf.Scan()
         buf.Scan()
         info.URI, err = addr.Parse(buf.TokenText())
         if err != nil {
            return nil, err
         }
         seg.Info = append(seg.Info, info)
         info = Information{}
      case "EXT-X-KEY":
         for buf.Scan() != '\n' {
            switch buf.TokenText() {
            case "IV":
               buf.Scan()
               buf.Scan()
               info.IV, err = hexDecode(buf.TokenText())
               if err != nil {
                  return nil, err
               }
            case "URI":
               buf.Scan()
               buf.Scan()
               ref, err := strconv.Unquote(buf.TokenText())
               if err != nil {
                  return nil, err
               }
               seg.Key, err = addr.Parse(ref)
               if err != nil {
                  return nil, err
               }
            }
         }
      }
   }
   return &seg, nil
}
