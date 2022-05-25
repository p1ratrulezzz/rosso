package format

import (
   "bytes"
   "encoding/json"
   "io"
   "net/http"
   "net/http/httputil"
   "os"
   "path/filepath"
   "strconv"
   "strings"
   "time"
   "unicode/utf8"
)

// mimesniff.spec.whatwg.org#binary-data-byte
func IsString(buf []byte) bool {
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
   // []byte{0xE0, '<'}
   return utf8.Valid(buf)
}

func Create[T any](value T, elem ...string) error {
   name := filepath.Join(elem...)
   err := os.MkdirAll(filepath.Dir(name), os.ModePerm)
   if err != nil {
      return err
   }
   os.Stderr.WriteString("Create " + name + "\n")
   file, err := os.Create(name)
   if err != nil {
      return err
   }
   defer file.Close()
   enc := json.NewEncoder(file)
   enc.SetIndent("", " ")
   return enc.Encode(value)
}

func Open[T any](elem ...string) (*T, error) {
   name := filepath.Join(elem...)
   file, err := os.Open(name)
   if err != nil {
      return nil, err
   }
   defer file.Close()
   value := new(T)
   if err := json.NewDecoder(file).Decode(value); err != nil {
      return nil, err
   }
   return value, nil
}

func Label[T Number](value T, unit ...string) string {
   var (
      i int
      symbol string
      val = float64(value)
   )
   for i, symbol = range unit {
      if val < 1000 {
         break
      }
      val /= 1000
   }
   if i >= 1 {
      i = 3
   }
   return strconv.FormatFloat(val, 'f', i, 64) + symbol
}

func LabelNumber[T Number](value T) string {
   return Label(value, "", " K", " M", " B", " T")
}

func LabelRate[T Number](value T) string {
   return Label(value, " B/s", " kB/s", " MB/s", " GB/s", " TB/s")
}

func LabelSize[T Number](value T) string {
   return Label(value, " B", " kB", " MB", " GB", " TB")
}

type LogLevel int

func (l LogLevel) Dump(req *http.Request) error {
   quote := func(b []byte) []byte {
      if !IsString(b) {
         b = strconv.AppendQuote(nil, string(b))
      }
      if !bytes.HasSuffix(b, []byte{'\n'}) {
         b = append(b, '\n')
      }
      return b
   }
   switch l {
   case 0:
      os.Stderr.WriteString(req.Method)
      os.Stderr.WriteString(" ")
      os.Stderr.WriteString(req.URL.String())
      os.Stderr.WriteString("\n")
   case 1:
      buf, err := httputil.DumpRequest(req, true)
      if err != nil {
         return err
      }
      os.Stderr.Write(quote(buf))
   case 2:
      buf, err := httputil.DumpRequestOut(req, true)
      if err != nil {
         return err
      }
      os.Stderr.Write(quote(buf))
   }
   return nil
}

type Number interface {
   float64 | int | int64 | ~uint64
}

func ProgressBytes(dst io.Writer, bytes int64) *Progress {
   return &Progress{Writer: dst, bytes: bytes}
}

func ProgressChunks(dst io.Writer, chunks int) *Progress {
   return &Progress{Writer: dst, chunks: chunks}
}

func (p *Progress) AddChunk(bytes int64) {
   p.bytesRead += bytes
   p.chunksRead += 1
   p.bytes = int64(p.chunks) * p.bytesRead / p.chunksRead
}

func (p *Progress) Write(buf []byte) (int, error) {
   if p.time.IsZero() {
      p.time = time.Now()
      p.timeLap = time.Now()
   }
   since := time.Since(p.timeLap)
   if since >= time.Second {
      os.Stderr.WriteString(p.String())
      os.Stderr.WriteString("\n")
      p.timeLap = p.timeLap.Add(since)
   }
   write, err := p.Writer.Write(buf)
   p.bytesWritten += write
   return write, err
}

type Progress struct {
   io.Writer
   bytes int64
   bytesRead int64
   bytesWritten int
   chunks int
   chunksRead int64
   time time.Time
   timeLap time.Time
}

func (p Progress) String() string {
   percent := func(value int, total int64) string {
      var ratio float64
      if total != 0 {
         ratio = 100 * float64(value) / float64(total)
      }
      return strconv.FormatFloat(ratio, 'f', 1, 64) + "%"
   }
   ratio := percent(p.bytesWritten, p.bytes)
   rate := float64(p.bytesWritten) / time.Since(p.time).Seconds()
   var buf strings.Builder
   buf.WriteString(ratio)
   buf.WriteByte('\t')
   buf.WriteString(LabelSize(p.bytesWritten))
   buf.WriteByte('\t')
   buf.WriteString(LabelRate(rate))
   return buf.String()
}
