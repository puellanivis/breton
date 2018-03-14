package m3u8

import (
	"fmt"
	"strings"
	"sync"
)

type enum struct {
	toValue map[string]int
	toName  map[int]string
}

var (
	enumsMutex sync.Mutex
	enums      map[string]*enum
)

func getEnum(valid string) *enum {
	enumsMutex.Lock()
	defer enumsMutex.Unlock()

	if e, ok := enums[valid]; ok {
		return e
	}

	if enums == nil {
		enums = make(map[string]*enum)
	}

	e := &enum{
		toValue: make(map[string]int),
		toName:  make(map[int]string),
	}

	values := strings.Split(valid, ",")
	for i, value := range values {
		if value != "" {
			e.toValue[value] = i
			e.toName[i] = value
		}
	}

	enums[valid] = e

	return e
}

func (e *enum) Value(i int) (string, error) {
	if name, ok := e.toName[i]; ok {
		return name, nil
	}

	return "", fmt.Errorf("invalid enum value: %d", i)
}

func (e *enum) Index(value string) (int, error) {
	if value == "" {
		return 0, nil
	}

	if i, ok := e.toValue[value]; ok {
		return i, nil
	}

	return 0, fmt.Errorf("invalid enum: %q", value)
}

func (e *enum) Test(value string) (string, error) {
	_, err := e.Index(value)
	return value, err
}
