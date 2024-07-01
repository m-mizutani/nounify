package testutil

import (
	"encoding/json"
	"testing"

	"github.com/m-mizutani/gt"
)

func Transcode(t *testing.T, dst, src any) {
	t.Helper()

	raw := gt.R1(json.Marshal(src)).NoError(t)
	gt.NoError(t, json.Unmarshal(raw, dst))
}

func DecodeJSON(t *testing.T, src []byte) any {
	t.Helper()

	var data any
	gt.NoError(t, json.Unmarshal(src, &data))
	return data
}
