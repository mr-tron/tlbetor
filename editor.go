package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"reflect"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
	"github.com/tonkeeper/tongo/abi"
	"github.com/tonkeeper/tongo/boc"
	"github.com/tonkeeper/tongo/tlb"
	"golang.org/x/exp/maps"
)

var opcode *uint32

func tlbDecodingWindow() *fyne.Container {

	cellsForm := widget.NewEntry()
	cellsForm.MultiLine = true
	cellsForm.PlaceHolder = ""

	bocForm := widget.NewEntry()
	bocForm.PlaceHolder = "base64 or hex encoded bocForm"
	bocForm.MultiLine = true
	bocForm.Wrapping = fyne.TextWrapBreak

	selector := widget.NewLabel("Selector:")
	decodeType := widget.NewSelect(maps.Keys(abi.KnownMsgInTypes), func(s string) {
		if opcode != nil {
			selector.SetText(fmt.Sprintf("Selector: %v", *opcode))
		} else {
			selector.SetText("Selector: ")
		}
	})
	decodedForm := widget.NewEntry()
	decodedForm.MultiLine = true

	var lock string
	bocForm.OnChanged = bocDecoding(&lock, cellsForm, decodedForm, decodeType)
	cellsForm.OnChanged = cellsDecoding(&lock, bocForm)
	decodedForm.OnChanged = decodedChanged(&lock, bocForm, decodeType)

	tlbWorkingLayout := container.New(layout.NewGridLayout(1),
		bocForm,
		cellsForm,
		container.NewBorder(
			container.New(layout.NewHBoxLayout(), selector, decodeType),
			nil,
			nil,
			nil,
			decodedForm,
		),
	)
	return tlbWorkingLayout
}

func decodedChanged(lock *string, bocForm *widget.Entry, decodeType *widget.Select) func(string) {
	return func(s string) {
		if *lock == "decoded" {
			return
		}
		if *lock == "" {
			*lock = "decoded"
			defer func() { *lock = "" }()
		}
		t, prs := abi.KnownMsgInTypes[decodeType.Selected]
		if !prs {
			fmt.Println("not found type", decodeType.Selected)
			return
		}
		p := reflect.New(reflect.TypeOf(t))
		if err := json.Unmarshal([]byte(s), p.Interface()); err != nil {
			fmt.Println("json", err.Error())
			return
		}

		if opcode == nil {
			fmt.Println("nil opcode")
			return
		}
		c := boc.NewCell()
		c.WriteUint(uint64(*opcode), 32)
		err := tlb.Marshal(c, p.Elem().Interface())
		if err != nil {
			fmt.Println("tlb", err.Error())
			return
		}
		s, err = c.ToBocBase64()
		if err != nil {
			fmt.Println("boc", err.Error())
			return
		}
		bocForm.SetText(s)
	}
}

func cellsDecoding(lock *string, bocForm *widget.Entry) func(string) {
	return func(s string) {
		if *lock == "cells" {
			return
		}
		if *lock == "" {
			*lock = "cells"
			defer func() { *lock = "" }()
		}
		cells, err := fromStdHexString(s)
		if err != nil {
			bocForm.SetText(fmt.Sprintf("can't parse cells: %v", "err"))
			return
		}
		s, err = cells[0].ToBocBase64()
		bocForm.SetText(s)
	}
}

func bocDecoding(lock *string, cellsForm, decodeForm *widget.Entry, decodedType *widget.Select) func(string) {
	return func(s string) {
		if *lock == "boc" {
			return
		}
		if *lock == "" {
			*lock = "boc"
			defer func() { *lock = "" }()
		}
		c, err := boc.DeserializeSinglRootBase64(s)
		if err != nil {
			cells, err := boc.DeserializeBocHex(s)
			if err != nil {
				cellsForm.SetText("can't decode boc")
				return
			}
			c = cells[0]
		}
		cellsForm.SetText(c.ToString())
		op, operation, value, err := abi.InternalMessageDecoder(c, nil)
		if err != nil {
			decodeForm.SetText(fmt.Sprintf("can't decode boc: %s", err))
			return
		}
		if operation != nil {
			opcode = op
			decodedType.SetSelected(*operation)
		}
		buffer := bytes.NewBuffer(nil)
		encoder := json.NewEncoder(buffer)
		encoder.SetIndent("", "  ")
		err = encoder.Encode(value)
		if err != nil {
			fmt.Println(err)
			return
		}
		decodeForm.SetText(buffer.String())
	}
}
