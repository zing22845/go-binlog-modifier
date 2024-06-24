package binlog_modifier

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"hash/crc32"
	"io"

	"github.com/go-mysql-org/go-mysql/replication"
)

type BinlogModifier struct {
	Reader           io.Reader // must read from the header of a binlog file
	WriterAt         io.WriterAt
	IsModifyPosition bool
	DeltaPosition    int64
	OnEventFunc      func(event *replication.BinlogEvent) error
}

func (bm *BinlogModifier) InitOnEventFunc(isCheckForeignKey bool) {
	bm.OnEventFunc = func(event *replication.BinlogEvent) error {
		switch event.Event.(type) {
		case *replication.FormatDescriptionEvent:
		case *replication.QueryEvent:
			if !isCheckForeignKey {
				qe := event.Event.(*replication.QueryEvent)
				if qe.StatusVars[0] == Q_FLAGS2_CODE {
					idx := bytes.Index(event.RawData, qe.StatusVars)
					// modify FK check flag
					flags2 := binary.LittleEndian.Uint32(qe.StatusVars[1:])
					flags2 |= OPTION_NO_FOREIGN_KEY_CHECKS
					binary.LittleEndian.PutUint32(event.RawData[idx+1:], flags2)
					// modify checksum
					length := len(event.RawData)
					checksum := crc32.ChecksumIEEE(event.RawData[:length-replication.BinlogChecksumLength])
					binary.LittleEndian.PutUint32(event.RawData[length-replication.BinlogChecksumLength:], checksum)
				}
			}
		default:
		}
		_, err := bm.WriterAt.WriteAt(event.RawData, int64(event.Header.LogPos-event.Header.EventSize))
		return err
	}
	bm.IsModifyPosition = false
	bm.DeltaPosition = 0
}

func (bm *BinlogModifier) Run() (err error) {
	parser := replication.NewBinlogParser()
	// read file header
	fh := make([]byte, 4)
	_, err = bm.Reader.Read(fh)
	if err != nil {
		return err
	}
	if !bytes.Equal(fh, replication.BinLogFileHeader) {
		return fmt.Errorf("invalid binlog file header, expect %v but got %v", replication.BinLogFileHeader, fh)
	}
	// write file header
	_, err = bm.WriterAt.WriteAt(fh, 0)
	if err != nil {
		return err
	}
	// copy format description event
	isEOF, err := parser.ParseSingleEvent(bm.Reader, bm.OnEventFunc)
	if err != nil {
		return err
	}
	if isEOF {
		return nil
	}
	return parser.ParseReader(bm.Reader, bm.OnEventFunc)
}
