package protobuf

import (
   "encoding/json"
   "google.golang.org/protobuf/encoding/protowire"
)

func consumeJSON(buf []byte) (interface{}, error) {
   if buf[0] == '[' {
      var raw []json.RawMessage
      err := json.Unmarshal(buf, &raw)
      if err != nil {
         return nil, err
      }
      var toks []interface{}
      for _, buf := range raw {
         var tok token
         err := json.Unmarshal(buf, &tok)
         if err != nil {
            return nil, err
         }
         toks = append(toks, tok)
      }
      return toks, nil
   }
   var raw struct {
      Type protowire.Type
      Value json.RawMessage
   }
   err := json.Unmarshal(buf, &raw)
   if err != nil {
      return nil, err
   }
   var tok interface{}
   switch raw.Type {
   case protowire.Fixed32Type:
      var val uint32
      err := json.Unmarshal(raw.Value, &val)
      if err != nil {
         return nil, err
      }
      tok = val
   case protowire.Fixed64Type, protowire.VarintType:
      var val uint64
      err := json.Unmarshal(raw.Value, &val)
      if err != nil {
         return nil, err
      }
      tok = val
   case protowire.BytesType:
      if raw.Value[0] == '"' {
         var val string
         err := json.Unmarshal(raw.Value, &val)
         if err != nil {
            return nil, err
         }
         tok = val
      } else {
         val := make(message)
         err := json.Unmarshal(raw.Value, &val)
         if err != nil {
            return nil, err
         }
         tok = val
      }
   }
   return token{raw.Type, tok}, nil
}

type fixed32 uint32

func (f fixed32) MarshalJSON() ([]byte, error) {
   tok := token{
      protowire.Fixed32Type, uint32(f),
   }
   return json.Marshal(tok)
}

type fixed64 uint64

func (f fixed64) MarshalJSON() ([]byte, error) {
   tok := token{
      protowire.Fixed64Type, uint64(f),
   }
   return json.Marshal(tok)
}

func (m message) MarshalJSON() ([]byte, error) {
   mes := map[protowire.Number]interface{}(m)
   return json.Marshal(mes)
}

func (m *message) UnmarshalJSON(buf []byte) error {
   var raw map[protowire.Number]json.RawMessage
   err := json.Unmarshal(buf, &raw)
   if err != nil {
      return err
   }
   for key, buf := range raw {
      val, err := consumeJSON(buf)
      if err != nil {
         return err
      }
      (*m)[key] = val
   }
   return nil
}

type varint uint64

func (v varint) MarshalJSON() ([]byte, error) {
   tok := token{
      protowire.VarintType, uint64(v),
   }
   return json.Marshal(tok)
}
