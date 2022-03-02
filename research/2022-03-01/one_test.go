package hls

import (
   "encoding/json"
   "fmt"
   "net/http"
   "strings"
   "testing"
)

func oneHref() (string, error) {
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

func twoHref() (*master, error) {
   href, err := oneHref()
   if err != nil {
      return nil, err
   }
   res, err := http.Get(href)
   if err != nil {
      return nil, err
   }
   defer res.Body.Close()
   return one(res)
}

func TestOne(t *testing.T) {
   mas, err := twoHref()
   if err != nil {
      t.Fatal(err)
   }
   for _, med := range mas.media {
      fmt.Printf("%+v\n", med)
   }
   for _, str := range mas.stream {
      fmt.Printf("%+v\n", str)
   }
}
