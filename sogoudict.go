// Read content of a Sogou Pinyin Dictionary File (.scel).
package sogoudict

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"
	"os"
	"sort"
	"unicode/utf16"
	"unicode/utf8"
)

var (
	// not a .scel file at all
	ErrInvalidDict = errors.New("not a valid sogou dict")

	// .scel file might be corrupted
	ErrCorruptedDict = errors.New("dict file might be corrupted")
)

var (
	_SogouDictID   = []byte{64, 21, 0, 0, 68, 67, 83, 1, 1, 0, 0, 0}
	_PinyinTableID = []byte{157, 1, 0, 0}
	_Uint16        = binary.LittleEndian.Uint16
)

const (
	_NameOffset, _NameSize               = 304, 520
	_CategoryOffset, _CategorySize       = 824, 520
	_DescriptionOffset, _DescriptionSize = 1344, 2048
	_ExamplesOffset, _ExamplesSize       = 3392, 2048

	_SogouDictIDOffset, _SogouDictIDSize     = 0, 12
	_PinYinTableIDOffset, _PinYinTableIDSize = 5440, 4
	_PinYinTableOffset, _PinYinTableSize     = 5444, 4324
	_ItemsOffset                             = 9768
)

type SogouDictItem struct {
	weight int

	Abbr   []string `json:"abbr"`   // abbreviated pinyin of each Chinese character
	Pinyin []string `json:"pinyin"` // pinyin of each Chinese character
	Text   string   `json:"text"`   // the Chinese word
}

type SogouDict struct {
	// basic info
	Name        string `json:"name"`
	Category    string `json:"category"`
	Description string `json:"description"`
	Examples    string `json:"examples"`

	// content of the .scel file, sorted by the frequency of word usage
	Items []SogouDictItem `json:"items"`
}

type byWeight []SogouDictItem

func (a byWeight) Len() int           { return len(a) }
func (a byWeight) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a byWeight) Less(i, j int) bool { return a[i].weight < a[j].weight }

func convertUTF16ToUTF8(in []byte) []byte {
	ret := &bytes.Buffer{}
	buf := make([]byte, 4)
	for i := 0; i < len(in); i += 2 {
		r := utf16.Decode([]uint16{uint16(in[i]) + (uint16(in[i+1]) << 8)})
		n := utf8.EncodeRune(buf, r[0])
		ret.Write(buf[:n])
	}
	return ret.Bytes()
}

func convert(in []byte) (out string) {
	out = string(bytes.Trim(convertUTF16ToUTF8(in), "\x00"))
	return
}

func check(rs io.ReadSeeker, offset, size int64, expected []byte) bool {
	var err error
	actual := make([]byte, size)

	_, err = rs.Seek(offset, 0)
	if err != nil {
		return false
	}

	_, err = rs.Read(actual)
	if err != nil {
		return false
	}

	if !bytes.Equal(actual, expected) {
		return false
	}

	return true
}

func getInfo(rs io.ReadSeeker, offset, size int64) (out string, err error) {
	info := make([]byte, size)

	_, err = rs.Seek(offset, 0)
	if err != nil {
		return
	}

	_, err = rs.Read(info)
	if err != nil {
		return
	}

	out = convert(info)
	return
}

func getItems(rs io.ReadSeeker) (items []SogouDictItem, err error) {
	if !check(rs, _PinYinTableIDOffset, _PinYinTableIDSize, _PinyinTableID) {
		err = ErrCorruptedDict
		return
	}

	integer := make([]byte, 2)

	table := make(map[uint16]string)

	for {
		var pos int64
		pos, err = rs.Seek(0, 1)
		if err != nil {
			return
		}
		if pos >= _ItemsOffset {
			break
		}

		_, err = rs.Read(integer)
		if err != nil {
			return
		}
		index := _Uint16(integer)

		_, err = rs.Read(integer)
		if err != nil {
			return
		}

		var py []byte
		pyLen := int(_Uint16(integer))
		for pyLen > 0 {
			_, err = rs.Read(integer)
			if err != nil {
				return
			}
			for _, b := range integer {
				if b == 0 {
					continue
				}
				py = append(py, b)
			}
			pyLen -= len(integer)
		}
		table[index] = string(py)
	}

	_, err = rs.Seek(_ItemsOffset, 0)
	if err != nil {
		return
	}

	for {
		_, err = rs.Read(integer)
		if err != nil {
			if err == io.EOF {
				err = nil
				break
			}
			return
		}
		count := int(_Uint16(integer))

		_, err = rs.Read(integer)
		if err != nil {
			return
		}
		pinyinLen := int(_Uint16(integer))

		var pinyin []string
		var abbr []string

		for pinyinLen > 0 {
			_, err = rs.Read(integer)
			if err != nil {
				return
			}
			py := table[_Uint16(integer)]
			if len(py) == 0 {
				//err = ErrCorruptedDict
				//return
				continue
			}
			pinyin = append(pinyin, py)
			abbr = append(abbr, py[0:1])
			pinyinLen -= len(integer)
		}

		for i := 0; i < count; i++ {
			_, err = rs.Read(integer)
			if err != nil {
				if err == io.EOF {
					err = nil
					break
				}
				return
			}

			word := make([]byte, _Uint16(integer))
			_, err = rs.Read(word)
			if err != nil {
				if err == io.EOF {
					err = nil
					break
				}
				return
			}

			_, err = rs.Read(integer)
			if err != nil {
				if err == io.EOF {
					err = nil
					break
				}
				return
			}

			weight := make([]byte, _Uint16(integer))
			_, err = rs.Read(weight)
			if err != nil {
				if err == io.EOF {
					err = nil
					break
				}
				return
			}

			items = append(items, SogouDictItem{
				weight: int(_Uint16(weight)),

				Abbr:   abbr,
				Pinyin: pinyin,
				Text:   convert(word),
			})
		}
	}

	sort.Sort(byWeight(items))

	return
}

// Read content from a Reader.
func Parse(rs io.ReadSeeker) (dict SogouDict, err error) {
	if !check(rs, _SogouDictIDOffset, _SogouDictIDSize, _SogouDictID) {
		err = ErrInvalidDict
		return
	}

	dict.Name, err = getInfo(rs, _NameOffset, _NameSize)
	if err != nil {
		return
	}

	dict.Category, err = getInfo(rs, _CategoryOffset, _CategorySize)
	if err != nil {
		return
	}

	dict.Description, err = getInfo(rs, _DescriptionOffset, _DescriptionSize)
	if err != nil {
		return
	}

	dict.Examples, err = getInfo(rs, _ExamplesOffset, _ExamplesSize)
	if err != nil {
		return
	}

	dict.Items, err = getItems(rs)
	return
}

// Read content from a file.
func ParseFile(file string) (dict SogouDict, err error) {
	var f *os.File
	f, err = os.Open(file)
	if err != nil {
		return
	}
	defer f.Close()
	var rs io.ReadSeeker
	rs = f
	dict, err = Parse(rs)
	return
}
