package convert

import "encoding/xml"

type TextRes struct {
	XMLName      xml.Name `xml:"xml"`
	ToUserName   string   `xml:"ToUserName"`
	FromUserName string   `xml:"FromUserName"`
	CreateTime   int64    `xml:"CreateTime"`
	MsgType      string   `xml:"MsgType"`
	Content      string   `xml:"Content"`
}

func ToTextRes(body []byte) *TextRes {
	var msg TextRes
	err := xml.Unmarshal(body, &msg)
	if err != nil {
		panic(err)
	}
	return &msg
}

func (msg *TextRes) ToXml() []byte {
	body, err := xml.Marshal(msg)
	if err != nil {
		panic(err)
	}
	return body
}
