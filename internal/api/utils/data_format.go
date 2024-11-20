package utils

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"github.com/ercross/payment_gateways/internal/api/v1/dto"
	"net/http"
	"sort"
	"strings"
)

var dataTypeTextToDataType = map[string]dto.DataFormat{
	"application/json":     dto.DataFormatJSON,
	"text/xml":             dto.DataFormatXML,
	"application/soap+xml": dto.DataFormatXML,
	"application/xml":      dto.DataFormatXML,
}

type DecodableRequest interface {
	IsDecodable() bool
}

// DecodeRequest decodes incoming request based on content type
func DecodeRequest(r *http.Request, request DecodableRequest) error {
	contentType := r.Header.Get("Content-Type")
	dataType := dataTypeFromRequestContentType(contentType)
	switch dataType {
	case dto.DataFormatJSON:
		return json.NewDecoder(r.Body).Decode(request)
	case dto.DataFormatXML:
		return xml.NewDecoder(r.Body).Decode(request)
	default:
		return fmt.Errorf("unsupported content type")
	}
}

func dataTypeFromRequestContentType(contentType string) dto.DataFormat {
	switch strings.TrimSpace(contentType) {
	case "application/json":
		return dto.DataFormatJSON
	case "text/xml", "application/soap+xml", "application/xml":
		return dto.DataFormatXML
	default:
		return dto.DataFormatJSON
	}
}

// DetermineResponseContentDataType use Accept or Content-Type header value to determine
// which dto.DataFormat is more preferred by client
func DetermineResponseContentDataType(r *http.Request) dto.DataFormat {
	acceptHeader := r.Header.Get("Accept")
	if acceptHeader == "" {
		return dataTypeFromRequestContentType(r.Header.Get("Content-Type"))
	}

	type MIMEType struct {
		Type    dto.DataFormat
		Quality float64
	}

	parts := strings.Split(acceptHeader, ",")
	if len(parts) == 0 {
		parts = strings.Split(acceptHeader, " ")
	}
	if len(parts) == 0 {

		if strings.TrimSpace(acceptHeader) == "*/*" {
			return dataTypeFromRequestContentType(r.Header.Get("Content-Type"))
		}
		return dto.DataFormatJSON
	}

	var mimeTypes []MIMEType
	defaultQuality := 1.0

	for _, part := range parts {
		part = strings.TrimSpace(part)
		mimeAndQuality := strings.Split(part, ";q=")
		var mimeType MIMEType
		if len(mimeAndQuality) == 1 {
			if dt, ok := dataTypeTextToDataType[strings.TrimSpace(mimeAndQuality[0])]; ok {
				mimeType.Type = dt
				mimeType.Quality = defaultQuality
				mimeTypes = append(mimeTypes, mimeType)
			}
		} else if len(mimeAndQuality) == 2 {
			if dt, ok := dataTypeTextToDataType[strings.TrimSpace(mimeAndQuality[0])]; ok {
				mimeType.Type = dt
				fmt.Sscanf(mimeAndQuality[1], "%f", &mimeType.Quality)
				mimeTypes = append(mimeTypes, mimeType)
			}
		}
	}

	if len(mimeTypes) == 0 {
		return dto.DataFormatJSON
	}

	// Sort by quality in descending order
	sort.Slice(mimeTypes, func(i, j int) bool {
		return mimeTypes[i].Quality > mimeTypes[j].Quality
	})

	return mimeTypes[0].Type
}
