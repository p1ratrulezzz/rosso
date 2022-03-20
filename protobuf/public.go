package protobuf

import (
   "google.golang.org/protobuf/encoding/protowire"
   "io"
)

type Message map[protowire.Number]any

func Decode(src io.Reader) (Message, error) {
   buf, err := io.ReadAll(src)
   if err != nil {
      return nil, err
   }
   return Unmarshal(buf)
}

func Unmarshal(buf []byte) (Message, error) {
   mes := make(Message)
   for len(buf) >= 1 {
      num, typ, fLen := protowire.ConsumeField(buf)
      err := protowire.ParseError(fLen)
      if err != nil {
         return nil, err
      }
      _, _, tLen := protowire.ConsumeTag(buf[:fLen])
      if err := protowire.ParseError(tLen); err != nil {
         return nil, err
      }
      val := buf[tLen:fLen]
      switch typ {
      case protowire.BytesType:
         err = mes.consumeBytes(num, val)
      case protowire.Fixed64Type:
         err = mes.consumeFixed64(num, val)
      case protowire.Fixed32Type:
         err = mes.consumeFixed32(num, val)
      case protowire.VarintType:
         err = mes.consumeVarint(num, val)
      }
      if err != nil {
         return nil, err
      }
      buf = buf[fLen:]
   }
   return mes, nil
}

func (m Message) Add(num protowire.Number, val Message) {
   switch value := m[num].(type) {
   case nil:
      m[num] = val
   case Message:
      m[num] = []Message{value, val}
   case []Message:
      m[num] = append(value, val)
   }
}

func (m Message) Get(num protowire.Number) Message {
   val, _ := m[num].(Message)
   return val
}

func (m Message) GetMessages(num protowire.Number) []Message {
   switch val := m[num].(type) {
   case []Message:
      return val
   case Message:
      return []Message{val}
   }
   return nil
}

func (m Message) GetString(num protowire.Number) string {
   switch val := m[num].(type) {
   case string:
      return val
   case Message:
      return m.GetString(-num)
   }
   return ""
}

func (m Message) GetUint64(num protowire.Number) uint64 {
   val, _ := m[num].(uint64)
   return val
}

func (m Message) Marshal() []byte {
   var buf []byte
   for num, val := range m {
      if num >= protowire.MinValidNumber {
         buf = appendField(buf, num, val)
      }
   }
   return buf
}
