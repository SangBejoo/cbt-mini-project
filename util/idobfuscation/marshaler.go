package idobfuscation

import (
	"encoding/json"
	"io"

	"cbt-test-mini-project/util/idcodec"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
)

type CustomJSONMarshaler struct {
	runtime.JSONPb
}

func (c *CustomJSONMarshaler) NewDecoder(r io.Reader) runtime.Decoder {
	return runtime.DecoderFunc(func(v interface{}) error {
		data, err := io.ReadAll(r)
		if err != nil {
			return err
		}

		if len(data) == 0 {
			return c.JSONPb.Unmarshal(data, v)
		}

		var genericData interface{}
		if err := json.Unmarshal(data, &genericData); err != nil {
			return c.JSONPb.Unmarshal(data, v)
		}

		walkAndDecodeIDs(genericData)

		cleanData, err := json.Marshal(genericData)
		if err != nil {
			return err
		}

		return c.JSONPb.Unmarshal(cleanData, v)
	})
}

func (c *CustomJSONMarshaler) Unmarshal(data []byte, v interface{}) error {
	var genericData interface{}
	if err := json.Unmarshal(data, &genericData); err != nil {
		return c.JSONPb.Unmarshal(data, v)
	}

	walkAndDecodeIDs(genericData)

	cleanData, err := json.Marshal(genericData)
	if err != nil {
		return err
	}
	return c.JSONPb.Unmarshal(cleanData, v)
}

func walkAndDecodeIDs(v interface{}) {
	switch val := v.(type) {
	case map[string]interface{}:
		for k, mapVal := range val {
			walkAndDecodeIDs(mapVal)

			if idcodec.IsIDField(k) {
				if strVal, ok := mapVal.(string); ok {
					if !isNumeric(strVal) {
						if decoded, err := idcodec.Decode(strVal); err == nil {
							val[k] = decoded
						}
					}
				}
			}
		}
	case []interface{}:
		for _, item := range val {
			walkAndDecodeIDs(item)
		}
	}
}
