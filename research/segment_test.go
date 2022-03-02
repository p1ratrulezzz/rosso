package hls

import (
   "io"
   "net/http"
   "os"
   "strings"
   "testing"
)

func doMaster() (*Master, error) {
   var buf strings.Builder
   buf.WriteString("http://link.theplatform.com/s/dJ5BDC/media/guid/2198311517")
   buf.WriteString("/3htV4fvVt4Z8gDZHqlzPOGLSMgcGc_vy")
   req, err := http.NewRequest("GET", buf.String(), nil)
   if err != nil {
      return nil, err
   }
   // We need "MPEG4", otherwise you get a "EXT-X-KEY" with "skd" scheme:
   req.URL.RawQuery = "formats=MPEG4,M3U"
   // This redirects:
   res, err := new(http.Client).Do(req)
   if err != nil {
      return nil, err
   }
   defer res.Body.Close()
   return NewMaster(res)
}

func doSegment() (*Segment, error) {
   mas, err := doMaster()
   if err != nil {
      return nil, err
   }
   res, err := http.Get(mas.Stream[0].URI)
   if err != nil {
      return nil, err
   }
   defer res.Body.Close()
   return NewSegment(res)
}

func doKey(seg *Segment) (*Block, error) {
   res, err := http.Get(seg.Key.URI)
   if err != nil {
      return nil, err
   }
   defer res.Body.Close()
   buf, err := io.ReadAll(res.Body)
   if err != nil {
      return nil, err
   }
   return NewCipher(buf)
}

func TestSegment(t *testing.T) {
   seg, err := doSegment()
   if err != nil {
      t.Fatal(err)
   }
   block, err := doKey(seg)
   if err != nil {
      t.Fatal(err)
   }
   res, err := http.Get(seg.Info[1].URI)
   if err != nil {
      t.Fatal(err)
   }
   defer res.Body.Close()
   buf, err := io.ReadAll(res.Body)
   if err != nil {
      t.Fatal(err)
   }
   buf = block.Decrypt(buf)
   if err := os.WriteFile("ignore.ts", buf, os.ModePerm); err != nil {
      t.Fatal(err)
   }
}
