package protocol

import (
	"encoding/json"
	"fmt"
	"io"
)

type MsgType string

const (
	MsgAuth      MsgType = "AUTH"
	MsgAuthOK    MsgType = "AUTH_OK"
	MsgNewTunnel MsgType = "NEW_TUNNEL"
	MsgDataOpen  MsgType = "DATA_OPEN"
)

type Message struct {
	Type MsgType     `json:"type"`
	Data interface{} `json:"data"`
}

type AuthData struct {
	Secret   string `json:"secret"`
	TunnelID string `json:"tunnel_id,omitempty"`
}

type AuthOKData struct {
	TunnelID   string `json:"tunnel_id"`
	PublicPort int    `json:"public_port"`
}

type NewTunnelData struct {
	TunnelID string `json:"tunnel_id"`
	ConnID   string `json:"conn_id"`
}

type DataOpenData struct {
	TunnelID string `json:"tunnel_id"`
	ConnID   string `json:"conn_id"`
}

func Encode(w io.Writer, msg Message) error {
	data, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("encode: %w", err)
	}
	data = append(data, '\n')
	_, err = w.Write(data)
	return err
}

func Decode(r io.Reader) (Message, error) {
	dec := json.NewDecoder(r)
	var msg Message
	if err := dec.Decode(&msg); err != nil {
		return Message{}, fmt.Errorf("decode: %w", err)
	}

	switch msg.Type {
	case MsgAuth:
		var d AuthData
		b, _ := json.Marshal(msg.Data)
		json.Unmarshal(b, &d)
		msg.Data = d
	case MsgAuthOK:
		var d AuthOKData
		b, _ := json.Marshal(msg.Data)
		json.Unmarshal(b, &d)
		msg.Data = d
	case MsgNewTunnel:
		var d NewTunnelData
		b, _ := json.Marshal(msg.Data)
		json.Unmarshal(b, &d)
		msg.Data = d
	case MsgDataOpen:
		var d DataOpenData
		b, _ := json.Marshal(msg.Data)
		json.Unmarshal(b, &d)
		msg.Data = d
	}

	return msg, nil
}
