package protobuf

import (
   "errors"
   "github.com/89z/format"
   "google.golang.org/protobuf/encoding/protowire"
   "io"
   "sort"
)

type Message map[Number]Encoder

func Unmarshal(buf []byte) (Message, error) {
   if len(buf) == 0 {
      return nil, io.ErrUnexpectedEOF
   }
   mes := make(Message)
   for len(buf) >= 1 {
      num, typ, length := protowire.ConsumeTag(buf)
      err := protowire.ParseError(length)
      if err != nil {
         return nil, err
      }
      buf = buf[length:]
      switch typ {
      case protowire.VarintType:
         buf, err = mes.consume_varint(num, buf)
      case protowire.Fixed64Type:
         buf, err = mes.consume_fixed64(num, buf)
      case protowire.Fixed32Type:
         buf, err = mes.consume_fixed32(num, buf)
      case protowire.BytesType:
         buf, err = mes.consume_raw(num, buf)
      case protowire.StartGroupType:
      case protowire.EndGroupType:
      default:
         return nil, errors.New("cannot parse reserved wire type")
      }
      if err != nil {
         return nil, err
      }
   }
   return mes, nil
}

func (m Message) Marshal() []byte {
   var (
      nums []Number
      bufs []byte
   )
   for num := range m {
      nums = append(nums, num)
   }
   sort.Slice(nums, func(a, b int) bool {
      return nums[a] < nums[b]
   })
   for _, num := range nums {
      bufs = m[num].encode(bufs, num)
   }
   return bufs
}

func (m Message) consume_fixed32(num Number, buf []byte) ([]byte, error) {
   val, length := protowire.ConsumeFixed32(buf)
   err := protowire.ParseError(length)
   if err != nil {
      return nil, err
   }
   if err := add(m, num, Fixed32(val)); err != nil {
      return nil, err
   }
   return buf[length:], nil
}

func (m Message) consume_fixed64(num Number, buf []byte) ([]byte, error) {
   val, length := protowire.ConsumeFixed64(buf)
   err := protowire.ParseError(length)
   if err != nil {
      return nil, err
   }
   if err := add(m, num, Fixed64(val)); err != nil {
      return nil, err
   }
   return buf[length:], nil
}

func (m Message) consume_raw(num Number, buf []byte) ([]byte, error) {
   var (
      length int
      val Raw
   )
   val.Bytes, length = protowire.ConsumeBytes(buf)
   err := protowire.ParseError(length)
   if err != nil {
      return nil, err
   }
   if format.String(val.Bytes) {
      val.String = string(val.Bytes)
   }
   val.Message, _ = Unmarshal(val.Bytes)
   if err := add(m, num, val); err != nil {
      return nil, err
   }
   return buf[length:], nil
}

func (m Message) consume_varint(num Number, buf []byte) ([]byte, error) {
   val, length := protowire.ConsumeVarint(buf)
   err := protowire.ParseError(length)
   if err != nil {
      return nil, err
   }
   if err := add(m, num, Varint(val)); err != nil {
      return nil, err
   }
   return buf[length:], nil
}

func (m Message) encode(buf []byte, num Number) []byte {
   buf = protowire.AppendTag(buf, num, protowire.BytesType)
   return protowire.AppendBytes(buf, m.Marshal())
}

func (Message) get_type() string { return "Message" }
