package hls

import (
   "encoding/json"
   "fmt"
   "net/http"
   "strings"
   "testing"
)

func hrefOne() (string, error) {
   var buf strings.Builder
   buf.WriteString("http://open.live.bbc.co.uk")
   buf.WriteString("/mediaselector/6/select/version/2.0/mediaset/pc/vpid/")
   buf.WriteString("p0bkp6nx")
   res, err := http.Get(buf.String())
   if err != nil {
      return "", nil
   }
   defer res.Body.Close()
   var set struct {
      Media []struct {
         Connection []struct {
            Href string
         }
      }
   }
   if err := json.NewDecoder(res.Body).Decode(&set); err != nil {
      return "", err
   }
   return set.Media[1].Connection[0].Href, nil
}

func TestMaster(t *testing.T) {
   href, err := hrefOne()
   if err != nil {
      t.Fatal(err)
   }
   res, err := http.Get(href)
   if err != nil {
      t.Fatal(err)
   }
   defer res.Body.Close()
   mas, err := NewMaster(res)
   if err != nil {
      t.Fatal(err)
   }
   for _, med := range mas.Media {
      fmt.Printf("%+v\n", med)
   }
   for _, str := range mas.Stream {
      fmt.Printf("%+v\n", str)
   }
}
