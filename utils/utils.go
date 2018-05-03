package utils

import (
	"errors"
	"fmt"
	"strconv"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"time"
)


func FloatFromString(raw interface{}) (float64, error) {
	str, ok := raw.(string)
	if !ok {
		return 0, errors.New(fmt.Sprintf("unable to parse, value not string: %T", raw))
	}
	flt, err := strconv.ParseFloat(str, 64)
	if err != nil {
		return 0, errors.New(fmt.Sprintf("unable to parse as float: %s", str))
	}
	return flt, nil
}



// HmacSigner uses HMAC SHA256 for signing payloads.
type HmacSigner struct {
	Key []byte
}

// Sign signs provided payload and returns encoded string sum.
func (hs *HmacSigner) Sign(payload []byte) string {
	mac := hmac.New(sha256.New, hs.Key)
	mac.Write(payload)
	return hex.EncodeToString(mac.Sum(nil))
}

func TimeFromUnixTimestampFloat(raw interface{}) (time.Time, error) {
	ts, ok := raw.(float64)
	if !ok {
		return time.Time{}, errors.New(fmt.Sprintf("unable to parse, value not int64: %T", raw))
	}
	return time.Unix(0, int64(ts)*int64(time.Millisecond)), nil
}