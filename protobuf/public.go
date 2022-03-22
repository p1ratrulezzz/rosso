package protobuf

import (
   "github.com/89z/format"
   "google.golang.org/protobuf/encoding/protowire"
   "io"
)

type Message map[Number]Token

func Decode(in io.Reader) (Message, error) {
   buf, err := io.ReadAll(in)
   if err != nil {
      return nil, err
   }
   return Unmarshal(buf)
}

func Unmarshal(in []byte) (Message, error) {
   mes := make(Message)
   for len(in) >= 1 {
      num, typ, fLen := protowire.ConsumeField(in)
      if err := protowire.ParseError(fLen); err != nil {
         return nil, err
      }
      _, _, tLen := protowire.ConsumeTag(in[:fLen])
      if err := protowire.ParseError(tLen); err != nil {
         return nil, err
      }
      buf := in[tLen:fLen]
      switch typ {
      case protowire.BytesType:
         val, vLen := protowire.ConsumeBytes(buf)
         if err := protowire.ParseError(vLen); err != nil {
            return nil, err
         }
         if len(val) >= 1 {
            embed, err := Unmarshal(val)
            if err != nil {
               add(mes, num, String(val))
            } else if format.IsBinary(val) {
               add(mes, num, embed)
            } else {
               add(mes, num, String(val))
               add(mes, -num, embed)
            }
         } else {
            add(mes, num, String(""))
         }
      case protowire.Fixed32Type:
         val, vLen := protowire.ConsumeFixed32(buf)
         if err := protowire.ParseError(vLen); err != nil {
            return nil, err
         }
         add(mes, num, Uint32(val))
      case protowire.Fixed64Type:
         val, vLen := protowire.ConsumeFixed64(buf)
         if err := protowire.ParseError(vLen); err != nil {
            return nil, err
         }
         add(mes, num, Uint64(val))
      case protowire.VarintType:
         val, vLen := protowire.ConsumeVarint(buf)
         if err := protowire.ParseError(vLen); err != nil {
            return nil, err
         }
         add(mes, num, Uint64(val))
      }
      in = in[fLen:]
   }
   return mes, nil
}

func (m Message) Add(num Number, val Message) {
   add(m, num, val)
}

func (m Message) Get(num Number) Message {
   switch value := m[num].(type) {
   case Message:
      return value
   case String:
      return m.Get(-num)
   }
   return nil
}

func (m Message) GetMessages(num Number) []Message {
   switch value := m[num].(type) {
   case tokens[Message]:
      return value
   case Message:
      return []Message{value}
   }
   return nil
}

func (m Message) GetString(num Number) String {
   return get[String](m, num)
}

func (m Message) GetUint64(num Number) Uint64 {
   return get[Uint64](m, num)
}

type Number = protowire.Number

type String string

type Token interface {
   appendField([]byte, Number) []byte
}

type Uint32 uint32

type Uint64 uint64

func (m Message) Marshal() []byte {
   var buf []byte
   for num, tok := range m {
      if num >= protowire.MinValidNumber {
         buf = tok.appendField(buf, num)
      }
   }
   return buf
}
