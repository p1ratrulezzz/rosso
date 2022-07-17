package strconv

import (
   "strconv"
   "unicode/utf8"
)

func FormatInt[T Signed](value T, base int) string {
   return strconv.FormatInt(int64(value), base)
}

func Itoa[T Signed](value T) string {
   return FormatInt(value, 10)
}

func Number[T Ordered](value T) string {
   return label(value, "", " K", " M", " B", " T")
}

func Percent[T, U Signed](value T, total U) string {
   var ratio float64
   if total != 0 {
      ratio = 100 * float64(value) / float64(total)
   }
   return strconv.FormatFloat(ratio, 'f', 1, 64) + "%"
}

func Quote[T String](value T) string {
   return strconv.Quote(string(value))
}

func Rate[T, U Ordered](value T, total U) string {
   var ratio float64
   if total != 0 {
      ratio = float64(value) / float64(total)
   }
   return label(ratio, " B/s", " kB/s", " MB/s", " GB/s", " TB/s")
}

func Size[T Ordered](value T) string {
   return label(value, " B", " kB", " MB", " GB", " TB")
}

// mimesniff.spec.whatwg.org#binary-data-byte
func Valid(buf []byte) bool {
   for _, b := range buf {
      if b <= 0x08 {
         return false
      }
      if b == 0x0B {
         return false
      }
      if b >= 0x0E && b <= 0x1A {
         return false
      }
      if b >= 0x1C && b <= 0x1F {
         return false
      }
   }
   return utf8.Valid(buf)
}

func label[T Ordered](value T, units ...string) string {
   var (
      i int
      unit string
      val = float64(value)
   )
   for i, unit = range units {
      if val < 1000 {
         break
      }
      val /= 1000
   }
   if i >= 1 {
      i = 3
   }
   return strconv.FormatFloat(val, 'f', i, 64) + unit
}

type Ordered interface {
   Signed | Unsigned | ~float32 | ~float64
}

type Signed interface {
   ~int | ~int8 | ~int16 | ~int32 | ~int64
}

type String interface {
   ~[]byte | ~[]rune | ~byte | ~rune | ~string
}

type Unsigned interface {
   ~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 | ~uintptr
}
