package protobuf

import (
   "fmt"
   "testing"
)

var checkin = Message{
   Tag{4, "checkin"}:Message{
      Tag{1, "build"}:Message{
         Tag{10, "sdkVersion"}: uint64(29),
      },
   },
}

func TestGet(t *testing.T) {
   {
      get := checkin.Get(4, "checkin").Get(1, "build")
      fmt.Println(get)
   }
   {
      get := checkin.Get2(Tag{4, "checkin"}, Tag{1, "build"})
      fmt.Println(get)
   }
   {
      get := checkin.Get3([]Tag{{4, "checkin"}, {1, "build"}})
      fmt.Println(get)
   }
}
