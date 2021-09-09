package ja3

import (
   "crypto/sha256"
   "fmt"
   "github.com/refraction-networking/utls"
   "net"
   "net/http"
   "strconv"
   "strings"
)

const (
   // ChromeAuto mocks Chrome 78
   ChromeAuto = "769,47–53–5–10–49161–49162–49171–49172–50–56–19–4,0–10–11,23–24–25,0"
   // SafariAuto mocks Safari 604.1
   SafariAuto = "771,4865-4866-4867-49196-49195-49188-49187-49162-49161-52393-49200-49199-49192-49191-49172-49171-52392-157-156-61-60-53-47-49160-49170-10,65281-0-23-13-5-18-16-11-51-45-43-10-21,29-23-24-25,0"
)

// extMap maps extension values to the TLSExtension object associated with the
// number. Some values are not put in here because they must be applied in a
// special way. For example, "10" is the SupportedCurves extension which is also
// used to calculate the JA3 signature. These JA3-dependent values are applied
// after the instantiation of the map.
var extMap = map[string]tls.TLSExtension{
   "0": &tls.SNIExtension{},
   "5": &tls.StatusRequestExtension{},
   // These are applied later
   // "10": &tls.SupportedCurvesExtension{...}
   // "11": &tls.SupportedPointsExtension{...}
   "13": &tls.SignatureAlgorithmsExtension{
      SupportedSignatureAlgorithms: []tls.SignatureScheme{
         tls.ECDSAWithP256AndSHA256,
         tls.PSSWithSHA256,
         tls.PKCS1WithSHA256,
         tls.ECDSAWithP384AndSHA384,
         tls.PSSWithSHA384,
         tls.PKCS1WithSHA384,
         tls.PSSWithSHA512,
         tls.PKCS1WithSHA512,
         tls.PKCS1WithSHA1,
      },
   },
   "16": &tls.ALPNExtension{
      AlpnProtocols: []string{"h2", "http/1.1"},
   },
   "18": &tls.SCTExtension{},
   "21": &tls.UtlsPaddingExtension{GetPaddingLen: tls.BoringPaddingStyle},
   "23": &tls.UtlsExtendedMasterSecretExtension{},
   "27": &tls.FakeCertCompressionAlgsExtension{},
   "28": &tls.FakeRecordSizeLimitExtension{},
   "35": &tls.SessionTicketExtension{},
   "43": &tls.SupportedVersionsExtension{
      Versions: []uint16{
         tls.GREASE_PLACEHOLDER,
         tls.VersionTLS13,
         tls.VersionTLS12,
         tls.VersionTLS11,
         tls.VersionTLS10,
      },
   },
   "44": &tls.CookieExtension{},
   "45": &tls.PSKKeyExchangeModesExtension{
      Modes: []uint8{tls.PskModeDHE},
   },
   "51":    &tls.KeyShareExtension{
      KeyShares: []tls.KeyShare{},
   },
   "13172": &tls.NPNExtension{},
   "65281": &tls.RenegotiationInfoExtension{
      Renegotiation: tls.RenegotiateOnceAsClient,
   },
}

// NewTransport creates an http.Transport which mocks the given JA3 signature when HTTPS is used
func NewTransport(ja3 string) (*http.Transport, error) {
	return NewTransportWithConfig(ja3, &tls.Config{})
}

// NewTransportWithConfig creates an http.Transport object given a utls.Config
func NewTransportWithConfig(ja3 string, config *tls.Config) (*http.Transport, error) {
	spec, err := stringToSpec(ja3)
	if err != nil {
		return nil, err
	}

	dialtls := func(network, addr string) (net.Conn, error) {
		dialConn, err := net.Dial(network, addr)
		if err != nil {
			return nil, err
		}

		config.ServerName = strings.Split(addr, ":")[0]

		uTLSConn := tls.UClient(dialConn, config, tls.HelloCustom)
		if err := uTLSConn.ApplyPreset(spec); err != nil {
			return nil, err
		}
		if err := uTLSConn.Handshake(); err != nil {
			return nil, err
		}
		return uTLSConn, nil
	}

	return &http.Transport{DialTLS: dialtls}, nil
}

// stringToSpec creates a ClientHelloSpec based on a JA3 string
func stringToSpec(ja3 string) (*tls.ClientHelloSpec, error) {
	tokens := strings.Split(ja3, ",")

	version := tokens[0]
	ciphers := strings.Split(tokens[1], "-")
	extensions := strings.Split(tokens[2], "-")
	curves := strings.Split(tokens[3], "-")
	if len(curves) == 1 && curves[0] == "" {
		curves = []string{}
	}
	pointFormats := strings.Split(tokens[4], "-")
	if len(pointFormats) == 1 && pointFormats[0] == "" {
		pointFormats = []string{}
	}

	// parse curves
	var targetCurves []tls.CurveID
	for _, c := range curves {
		cid, err := strconv.ParseUint(c, 10, 16)
		if err != nil {
			return nil, err
		}
		targetCurves = append(targetCurves, tls.CurveID(cid))
	}
	extMap["10"] = &tls.SupportedCurvesExtension{Curves: targetCurves}

	// parse point formats
	var targetPointFormats []byte
	for _, p := range pointFormats {
		pid, err := strconv.ParseUint(p, 10, 8)
		if err != nil {
			return nil, err
		}
		targetPointFormats = append(targetPointFormats, byte(pid))
	}
	extMap["11"] = &tls.SupportedPointsExtension{SupportedPoints: targetPointFormats}

	// build extenions list
	var exts []tls.TLSExtension
	for _, e := range extensions {
		te, ok := extMap[e]
		if !ok {
			return nil, ErrExtensionNotExist(e)
		}
		exts = append(exts, te)
	}
	// build SSLVersion
	vid64, err := strconv.ParseUint(version, 10, 16)
	if err != nil {
		return nil, err
	}
	vid := uint16(vid64)

	// build CipherSuites
	var suites []uint16
	for _, c := range ciphers {
		cid, err := strconv.ParseUint(c, 10, 16)
		if err != nil {
			return nil, err
		}
		suites = append(suites, uint16(cid))
	}

	return &tls.ClientHelloSpec{
		TLSVersMin:         vid,
		TLSVersMax:         vid,
		CipherSuites:       suites,
		CompressionMethods: []byte{0},
		Extensions:         exts,
		GetSessionID:       sha256.Sum256,
	}, nil
}

// ErrExtensionNotExist is returned when an extension is not supported by the
// library
type ErrExtensionNotExist string

// Error is the error value which contains the extension that does not exist
func (e ErrExtensionNotExist) Error() string {
   return fmt.Sprintf("Extension does not exist %q", e)
}

// JA3Client contains is similar to http.Client
type JA3Client struct {
   *http.Client
   *tls.Config
}

// NewWithString creates a JA3 client with the specified JA3 string
func NewWithString(ja3 string) (*JA3Client, error) {
   tr, err := NewTransport(ja3)
   if err != nil {
      return nil, err
   }
   client := &http.Client{Transport: tr}
   return &JA3Client{
      client, &tls.Config{},
   }, nil
}
