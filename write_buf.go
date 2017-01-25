package pgmock

import (
	"encoding/binary"
)

func newWriteBuf(t byte) *WriteBuf {
	buf := []byte{t, 0, 0, 0, 0}
	return &WriteBuf{buf: buf, sizeIdx: 1}
}

// WriteBuf is used build messages in the PostgreSQL wire format
type WriteBuf struct {
	buf     []byte
	sizeIdx int
}

func (wb *WriteBuf) startMsg(t byte) {
	wb.closeMsg()
	wb.buf = append(wb.buf, t, 0, 0, 0, 0)
	wb.sizeIdx = len(wb.buf) - 4
}

func (wb *WriteBuf) closeMsg() {
	binary.BigEndian.PutUint32(wb.buf[wb.sizeIdx:wb.sizeIdx+4], uint32(len(wb.buf)-wb.sizeIdx))
}

func (wb *WriteBuf) WriteByte(b byte) {
	wb.buf = append(wb.buf, b)
}

func (wb *WriteBuf) WriteCString(s string) {
	wb.buf = append(wb.buf, []byte(s)...)
	wb.buf = append(wb.buf, 0)
}

func (wb *WriteBuf) WriteInt16(n int16) {
	b := make([]byte, 2)
	binary.BigEndian.PutUint16(b, uint16(n))
	wb.buf = append(wb.buf, b...)
}

func (wb *WriteBuf) WriteUint16(n uint16) {
	b := make([]byte, 2)
	binary.BigEndian.PutUint16(b, n)
	wb.buf = append(wb.buf, b...)
}

func (wb *WriteBuf) WriteInt32(n int32) {
	b := make([]byte, 4)
	binary.BigEndian.PutUint32(b, uint32(n))
	wb.buf = append(wb.buf, b...)
}

func (wb *WriteBuf) WriteUint32(n uint32) {
	b := make([]byte, 4)
	binary.BigEndian.PutUint32(b, n)
	wb.buf = append(wb.buf, b...)
}

func (wb *WriteBuf) WriteInt64(n int64) {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, uint64(n))
	wb.buf = append(wb.buf, b...)
}

func (wb *WriteBuf) WriteBytes(b []byte) {
	wb.buf = append(wb.buf, b...)
}
