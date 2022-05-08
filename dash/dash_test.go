package dash

import (
   "fmt"
   "net/url"
   "os"
   "testing"
)

type testType struct {
   name string
   addr string
}

var tests = []testType{
   {
      "channel4.mpd",
      "https://ak-jos-c4assets-com.akamaized.net/CH4_44_7_900_18926001001003_001/CH4_44_7_900_18926001001003_001_J01.ism/stream.mpd",
   }, {
      "roku.mpd",
      "https://vod.delivery.roku.com/41e834bbaecb4d27890094e3d00e8cfb/aaf72928242741a6ab8d0dfefbd662ca/87fe48887c78431d823a845b377a0c0f/index.mpd",
   },
}

func TestDASH(t *testing.T) {
   for _, test := range tests {
      addr, err := url.Parse(test.addr)
      if err != nil {
         t.Fatal(err)
      }
      file, err := os.Open(test.name)
      if err != nil {
         t.Fatal(err)
      }
      period, err := NewPeriod(file)
      if err != nil {
         t.Fatal(err)
      }
      if err := file.Close(); err != nil {
         t.Fatal(err)
      }
      video := period.Video(0)
      addrs, err := video.URL(addr)
      if err != nil {
         t.Fatal(err)
      }
      for _, addr := range addrs {
         fmt.Println(addr)
      }
   }
}
