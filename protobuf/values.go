package protobuf

import (
   "google.golang.org/protobuf/encoding/protowire"
)

func consumeBytes(b []byte) ([]byte, error) {
   val, vLen := protowire.ConsumeBytes(b)
   err := protowire.ParseError(vLen)
   if err != nil {
      return nil, err
   }
   return val, nil
}

func consumeField(b []byte) (protowire.Number, protowire.Type, int, error) {
   num, typ, fLen := protowire.ConsumeField(b)
   err := protowire.ParseError(fLen)
   if err != nil {
      return 0, 0, 0, err
   }
   return num, typ, fLen, nil
}

func consumeFixed32(b []byte) (uint32, error) {
   val, vLen := protowire.ConsumeFixed32(b)
   err := protowire.ParseError(vLen)
   if err != nil {
      return 0, err
   }
   return val, nil
}

func consumeFixed64(b []byte) (uint64, error) {
   val, vLen := protowire.ConsumeFixed64(b)
   err := protowire.ParseError(vLen)
   if err != nil {
      return 0, err
   }
   return val, nil
}

func consumeGroup(num protowire.Number, b []byte) ([]byte, error) {
   val, vLen := protowire.ConsumeGroup(num, b)
   err := protowire.ParseError(vLen)
   if err != nil {
      return nil, err
   }
   return val, nil
}

func consumeTag(b []byte) (int, error) {
   _, _, tLen := protowire.ConsumeTag(b)
   err := protowire.ParseError(tLen)
   if err != nil {
      return 0, err
   }
   return tLen, nil
}

func consumeVarint(b []byte) (uint64, error) {
   val, vLen := protowire.ConsumeVarint(b)
   err := protowire.ParseError(vLen)
   if err != nil {
      return 0, err
   }
   return val, nil
}

func (m Message) GetString(keys ...protowire.Number) string {
   for _, key := range keys {
      switch val := m[key].(type) {
      case Message:
         m = val
      case string:
         return val
      }
   }
   return ""
}

func (m Message) GetUint64(keys ...protowire.Number) uint64 {
   for _, key := range keys {
      switch val := m[key].(type) {
      case Message:
         m = val
      case uint64:
         return val
      }
   }
   return 0
}

func (m Message) SetStrings(val []string, keys ...protowire.Number) {
   b := m
   for index, key := range keys {
      if index == len(keys)-1 {
         b[key] = val
      } else {
         c, ok := b[key].(Message)
         if !ok {
            c = make(Message)
            b[key] = c
         }
         b = c
      }
   }
}

func (m Message) SetUint64(val uint64, keys ...protowire.Number) {
   b := m
   for index, key := range keys {
      if index == len(keys)-1 {
         b[key] = val
      } else {
         c, ok := b[key].(Message)
         if !ok {
            c = make(Message)
            b[key] = c
         }
         b = c
      }
   }
}

func (m Message) addUint64(k protowire.Number, v uint64) {
   switch u := m[k].(type) {
   case nil:
      m[k] = v
   case uint64:
      m[k] = []uint64{u, v}
   case []uint64:
      m[k] = append(u, v)
   }
}

func (m Message) addUint32(k protowire.Number, v uint32) {
   switch u := m[k].(type) {
   case nil:
      m[k] = v
   case uint32:
      m[k] = []uint32{u, v}
   case []uint32:
      m[k] = append(u, v)
   }
}

func (m Message) add(k protowire.Number, v Message) {
   switch u := m[k].(type) {
   case nil:
      m[k] = v
   case Message:
      m[k] = []Message{u, v}
   case []Message:
      m[k] = append(u, v)
   }
}

func (m Message) addString(k protowire.Number, v string) {
   switch u := m[k].(type) {
   case nil:
      m[k] = v
   case string:
      m[k] = []string{u, v}
   case []string:
      m[k] = append(u, v)
   }
}

func (m Message) addBytes(k protowire.Number, v []byte) {
   switch u := m[k].(type) {
   case nil:
      m[k] = v
   case []byte:
      m[k] = [][]byte{u, v}
   case [][]byte:
      m[k] = append(u, v)
   }
}
