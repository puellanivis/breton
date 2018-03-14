package m3u8

import (
	"reflect"
	"testing"
	"time"
)

func TestMarshalKey(t *testing.T) {
	key := &Key{
		Method: "NONE",
		URI:    "scheme://host:port/path?query#fragment",

		KeyFormatVersions: []int{2, 3, 5, 7},
	}

	s, err := marshalAttributeList(key)
	if err != nil {
		t.Fatalf("%+v", err)
	}

	expected := `METHOD=NONE,URI="scheme://host:port/path?query#fragment",KEYFORMATVERSIONS="2/3/5/7"`
	if s != expected {
		t.Errorf("expected; but got:\n\t%s\n\t%s", expected, s)
	}

	test := new(Key)

	if err := unmarshalAttributeList(test, []byte(expected)); err != nil {
		t.Fatalf("%+v", err)
	}

	if !reflect.DeepEqual(test, key) {
		t.Errorf("unmarshal failed, expected; but got:\n\t%#v\n\t%#v", key, test)
	}
}

func TestMarshalStreamInf(t *testing.T) {
	sinf := &StreamInf{
		Bandwidth:  1234567,
		Codecs:     []string{"codec1", `"quoted text"`},
		Resolution: Resolution{Width: 1024, Height: 768},
		FrameRate:  50,
		Audio:      "audio-group-id",
	}

	s, err := marshalAttributeList(sinf)
	if err != nil {
		t.Fatalf("%+v", err)
	}

	expected := `BANDWIDTH=1234567,CODECS="codec1,\"quoted text\"",RESOLUTION=1024x768,FRAME-RATE=50,AUDIO="audio-group-id"`
	if s != expected {
		t.Errorf("expected; but got:\n\t%s\n\t%s", expected, s)
	}

	test := new(StreamInf)

	if err := unmarshalAttributeList(test, []byte(expected)); err != nil {
		t.Fatalf("%+v", err)
	}

	if !reflect.DeepEqual(test, sinf) {
		t.Errorf("unmarshal failed, expected; but got:\n\t%#v\n\t%#v", sinf, test)
	}

	sinf.FrameRate = 59.997
	s, err = marshalAttributeList(sinf)
	if err != nil {
		t.Fatalf("%+v", err)
	}

	expected = `BANDWIDTH=1234567,CODECS="codec1,\"quoted text\"",RESOLUTION=1024x768,FRAME-RATE=59.997,AUDIO="audio-group-id"`
	if s != expected {
		t.Errorf("expected; but got:\n\t%s\n\t%s", expected, s)
	}

	test = new(StreamInf)

	if err := unmarshalAttributeList(test, []byte(expected)); err != nil {
		t.Fatalf("%+v", err)
	}

	if !reflect.DeepEqual(test, sinf) {
		t.Errorf("unmarshal failed, expected; but got:\n\t%#v\n\t%#v", sinf, test)
	}
}

func TestMarshalDateRange(t *testing.T) {
	dr := &DateRange{
		ID:        "identification",
		StartDate: time.Date(2018, 03, 07, 15, 52, 36, 0, time.UTC),
		Duration:  5 * time.Second,
		ClientAttribute: map[string]interface{}{
			"STR": `"XYZ123"`,
			"INT": 42,
			"HEX": []byte{ 2, 3, 5, 7 },
		},
	}

	s, err := marshalAttributeList(dr)
	if err != nil {
		t.Fatalf("%+v", err)
	}

	expected := `ID="identification",START-DATE=2018-03-07T15:52:36Z,DURATION=5,X-HEX=0x02030507,X-INT=42,X-STR="\"XYZ123\""`
	if s != expected {
		t.Errorf("expected; but got:\n\t%s\n\t%s", expected, s)
	}

	test := new(DateRange)

	if err := unmarshalAttributeList(test, []byte(expected)); err != nil {
		t.Fatalf("%+v", err)
	}

	test.ClientAttribute = nil
	testDR := *dr
	testDR.ClientAttribute = nil
	if !reflect.DeepEqual(test, &testDR) {
		t.Errorf("unmarshal failed, expected; but got:\n\t%#v\n\t%#v", &testDR, test)
	}

	dr.EndOnNext = true
	s, err = marshalAttributeList(dr)
	if err != nil {
		t.Fatalf("%+v", err)
	}

	expected = `ID="identification",START-DATE=2018-03-07T15:52:36Z,DURATION=5,END-ON-NEXT=YES,X-HEX=0x02030507,X-INT=42,X-STR="\"XYZ123\""`
	if s != expected {
		t.Errorf("expected; but got:\n\t%s\n\t%s", expected, s)
	}

	test = new(DateRange)

	if err := unmarshalAttributeList(test, []byte(expected)); err != nil {
		t.Fatalf("%+v", err)
	}

	test.ClientAttribute = nil
	testDR = *dr
	testDR.ClientAttribute = nil
	if !reflect.DeepEqual(test, &testDR) {
		t.Errorf("unmarshal failed, expected; but got:\n\t%#v\n\t%#v", &testDR, test)
	}
}

func TestMarshalMedia(t *testing.T) {
	m := &Media{
		Type:       "AUDIO",
		GroupID:    "audio-mp4a.40.2",
		Name:       "Français",
		Default:    true,
		Autoselect: true,
		Language:   "fr",
		URI:        "path?query#fragment",
	}

	s, err := marshalAttributeList(m)
	if err != nil {
		t.Fatalf("%+v", err)
	}

	expected := `TYPE=AUDIO,GROUP-ID="audio-mp4a.40.2",NAME="Français",DEFAULT=YES,AUTOSELECT=YES,LANGUAGE="fr",URI="path?query#fragment"`
	if s != expected {
		t.Errorf("expected; but got:\n\t%s\n\t%s", expected, s)
	}

	test := new(Media)

	if err := unmarshalAttributeList(test, []byte(expected)); err != nil {
		t.Fatalf("%+v", err)
	}

	if !reflect.DeepEqual(test, m) {
		t.Errorf("unmarshal failed, expected; but got:\n\t%#v\n\t%#v", m, test)
	}

	m.Channels = []int{7, 2}

	s, err = marshalAttributeList(m)
	if err != nil {
		t.Fatalf("%+v", err)
	}

	expected = `TYPE=AUDIO,GROUP-ID="audio-mp4a.40.2",NAME="Français",DEFAULT=YES,AUTOSELECT=YES,LANGUAGE="fr",CHANNELS="7/2",URI="path?query#fragment"`
	if s != expected {
		t.Errorf("expected; but got:\n\t%s\n\t%s", expected, s)
	}

	test = new(Media)

	if err := unmarshalAttributeList(test, []byte(expected)); err != nil {
		t.Fatalf("%+v", err)
	}

	if !reflect.DeepEqual(test, m) {
		t.Errorf("unmarshal failed, expected; but got:\n\t%#v\n\t%#v", m, test)
	}
}
