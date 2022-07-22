package dash

import (
   "net/url"
   "strconv"
   "strings"
)

func (r Representation) String() string {
   var b []byte
   b = append(b, "ID:"...)
   b = append(b, r.ID...)
   if r.Width + r.Bandwidth >= 1 {
      b = append(b, "\n  "...)
   }
   if r.Width >= 1 {
      b = append(b, "Width:"...)
      b = strconv.AppendInt(b, r.Width, 10)
      b = append(b, " Height:"...)
      b = strconv.AppendInt(b, r.Height, 10)
   }
   if r.Bandwidth >= 1 {
      if r.Width >= 1 {
         b = append(b, ' ')
      }
      b = append(b, "Bandwidth:"...)
      b = strconv.AppendInt(b, r.Bandwidth, 10)
   }
   b = append(b, "\n  MimeType:"...)
   b = append(b, r.MimeType...)
   if r.Codecs != "" {
      b = append(b, " Codecs:"...)
      b = append(b, r.Codecs...)
   }
   if r.Adaptation.Lang != "" {
      b = append(b, " Lang:"...)
      b = append(b, r.Adaptation.Lang...)
   }
   if r.Adaptation.Role != nil {
      b = append(b, " Role:"...)
      b = append(b, r.Adaptation.Role.Value...)
   }
   return string(b)
}

type Segment struct {
   D int `xml:"d,attr"` // duration
   R int `xml:"r,attr"` // repeat
   T int `xml:"t,attr"` // time
}

func (s Segment) Time() string {
   return strconv.Itoa(s.T)
}

type SegmentTemplate struct {
   Initialization string `xml:"initialization,attr"`
   Media string `xml:"media,attr"`
   SegmentTimeline struct {
      S []Segment
   }
   StartNumber *int `xml:"startNumber,attr"`
}

func (r Representation) Media(base *url.URL) ([]*url.URL, error) {
   var start int
   if r.SegmentTemplate.StartNumber != nil {
      start = *r.SegmentTemplate.StartNumber
   }
   var refs []*url.URL
   for _, seg := range r.SegmentTemplate.SegmentTimeline.S {
      for seg.T = start; seg.R >= 0; seg.R-- {
         raw := r.replace_ID(r.SegmentTemplate.Media)
         if r.SegmentTemplate.StartNumber != nil {
            raw = strings.Replace(raw, "$Number$", seg.Time(), 1)
            seg.T++
            start++
         } else {
            raw = strings.Replace(raw, "$Time$", seg.Time(), 1)
            seg.T += seg.D
            start += seg.D
         }
         ref, err := base.Parse(raw)
         if err != nil {
            return nil, err
         }
         refs = append(refs, ref)
      }
   }
   return refs, nil
}

type Representations []Representation

func (p Presentation) Representation() Representations {
   var reps []Representation
   for i, ada := range p.Period.AdaptationSet {
      for _, rep := range ada.Representation {
         rep.Adaptation = &p.Period.AdaptationSet[i]
         if rep.Codecs == "" {
            rep.Codecs = ada.Codecs
         }
         if rep.ContentProtection == nil {
            rep.ContentProtection = ada.ContentProtection
         }
         if rep.MimeType == "" {
            rep.MimeType = ada.MimeType
         }
         if rep.SegmentTemplate == nil {
            rep.SegmentTemplate = ada.SegmentTemplate
         }
         reps = append(reps, rep)
      }
   }
   return reps
}

type Representation struct {
   Adaptation *Adaptation
   Bandwidth int64 `xml:"bandwidth,attr"`
   Codecs string `xml:"codecs,attr"`
   ContentProtection *ContentProtection
   Height int64 `xml:"height,attr"`
   ID string `xml:"id,attr"`
   MimeType string `xml:"mimeType,attr"`
   SegmentTemplate *SegmentTemplate
   Width int64 `xml:"width,attr"`
}

type Adaptation struct {
   Codecs string `xml:"codecs,attr"`
   ContentProtection *ContentProtection
   Lang string `xml:"lang,attr"`
   MimeType string `xml:"mimeType,attr"`
   Role *struct {
      Value string `xml:"value,attr"`
   }
   SegmentTemplate *SegmentTemplate
   Representation []Representation
}

type ContentProtection struct {
   Default_KID string `xml:"default_KID,attr"`
}

type Presentation struct {
   Period struct {
      AdaptationSet []Adaptation
   }
}

func (r Representation) Ext() string {
   switch r.MimeType {
   case "video/mp4":
      return ".m4v"
   case "audio/mp4":
      return ".m4a"
   }
   return ""
}

func (r Representation) Initialization() string {
   return r.replace_ID(r.SegmentTemplate.Initialization)
}

func (r Representation) Role() string {
   if r.Adaptation.Role == nil {
      return ""
   }
   return r.Adaptation.Role.Value
}

func (r Representation) replace_ID(s string) string {
   return strings.Replace(s, "$RepresentationID$", r.ID, 1)
}
