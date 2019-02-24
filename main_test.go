package main

import (
	"encoding/base64"
	"encoding/json"
	"testing"

	"github.com/Xe/bsnk/api"
)

func TestSelectTarget(t *testing.T) {
	cases := []struct {
		name          string
		data          string
		target, immed api.Coord
	}{
		{
			name: "right",
			data: "eyJnYW1lIjp7ImlkIjoiOWFhMGUzZDgtOTcyMS00YTc2LTk3ZjQtNTY5OWNjZDM1ZDI0In0sInR1cm4iOjAsImJvYXJkIjp7ImhlaWdodCI6MTEsIndpZHRoIjoxMSwiZm9vZCI6W3sieCI6NCwieSI6MTB9LHsieCI6NCwieSI6N31dLCJzbmFrZXMiOlt7ImlkIjoiZ3NfcEt4alRRTUs0V2o0Q1ZtN1R4OXdxTWZEIiwibmFtZSI6IlhlIC8gV2l0aGluIiwiaGVhbHRoIjoxMDAsImJvZHkiOlt7IngiOjEsInkiOjF9LHsieCI6MSwieSI6MX0seyJ4IjoxLCJ5IjoxfV19LHsiaWQiOiJnc183ZzM2S2gzVGRHRllHRlhncWN5WHBSeVMiLCJuYW1lIjoieHRhZ29uIC8gTmFnaW5pIiwiaGVhbHRoIjoxMDAsImJvZHkiOlt7IngiOjksInkiOjl9LHsieCI6OSwieSI6OX0seyJ4Ijo5LCJ5Ijo5fV19XX0sInlvdSI6eyJpZCI6ImdzX3BLeGpUUU1LNFdqNENWbTdUeDl3cU1mRCIsIm5hbWUiOiJYZSAvIFdpdGhpbiIsImhlYWx0aCI6MTAwLCJib2R5IjpbeyJ4IjoxLCJ5IjoxfSx7IngiOjEsInkiOjF9LHsieCI6MSwieSI6MX1dfX0=",
			target: api.Coord{
				X: 4,
				Y: 7,
			},
			immed: api.Coord{
				X: 2,
				Y: 1,
			},
		},
		{
			name: "down",
			data: "eyJnYW1lIjp7ImlkIjoiOWFhMGUzZDgtOTcyMS00YTc2LTk3ZjQtNTY5OWNjZDM1ZDI0In0sInR1cm4iOjAsImJvYXJkIjp7ImhlaWdodCI6MTEsIndpZHRoIjoxMSwiZm9vZCI6W3sieCI6MSwieSI6MH0seyJ4Ijo0LCJ5Ijo3fV0sInNuYWtlcyI6W3siaWQiOiJnc19wS3hqVFFNSzRXajRDVm03VHg5d3FNZkQiLCJuYW1lIjoiWGUgLyBXaXRoaW4iLCJoZWFsdGgiOjEwMCwiYm9keSI6W3sieCI6MSwieSI6MX0seyJ4IjoxLCJ5IjoxfSx7IngiOjEsInkiOjF9XX0seyJpZCI6ImdzXzdnMzZLaDNUZEdGWUdGWGdxY3lYcFJ5UyIsIm5hbWUiOiJ4dGFnb24gLyBOYWdpbmkiLCJoZWFsdGgiOjEwMCwiYm9keSI6W3sieCI6OSwieSI6OX0seyJ4Ijo5LCJ5Ijo5fSx7IngiOjksInkiOjl9XX1dfSwieW91Ijp7ImlkIjoiZ3NfcEt4alRRTUs0V2o0Q1ZtN1R4OXdxTWZEIiwibmFtZSI6IlhlIC8gV2l0aGluIiwiaGVhbHRoIjoxMDAsImJvZHkiOlt7IngiOjEsInkiOjF9LHsieCI6MSwieSI6MX0seyJ4IjoxLCJ5IjoxfV19fQ==",
			target: api.Coord{
				X: 1,
				Y: 0,
			},
			immed: api.Coord{
				X: 1,
				Y: 0,
			},
		},
	}

	for _, cs := range cases {
		t.Run(cs.name, func(t *testing.T) {
			decoded, err := base64.StdEncoding.DecodeString(cs.data)
			if err != nil {
				t.Fatal(err)
			}

			var gs api.SnakeRequest
			err = json.Unmarshal(decoded, &gs)
			if err != nil {
				t.Fatal(err)
			}

			target, immed := selectTarget(gs)

			if !target.Eq(cs.target) {
				t.Errorf("wanted target: %s, got: %s", cs.target, target)
			}

			if !immed.Eq(cs.immed) {
				t.Errorf("wanted immed: %s, got: %s", cs.immed, immed)
			}
		})
	}
}
