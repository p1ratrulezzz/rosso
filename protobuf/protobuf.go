// Protocol Buffers
package protobuf

import (
   "fmt"
   "github.com/89z/format"
   "google.golang.org/protobuf/encoding/protowire"
   "io"
   "strconv"
   "strings"
)

func appendField(buf []byte, num protowire.Number, val interface{}) []byte {
   switch val := val.(type) {
   case uint32:
      buf = protowire.AppendTag(buf, num, protowire.Fixed32Type)
      buf = protowire.AppendFixed32(buf, val)
   case uint64:
      buf = protowire.AppendTag(buf, num, protowire.VarintType)
      buf = protowire.AppendVarint(buf, val)
   case string:
      buf = protowire.AppendTag(buf, num, protowire.BytesType)
      buf = protowire.AppendString(buf, val)
   case []byte:
      buf = protowire.AppendTag(buf, num, protowire.BytesType)
      buf = protowire.AppendBytes(buf, val)
   case Message:
      buf = protowire.AppendTag(buf, num, protowire.BytesType)
      buf = protowire.AppendBytes(buf, val.Marshal())
   case []uint64:
      for _, ran := range val {
         buf = appendField(buf, num, ran)
      }
   case []string:
      for _, ran := range val {
         buf = appendField(buf, num, ran)
      }
   case []Message:
      for _, ran := range val {
         buf = appendField(buf, num, ran)
      }
   }
   return buf
}

type Message map[Tag]interface{}

func Unmarshal(buf []byte) (Message, error) {
   if len(buf) == 0 {
      return nil, io.ErrUnexpectedEOF
   }
   mes := make(Message)
   for len(buf) >= 1 {
      num, typ, fLen, err := consumeField(buf)
      if err != nil {
         return nil, err
      }
      tLen, err := consumeTag(buf[:fLen])
      if err != nil {
         return nil, err
      }
      bVal := buf[tLen:fLen]
      switch typ {
      case protowire.VarintType:
         val, err := consumeVarint(bVal)
         if err != nil {
            return nil, err
         }
         mes.addUint64(num, val)
      case protowire.Fixed64Type:
         val, err := consumeFixed64(bVal)
         if err != nil {
            return nil, err
         }
         mes.addUint64(num, val)
      case protowire.Fixed32Type:
         val, err := consumeFixed32(bVal)
         if err != nil {
            return nil, err
         }
         mes.addUint32(num, val)
      case protowire.BytesType:
         buf, err := consumeBytes(bVal)
         if err != nil {
            return nil, err
         }
         if !format.IsBinary(buf) {
            mes.addString(num, string(buf))
         } else {
            mNew, err := Unmarshal(buf)
            if err != nil {
               mes.addBytes(num, buf)
            } else {
               mes.Add(num, "", mNew)
            }
         }
      case protowire.StartGroupType:
         buf, err := consumeGroup(num, bVal)
         if err != nil {
            return nil, err
         }
         mNew, err := Unmarshal(buf)
         if err != nil {
            return nil, err
         }
         mes.Add(num, "", mNew)
      }
      buf = buf[fLen:]
   }
   return mes, nil
}

func (m Message) GoString() string {
   str := new(strings.Builder)
   str.WriteString("protobuf.Message{")
   first := true
   for key, val := range m {
      if first {
         first = false
      } else {
         str.WriteString(",\n")
      }
      fmt.Fprintf(str, "%#v:", key)
      switch typ := val.(type) {
      case uint32:
         fmt.Fprintf(str, "uint32(%v)", typ)
      case uint64:
         fmt.Fprintf(str, "uint64(%v)", typ)
      default:
         fmt.Fprintf(str, "%#v", val)
      }
   }
   str.WriteByte('}')
   return str.String()
}

func (m Message) Marshal() []byte {
   var buf []byte
   for key, val := range m {
      buf = appendField(buf, key.Number, val)
   }
   return buf
}

type Tag struct {
   protowire.Number
   String string
}

// encoding/json
func (t Tag) MarshalText() ([]byte, error) {
   num := int64(t.Number)
   return strconv.AppendInt(nil, num, 10), nil
}
