package hunspell

import (
	"os"
)

// FileHunspellDictionary ports DumontsHunspellDictionary as a pure-Go .dic word list
// (no native Hunspell/JNI). Affix rules are ignored; spell/suggest use the map backend.
type FileHunspellDictionary struct {
	*MapHunspellDictionary
	dictionaryPath string
	affixPath      string
	deleteOnClose  bool
}

// NewFileHunspellDictionary loads words from a Hunspell .dic (first line = count optional).
func NewFileHunspellDictionary(dictionaryPath, affixPath string, deleteOnClose bool) (*FileHunspellDictionary, error) {
	f, err := os.Open(dictionaryPath)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	m, err := NewMapHunspellDictionaryFromDic(f)
	if err != nil {
		return nil, err
	}
	return &FileHunspellDictionary{
		MapHunspellDictionary: m,
		dictionaryPath:        dictionaryPath,
		affixPath:             affixPath,
		deleteOnClose:         deleteOnClose,
	}, nil
}

func (d *FileHunspellDictionary) DictionaryPath() string { return d.dictionaryPath }
func (d *FileHunspellDictionary) AffixPath() string      { return d.affixPath }
func (d *FileHunspellDictionary) DeleteOnClose() bool    { return d.deleteOnClose }

func (d *FileHunspellDictionary) Close() error {
	if d == nil {
		return nil
	}
	if d.MapHunspellDictionary != nil {
		_ = d.MapHunspellDictionary.Close()
	}
	if !d.deleteOnClose {
		return nil
	}
	_ = os.Remove(d.dictionaryPath)
	if d.affixPath != "" {
		_ = os.Remove(d.affixPath)
	}
	return nil
}

var _ HunspellDictionary = (*FileHunspellDictionary)(nil)
