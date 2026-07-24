package hunspell

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
)

// pathPair ports Hunspell.PathPair (dictionary + affix absolute paths).
type pathPair struct {
	dictionary string
	affix      string
}

// resourcePair ports Hunspell.ResourcePair (classpath resource paths).
type resourcePair struct {
	dictionaryPath string
	affixPath      string
}

// Factory ports org.languagetool.rules.spelling.hunspell.Hunspell.Factory.
// Default creates FileHunspellDictionary (pure-Go word list; native Hunspell JNI not used).
type Factory interface {
	// CreateFromLocalFiles ports createFromLocalFiles — caller owns the files (deleteOnClose=false).
	CreateFromLocalFiles(languageCode, dictionary, affix string) (HunspellDictionary, error)
	// CreateFromStreams ports createFromStreams — implementations extract data before return.
	// Caller closes the streams. Default writes temp files (deleteOnClose=false, delete on process exit via Remove not automatic).
	CreateFromStreams(languageCode string, dictionary, affix io.Reader) (HunspellDictionary, error)
}

// defaultFactory ports Hunspell.viaTempFiles factory using FileHunspellDictionary.
type defaultFactory struct{}

func (defaultFactory) CreateFromLocalFiles(languageCode, dictionary, affix string) (HunspellDictionary, error) {
	_ = languageCode
	return NewFileHunspellDictionary(dictionary, affix, false)
}

func (defaultFactory) CreateFromStreams(languageCode string, dictionary, affix io.Reader) (HunspellDictionary, error) {
	pair, err := createTempFilesFromStreams(languageCode, dictionary, affix)
	if err != nil {
		return nil, err
	}
	// Java JAR path: deleteOnClose=false; temp files live for JVM lifetime (deleteOnExit).
	// Go: leave files; tests may clean via Close when deleteOnClose true is requested elsewhere.
	return NewFileHunspellDictionary(pair.dictionary, pair.affix, false)
}

func createTempFilesFromStreams(language string, dictionaryStream, affixStream io.Reader) (pathPair, error) {
	if language == "" {
		language = "hun"
	}
	dic, err := os.CreateTemp("", language+"-*.dic")
	if err != nil {
		return pathPair{}, err
	}
	dicPath := dic.Name()
	if _, err := io.Copy(dic, dictionaryStream); err != nil {
		_ = dic.Close()
		_ = os.Remove(dicPath)
		return pathPair{}, err
	}
	if err := dic.Close(); err != nil {
		_ = os.Remove(dicPath)
		return pathPair{}, err
	}

	aff, err := os.CreateTemp("", language+"-*.aff")
	if err != nil {
		_ = os.Remove(dicPath)
		return pathPair{}, err
	}
	affPath := aff.Name()
	if affixStream != nil {
		if _, err := io.Copy(aff, affixStream); err != nil {
			_ = aff.Close()
			_ = os.Remove(dicPath)
			_ = os.Remove(affPath)
			return pathPair{}, err
		}
	}
	if err := aff.Close(); err != nil {
		_ = os.Remove(dicPath)
		_ = os.Remove(affPath)
		return pathPair{}, err
	}
	return pathPair{dictionary: dicPath, affix: affPath}, nil
}

var (
	hunspellMu                sync.Mutex
	hunspellDictionaryFactory Factory = defaultFactory{}
	pathCache                         = map[pathPair]HunspellDictionary{}
	resourceCache                     = map[resourcePair]HunspellDictionary{}
)

// SetHunspellStreamFactory ports Hunspell.setHunspellStreamFactory.
func SetHunspellStreamFactory(f Factory) {
	hunspellMu.Lock()
	defer hunspellMu.Unlock()
	if f == nil {
		hunspellDictionaryFactory = defaultFactory{}
		return
	}
	hunspellDictionaryFactory = f
}

// GetDictionary ports Hunspell.getDictionary(Path, Path) — cache by absolute path pair.
// Files are NOT deleted on close (caller owns them).
func GetDictionary(dictionary, affix string) HunspellDictionary {
	hunspellMu.Lock()
	defer hunspellMu.Unlock()

	key := pathPair{dictionary: dictionary, affix: affix}
	if h, ok := pathCache[key]; ok && h != nil && !h.IsClosed() {
		return h
	}
	newH, err := hunspellDictionaryFactory.CreateFromLocalFiles(
		filepath.Base(dictionary), dictionary, affix)
	if err != nil {
		// Java wraps IOException in RuntimeException
		panic(fmt.Errorf("hunspell GetDictionary: %w", err))
	}
	pathCache[key] = newH
	return newH
}

// ForDictionaryInResources ports Hunspell.forDictionaryInResources overloads.
//
//	ForDictionaryInResources(lang) → dic/aff at ""+lang+".dic"/".aff"
//	ForDictionaryInResources(lang, resourcePath) → resourcePath+lang+".dic"/".aff"
//	ForDictionaryInResources(lang, dicPath, affPath) → explicit resource paths
//
// Resolves classpath paths via DiscoverHunspellDic (file:// twin). Missing resources panic
// like Java RuntimeException ("Could not find the dictionary...").
func ForDictionaryInResources(language string, resourcePaths ...string) HunspellDictionary {
	var dicPath, affPath string
	switch len(resourcePaths) {
	case 0:
		dicPath = language + ".dic"
		affPath = language + ".aff"
	case 1:
		base := resourcePaths[0]
		dicPath = base + language + ".dic"
		affPath = base + language + ".aff"
	default:
		dicPath = resourcePaths[0]
		affPath = resourcePaths[1]
	}
	// Normalize leading slash like Java resource paths
	if dicPath != "" && dicPath[0] != '/' {
		dicPath = "/" + dicPath
	}
	if affPath != "" && affPath[0] != '/' {
		affPath = "/" + affPath
	}

	hunspellMu.Lock()
	defer hunspellMu.Unlock()

	key := resourcePair{dictionaryPath: dicPath, affixPath: affPath}
	if cached, ok := resourceCache[key]; ok && cached != nil && !cached.IsClosed() {
		return cached
	}

	// Java: try file:// URLs then streams. Go: DiscoverHunspellDic for .dic + companion .aff.
	resolvedDic := DiscoverHunspellDic(dicPath)
	if resolvedDic == "" {
		panic(fmt.Sprintf("Could not find the dictionary for language %q in the classpath (%s)", language, dicPath))
	}
	resolvedAff := companionAff(resolvedDic)
	// If explicit aff classpath differs, still prefer companion next to dic (Java loads both from broker).

	dict, err := hunspellDictionaryFactory.CreateFromLocalFiles(language, resolvedDic, resolvedAff)
	if err != nil {
		panic(fmt.Errorf("Could not create dictionary for language %q: %w", language, err))
	}
	resourceCache[key] = dict
	// Also cache by path pair for getDictionary reuse (Java resource file:// path)
	pathCache[pathPair{dictionary: resolvedDic, affix: resolvedAff}] = dict
	return dict
}

// ClearHunspellCaches clears path/resource caches (tests; no Java twin).
func ClearHunspellCaches() {
	hunspellMu.Lock()
	defer hunspellMu.Unlock()
	pathCache = map[pathPair]HunspellDictionary{}
	resourceCache = map[resourcePair]HunspellDictionary{}
}
