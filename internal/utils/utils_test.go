package utils

import (
	"errors"
	"github.com/stretchr/testify/assert"
	"log"
	"testing"
)

type luhnWant struct {
	result bool
}

type luhntestCase struct {
	name   string
	number string
	want   luhnWant
}

var orderNumTests = []luhntestCase{
	{
		name:   "Correct order number",
		number: "7046336413",
		want: luhnWant{
			result: true,
		},
	},
	{
		name:   "Incorrect order number",
		number: "2345",
		want: luhnWant{
			result: false,
		},
	},
	{
		name:   "Empty order number",
		number: "",
		want: luhnWant{
			result: false,
		},
	},
}

func TestIsValidLuhnNumber(t *testing.T) {
	for _, test := range orderNumTests {
		t.Run(test.name, func(t *testing.T) {
			result := IsValidLuhnNumber(test.number)
			assert.Equal(t, result, true)
		})
	}
}

type hashPassWant struct {
	err error
}

type hashPassCase struct {
	name     string
	password string
	want     hashPassWant
}

var genHashPassTests = []hashPassCase{
	{
		name:     "Successful generate hash password",
		password: "test",
		want: hashPassWant{
			err: nil,
		},
	},
}

func TestGenerateHashPassword(t *testing.T) {
	for _, test := range genHashPassTests {
		t.Run(test.name, func(t *testing.T) {
			k, err := GenerateHashPassword(test.password)
			log.Println(k)
			assert.Equal(t, err, nil)
		})
	}
}

type compareHashPassWant struct {
	err error
}

type compareHashPassCase struct {
	name     string
	password string
	hash     string
	want     compareHashPassWant
}

var compareGenHashPassTests = []compareHashPassCase{
	{
		name:     "Successful compare hash and password",
		hash:     "243261243130242f69376e4730457a4f7a43476542377269433343304f7144424f742e513563313971696e6f4751684b4e76355a6c646f5456655636",
		password: "test",
		want: compareHashPassWant{
			err: nil,
		},
	},
	{
		name:     "Unsuccessful compare hash and password",
		hash:     "243261243130242f69376e4730457a4f7a43476542377269433343304f7144424f742e513563313971696e6f4751684b4e76355a6c646f5456655636",
		password: "test1",
		want: compareHashPassWant{
			err: errors.New("crypto/bcrypt: hashedPassword is not the hash of the given password"),
		},
	},
}

func TestCompareHashAndPassword(t *testing.T) {
	for _, test := range compareGenHashPassTests {
		t.Run(test.name, func(t *testing.T) {
			err := CompareHashAndPassword(test.hash, test.password)
			assert.Equal(t, err, test.want.err)
		})
	}
}
