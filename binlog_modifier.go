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
	Reader               io.Reader // must read from the header of a binlog file
	WriterAt             io.WriterAt
	IsVerifyChecksum     bool
	ChecksumAlgorithm    byte
	IsModifyPosition     bool
	DeltaPosition        int64
	OnEventFunc          func(event *replication.BinlogEvent) error
	WriteSize            int64
	format               *replication.FormatDescriptionEvent
	tableIDSize          int
	rowsEventFlagsOffset int
}

func (bm *BinlogModifier) DisableForeignKeyChecks() {
	bm.OnEventFunc = func(event *replication.BinlogEvent) error {
		switch e := event.Event.(type) {
		case *replication.FormatDescriptionEvent:
			bm.format = e
			bm.tableIDSize = 6
			if bm.format.EventTypeHeaderLengths[replication.TABLE_MAP_EVENT-1] == 6 {
				bm.tableIDSize = 4
			}
			bm.ChecksumAlgorithm = e.ChecksumAlgorithm
			bm.rowsEventFlagsOffset = replication.EventHeaderSize + bm.tableIDSize
		case *replication.QueryEvent:
			if e.StatusVars[0] == Q_FLAGS2_CODE {
				// modify FK check flag
				flags2 := binary.LittleEndian.Uint32(e.StatusVars[1:])
				flags2 |= OPTION_NO_FOREIGN_KEY_CHECKS
				binary.LittleEndian.PutUint32(event.RawData[QUERY_EVENT_STATUS_VARS_FIX_OFFSET+1:], flags2)
				// modify checksum
				bm.ModifyChecksum(event)
			}
		case *replication.TableMapEvent:
			e.Flags |= TM_REFERRED_FK_DB_F

			binary.LittleEndian.PutUint16(event.RawData[bm.rowsEventFlagsOffset:], e.Flags)
			// modify checksum
			bm.ModifyChecksum(event)
		case *replication.RowsEvent:
			e.Flags |= TM_REFERRED_FK_DB_F
			binary.LittleEndian.PutUint16(event.RawData[bm.rowsEventFlagsOffset:], e.Flags)
			// modify checksum
			bm.ModifyChecksum(event)
		default:
		}
		n, err := bm.WriterAt.WriteAt(event.RawData, int64(event.Header.LogPos-event.Header.EventSize))
		if err != nil {
			return err
		}
		bm.WriteSize += int64(n)
		return nil
	}
	bm.IsModifyPosition = false
	bm.DeltaPosition = 0
}

func (bm *BinlogModifier) Run() (err error) {
	parser := replication.NewBinlogParser()
	parser.SetVerifyChecksum(bm.IsVerifyChecksum)
	parser.SetRowsEventDecodeFunc(SkipRowsEventDecodeBody)
	parser.SetTableMapOptionalMetaDecodeFunc(SkipTableMapOptionalMetaDecode)
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
	n, err := bm.WriterAt.WriteAt(fh, 0)
	if err != nil {
		return err
	}
	bm.WriteSize += int64(n)
	return parser.ParseReader(bm.Reader, bm.OnEventFunc)
}

func (bm *BinlogModifier) ModifyChecksum(event *replication.BinlogEvent) {
	if bm.ChecksumAlgorithm == 1 {
		length := len(event.RawData)
		checksum := crc32.ChecksumIEEE(event.RawData[:length-replication.BinlogChecksumLength])
		binary.LittleEndian.PutUint32(event.RawData[length-replication.BinlogChecksumLength:], checksum)
	}
}

func SkipRowsEventDecodeBody(re *replication.RowsEvent, date []byte) (err error) {
	_, err = re.DecodeHeader(date)
	return err
}

func SkipTableMapOptionalMetaDecode([]byte) error {
	return nil
}
