// TLS and JA3 parsers
package crypto

import (
   "encoding/binary"
   "fmt"
   "github.com/refraction-networking/utls"
   "io"
   "net"
   "net/http"
   "strconv"
   "strings"
)

// 2454fe66222e468b886b8e552b5e2f3b
const AndroidJA3 =
   "769,49195-49196-52393-49199-49200-52392-158-159-49161-49162-49171-49172-" +
   "51-57-156-157-47-53,65281-0-23-35-13-16-11-10,23,0"

func extensionType(ext tls.TLSExtension) (uint16, error) {
   buf, err := io.ReadAll(ext)
   if err != nil || len(buf) <= 1 {
      return 0, err
   }
   return binary.BigEndian.Uint16(buf), nil
}

type ClientHello struct {
   *tls.ClientHelloSpec
   Version uint16
}

func ParseTLS(buf []byte) (*ClientHello, error) {
   // Content Type, Version
   if dLen := len(buf); dLen <= 2 {
      return nil, invalidSlice{2, dLen}
   }
   version := binary.BigEndian.Uint16(buf[1:])
   // unsupported extension 0x16
   fin := tls.Fingerprinter{AllowBluntMimicry: true}
   spec, err := fin.FingerprintClientHello(buf)
   if err != nil {
      return nil, err
   }
   return &ClientHello{spec, version}, nil
}

func ParseJA3(str string) (*ClientHello, error) {
   tokens := strings.Split(str, ",")
   if tLen := len(tokens); tLen <= 4 {
      return nil, invalidSlice{4, tLen}
   }
   var version uint16
   _, err := fmt.Sscan(tokens[0], &version)
   if err != nil {
      return nil, err
   }
   hello := ClientHello{
      new(tls.ClientHelloSpec), version,
   }
   // build CipherSuites
   cipherKeys := strings.Split(tokens[1], "-")
   for _, cipherKey := range cipherKeys {
      var cipher uint16
      _, err := fmt.Sscan(cipherKey, &cipher)
      if err != nil {
         return nil, err
      }
      hello.CipherSuites = append(hello.CipherSuites, cipher)
   }
   // build extenions list
   extKeys := strings.Split(tokens[2], "-")
   for _, extKey := range extKeys {
      var ext tls.TLSExtension
      switch extKey {
      case "0":
         ext = &tls.SNIExtension{}
      case "10":
         var curves []tls.CurveID
         curveKeys := strings.Split(tokens[3], "-")
         for _, curveKey := range curveKeys {
            var curve tls.CurveID
            _, err := fmt.Sscan(curveKey, &curve)
            if err != nil {
               return nil, err
            }
            curves = append(curves, curve)
         }
         ext = &tls.SupportedCurvesExtension{curves}
      case "11":
         var points []uint8
         pointKeys := strings.Split(tokens[4], "-")
         for _, pointKey := range pointKeys {
            var point uint8
            _, err := fmt.Sscan(pointKey, &point)
            if err != nil {
               return nil, err
            }
            points = append(points, point)
         }
         ext = &tls.SupportedPointsExtension{points}
      case "13":
         // this cant be empty, so just use the Go default
         ext = &tls.SignatureAlgorithmsExtension{
            []tls.SignatureScheme{
               0x804, 0x403, 0x807, 0x805, 0x806, 0x401,
               0x501, 0x601, 0x503, 0x603, 0x201, 0x203,
            },
         }
      case "16":
         // if we leave this empty, it will fail on any HTTP/2 servers
         ext = &tls.ALPNExtension{
            []string{"http/1.1"},
         }
      case "23":
         ext = &tls.UtlsExtendedMasterSecretExtension{}
      case "35":
         ext = &tls.SessionTicketExtension{}
      case "65281":
         ext = &tls.RenegotiationInfoExtension{}
      }
      hello.Extensions = append(hello.Extensions, ext)
   }
   return &hello, nil
}

func (c ClientHello) FormatJA3() (string, error) {
   buf := new(strings.Builder)
   // Version
   fmt.Fprint(buf, c.Version)
   // Cipher Suites
   buf.WriteByte(',')
   for key, val := range c.CipherSuites {
      if key >= 1 {
         buf.WriteByte('-')
      }
      fmt.Fprint(buf, val)
   }
   // Extensions
   buf.WriteByte(',')
   var (
      curves []tls.CurveID
      points []uint8
   )
   for key, val := range c.Extensions {
      if key >= 1 {
         buf.WriteByte('-')
      }
      typ, err := extensionType(val)
      if err != nil {
         return "", err
      }
      fmt.Fprint(buf, typ)
      switch ext := val.(type) {
      case *tls.SupportedCurvesExtension:
         curves = ext.Curves
      case *tls.SupportedPointsExtension:
         points = ext.SupportedPoints
      }
   }
   // Elliptic curves
   buf.WriteByte(',')
   for key, val := range curves {
      if key >= 1 {
         buf.WriteByte('-')
      }
      fmt.Fprint(buf, val)
   }
   // ECPF
   buf.WriteByte(',')
   for key, val := range points {
      if key >= 1 {
         buf.WriteByte('-')
      }
      fmt.Fprint(buf, val)
   }
   return buf.String(), nil
}

func (c ClientHello) Transport() *http.Transport {
   return &http.Transport{
      DialTLS: func(network, addr string) (net.Conn, error) {
         conn, err := net.Dial(network, addr)
         if err != nil {
            return nil, err
         }
         host, _, err := net.SplitHostPort(addr)
         if err != nil {
            return nil, err
         }
         config := &tls.Config{ServerName: host}
         uconn := tls.UClient(conn, config, tls.HelloCustom)
         if err := uconn.ApplyPreset(c.ClientHelloSpec); err != nil {
            return nil, err
         }
         if err := uconn.Handshake(); err != nil {
            return nil, err
         }
         return uconn, nil
      },
   }
}

type invalidSlice struct {
   index, length int
}

func (i invalidSlice) Error() string {
   index, length := int64(i.index), int64(i.length)
   var buf []byte
   buf = append(buf, "index out of range ["...)
   buf = strconv.AppendInt(buf, index, 10)
   buf = append(buf, "] with length "...)
   buf = strconv.AppendInt(buf, length, 10)
   return string(buf)
}
