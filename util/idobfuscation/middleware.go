package idobfuscation

import (
	"fmt"
	"net/http"
	"strconv"

	"cbt-test-mini-project/util/idcodec"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
)

func PathParamDecodingMiddleware() runtime.Middleware {
	return func(next runtime.HandlerFunc) runtime.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request, pathParams map[string]string) {
			for key, val := range pathParams {
				if !idcodec.IsIDField(key) {
					continue
				}
				if isNumeric(val) {
					continue
				}
				decoded, err := idcodec.Decode(val)
				if err != nil {
					http.Error(w, fmt.Sprintf(`{"status":false,"message":"invalid %s in path","code":"INVALID_ARGUMENT"}`, key), http.StatusBadRequest)
					return
				}
				pathParams[key] = strconv.FormatInt(decoded, 10)
			}

			query := r.URL.Query()
			queryModified := false
			for key, values := range query {
				if !idcodec.IsIDField(key) {
					continue
				}
				for i, val := range values {
					if isNumeric(val) {
						continue
					}
					decoded, err := idcodec.Decode(val)
					if err != nil {
						http.Error(w, fmt.Sprintf(`{"status":false,"message":"invalid %s in query","code":"INVALID_ARGUMENT"}`, key), http.StatusBadRequest)
						return
					}
					query[key][i] = strconv.FormatInt(decoded, 10)
					queryModified = true
				}
			}

			if queryModified {
				r.URL.RawQuery = query.Encode()
			}

			next(w, r, pathParams)
		}
	}
}

func isNumeric(s string) bool {
	if s == "" {
		return false
	}
	for _, c := range s {
		if c < '0' || c > '9' {
			return false
		}
	}
	return true
}

func EncodeIDsInJSON(fieldName string, value int64) interface{} {
	if !idcodec.IsIDField(fieldName) || value == 0 {
		return value
	}
	encoded := idcodec.MustEncode(value)
	if encoded == "" {
		return value
	}
	return encoded
}

func IsIDField(fieldName string) bool {
	return idcodec.IsIDField(fieldName)
}
