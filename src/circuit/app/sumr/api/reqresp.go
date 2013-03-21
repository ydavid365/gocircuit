package api

import (
	"circuit/app/sumr"
	"encoding/json"
	"io"
)

// Response is the common response object
type response struct {
	Sum float64 `json:"sum"`
}

// ReadRequestBatchFunc reads a request batch
type readRequestBatchFunc func(io.Reader) ([]interface{}, error)

type addRequest struct {
	change *change
}

func (r *addRequest) Key() sumr.Key {
	return r.change.Key()
}

func (r *addRequest) Value() float64 {
	return r.change.Value
}

func readAddRequest(dec *json.Decoder) (interface{}, error) {
	change, err := readChange(dec)
	if err != nil {
		return nil, err
	}
	return &addRequest{change: change}, nil
}

// readAddRequestBatch parses the body of an HTTP request for a batch of ADD requests.
// ADD requests are concatenated together, optionally separated by whitespace characters.
// Each individual ADD request is of the form:
//
//	{"t":12345, "k":{"p":"q", "r":"s"}, "v":023}
//
func readAddRequestBatch(r io.Reader) ([]interface{}, error) {
	dec := json.NewDecoder(r)
	var bch []interface{}
	for {
		r, err := readAddRequest(dec)
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}
		bch = append(bch, r)
	}
	return bch, nil
}

// A SumRequest returns the sum of all changes at a given feature
// On the wire, it looks like so
//
//	{
//		"k": { "fkey": "fvalue", ... },
//	}
//
type sumRequest struct {
	feature feature
}

func (r *sumRequest) Key() sumr.Key {
	return r.feature.Key()
}

func readSumRequest(dec *json.Decoder) (interface{}, error) {
	b := make(map[string]interface{})
	if err := dec.Decode(&b); err != nil {
		return nil, err
	}
	return makeSumRequestMap(b)
}

// readAddRequestBatch parses the body of an HTTP request for a batch of SUM requests.
// SUM requests are concatenated together, optionally separated by whitespace characters.
// Each individual SUM request is of the form:
//
//	{ "k":{"p":"q", "r":"s"} }
//
func readSumRequestBatch(r io.Reader) ([]interface{}, error) {
	dec := json.NewDecoder(r)
	var bch []interface{}
	for {
		r, err := readSumRequest(dec)
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}
		bch = append(bch, r)
	}
	return bch, nil
}

func makeSumRequestMap(b map[string]interface{}) (*sumRequest, error) {
	// Read feature
	feature_, ok := b["k"]
	if !ok {
		return nil, ErrNoFeature
	}
	feature, ok := feature_.(map[string]interface{})
	if !ok {
		return nil, ErrNoFeature
	}
	f, err := makeFeatureMap(feature)
	if err != nil {
		return nil, err
	}

	// Done
	return &sumRequest{feature: f}, nil
}