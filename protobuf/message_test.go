package protobuf

import (
   "fmt"
   "os"
   "testing"
)

func Test_Marshal(t *testing.T) {
   checkin := Message{
      4: Message{ // checkin
         1: Message{ // build
            10: Varint(29), // sdkVersion
         },
      },
      14: Varint(3), // version
      18: Message{ // deviceConfiguration
         1: Varint(999), // touchScreen
         2: Varint(999),
         3: Varint(999),
         4: Varint(999),
         5: Varint(999),
         6: Varint(999),
         7: Varint(999),
         8: Varint(999),
         9: Slice[String]{
            "org.apache.http.legacy",
            "android.test.runner",
            "global-miui11-empty.jar",
         },
         11: String("nativePlatform"),
         15: Slice[String]{
            "GL_OES_compressed_ETC1_RGB8_texture",
            "GL_KHR_texture_compression_astc_ldr",
         },
         26: Slice[Message]{
            {1: String("android.hardware.bluetooth")},
            {1: String("android.hardware.bluetooth_le")},
            {1: String("android.hardware.camera")},
            {1: String("android.hardware.camera.autofocus")},
            {1: String("android.hardware.camera.front")},
            {1: String("android.hardware.location")},
            {1: String("android.hardware.location.gps")},
            {1: String("android.hardware.location.network")},
            {1: String("android.hardware.microphone")},
            {1: String("android.hardware.opengles.aep")},
            {1: String("android.hardware.screen.landscape")},
            {1: String("android.hardware.screen.portrait")},
            {1: String("android.hardware.sensor.accelerometer")},
            {1: String("android.hardware.sensor.compass")},
            {1: String("android.hardware.sensor.gyroscope")},
            {1: String("android.hardware.telephony")},
            {1: String("android.hardware.touchscreen")},
            {1: String("android.hardware.touchscreen.multitouch")},
            {1: String("android.hardware.usb.host")},
            {1: String("android.hardware.wifi")},
            {1: String("android.software.device_admin")},
            {1: String("android.software.midi")},
         },
      },
   }
   fmt.Println(checkin)
}

func Test_Unmarshal(t *testing.T) {
   buf, err := os.ReadFile("com.pinterest.txt")
   if err != nil {
      t.Fatal(err)
   }
   response_wrapper, err := Unmarshal(buf)
   if err != nil {
      t.Fatal(err)
   }
   doc_V2 := response_wrapper.Message(1).Message(2).Message(4)
   if v := doc_V2.Message(13).Message(1).Messages(17); len(v) != 4 {
      t.Fatal("File", v)
   }
   if v, err := doc_V2.Message(13).Message(1).Varint(3); err != nil {
      t.Fatal(err)
   } else if v != 10218030 {
      t.Fatal("VersionCode", v)
   }
   if v, err := doc_V2.Message(13).Message(1).String(4); err != nil {
      t.Fatal(err)
   } else if v != "10.21.0" {
      t.Fatal("VersionString", v)
   }
   if v, err := doc_V2.Message(13).Message(1).Varint(9); err != nil {
      t.Fatal(err)
   } else if v != 47705639 {
      t.Fatal("Size", v)
   }
   if v, err := doc_V2.Message(13).Message(1).String(16); err != nil {
      t.Fatal(err)
   } else if v != "Jun 14, 2022" {
      t.Fatal("Date", v)
   }
   if v, err := doc_V2.String(5); err != nil {
      t.Fatal(err)
   } else if v != "Pinterest" {
      t.Fatal("title", v)
   }
   if v, err := doc_V2.String(6); err != nil {
      t.Fatal(err)
   } else if v != "Pinterest" {
      t.Fatal("creator", v)
   }
   if v, err := doc_V2.Message(8).String(2); err != nil {
      t.Fatal(err)
   } else if v != "USD" {
      t.Fatal("currencyCode", v)
   }
   if v, err := doc_V2.Message(13).Message(1).Varint(70); err != nil {
      t.Fatal(err)
   } else if v != 750510010 {
      t.Fatal("NumDownloads", v)
   }
}
