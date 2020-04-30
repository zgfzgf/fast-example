package main

import (
	"bytes"
	"fmt"
	fast "github.com/co11ter/goFAST"
	"strings"
)

var xmlMessageTemplate = `
<?xml version="1.0" encoding="UTF-8"?>
<templates xmlns="http://www.fixprotocol.org/ns/fast/td/1.1">
   <template name="Done" id="1" xmlns="http://www.fixprotocol.org/ns/fast/td/1.1">
      <int64 name="Id" id="11"  />
      <string name="Name" id="12" />
      <sequence name="LogMatch">
         <length name="MatchLength" id="151"/>
         <int64 name="OrderId" id="21"/>
            <string name="Side" id="22"/>
            <int64 name="TradeId" id="23"/>
      </sequence>
      <sequence name="LogDone">
         <length name="DoneLength" id="152"/>
         <int64 name="OrderId" id="31"/>
            <string name="Side" id="32"/>
            <string name="Done" id="33" />
      </sequence>
   </template>
</templates>`

type Match struct {
	OrderId int64
	Side    string
	TradeId int64
}

type Done struct {
	OrderId int64
	Side    string
	Done    string
}

type Message struct {
	Id        int64
	Name      string
	Matchs    []Match
	Dones     []Done
	seqLocked bool
	seqIndex  int
}

func (msg *Message) GetTemplateID() uint {
	return 1
}

func (msg *Message) GetLength(field *fast.Field) {
	if field.Name == "LogMatch" {
		field.Value = len(msg.Matchs)
	} else if field.Name == "LogDone" {
		field.Value = len(msg.Dones)
	}
}

func (msg *Message) SetTemplateID(tid uint) {

}

func (msg *Message) SetLength(field *fast.Field) {
	if field.Name == "LogMatch" && len(msg.Matchs) < field.Value.(int) {
		msg.Matchs = make([]Match, field.Value.(int))
	} else if field.Name == "LogDone" && len(msg.Dones) < field.Value.(int) {
		msg.Dones = make([]Done, field.Value.(int))
	}
}

func (msg *Message) Lock(field *fast.Field) bool {
	msg.seqLocked = field.Name == "LogMatch" || field.Name == "LogDone"
	if msg.seqLocked {
		msg.seqIndex = field.Value.(int)
	}
	return msg.seqLocked
}

func (msg *Message) Unlock() {
	msg.seqLocked = false
	msg.seqIndex = 0
}

func (msg *Message) GetValue(field *fast.Field) {
	switch field.ID {
	case 11:
		field.Value = msg.Id
	case 12:
		field.Value = msg.Name
	case 21:
		field.Value = msg.Matchs[msg.seqIndex].OrderId
	case 22:
		field.Value = msg.Matchs[msg.seqIndex].Side
	case 23:
		field.Value = msg.Matchs[msg.seqIndex].TradeId
	case 31:
		field.Value = msg.Dones[msg.seqIndex].OrderId
	case 32:
		field.Value = msg.Dones[msg.seqIndex].Side
	case 33:
		field.Value = msg.Dones[msg.seqIndex].Done
		//case 31:
		// if msg.Sequence[msg.seqIndex].Done !=""{
		//    field.Value = msg.Sequence[msg.seqIndex].Done
		// }
	}
}

func (msg *Message) SetValue(field *fast.Field) {
	switch field.ID {
	case 11:
		msg.Id = field.Value.(int64)
	case 12:
		msg.Name = field.Value.(string)
	case 21:
		msg.Matchs[msg.seqIndex].OrderId = field.Value.(int64)
	case 22:
		msg.Matchs[msg.seqIndex].Side = field.Value.(string)
	case 23:
		msg.Matchs[msg.seqIndex].TradeId = field.Value.(int64)
	case 31:
		msg.Dones[msg.seqIndex].OrderId = field.Value.(int64)
	case 32:
		msg.Dones[msg.seqIndex].Side = field.Value.(string)
	case 33:
		msg.Dones[msg.seqIndex].Done = field.Value.(string)
	}

}

func main() {
	var buf bytes.Buffer
	var msg = Message{
		Id:   1001,
		Name: "Name",
		//Matchs: []Match {
		// {OrderId: 11, Side:"buy", TradeId:22},
		// {OrderId: 12, Side:"buy", TradeId:23},
		// //{OrderId: 12, Side:"buy"},
		//},
		Dones: []Done{
			{OrderId: 11, Side: "buy", Done: "done"},
			{OrderId: 12, Side: "buy", Done: "done"},
		},
	}

	tpls, err := fast.ParseXMLTemplate(strings.NewReader(xmlMessageTemplate))
	if err != nil {
		panic(err)
	}
	encoder := fast.NewEncoder(&buf, tpls...)

	if err := encoder.Encode(&msg); err != nil {
		panic(err)
	}
	fmt.Printf("%x", buf.Bytes())

	var receive Message
	reader := bytes.NewReader(
		buf.Bytes(),
	)

	decoder := fast.NewDecoder(
		reader,
		tpls...,
	)

	if err := decoder.Decode(&receive); err != nil {
		panic(err)
	}
	fmt.Print(receive)

}
