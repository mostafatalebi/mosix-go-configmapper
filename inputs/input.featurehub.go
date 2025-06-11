package inputs

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"
)

type FHInput struct {
	client   *http.Client
	features map[string]FHValue

	lock   *sync.RWMutex
	server string
	apiKey string

	enableAutoRefresh bool

	// if auto refreshing is enabled, this value cannot be zero
	autoRefreshInterval time.Duration

	refreshCount int64
}

const InputFHName = "feature-hub"

func NewFHInput(addr, apiKey string) (*FHInput, error) {
	var fh = &FHInput{
		server:       addr,
		apiKey:       apiKey,
		lock:         &sync.RWMutex{},
		refreshCount: 0,
	}
	if addr == "" || apiKey == "" {
		return nil, errors.New("addr and apiKey cannot be empty")
	}

	fh.client = &http.Client{}
	fh.client.Timeout = time.Second * 10
	var qualifiedUrl = fh.getUrl()
	var err = fh.fetchFeaturesWithRequest(qualifiedUrl)
	if err != nil {
		return nil, err
	}

	return fh, nil
}

// AutoRefreshing
// Warning: DO NOT USE
// this is meant for direct specific usages and/or test purposes
// Use Reload() function of InputController
func (fh *FHInput) AutoRefreshing(enable bool, duration time.Duration) *FHInput {
	fh.enableAutoRefresh = enable
	fh.autoRefreshInterval = duration

	if fh.enableAutoRefresh {
		if fh.autoRefreshInterval == 0 {
			fh.autoRefreshInterval = time.Second * 30
		}
		var qualifiedUrl = fh.getUrl()
		go func() {
			for {
				time.Sleep(fh.autoRefreshInterval)
				if err := fh.fetchFeaturesWithRequest(qualifiedUrl); err != nil {
					fmt.Printf("[feature-hub] -> error in auto-refreshing, skipped this round: %s\n", err.Error())
				} else {
					fh.refreshCount++
				}
			}
		}()
	}
	return fh
}
func (fh *FHInput) GetFeaturesCount() int {
	if fh.features != nil {
		return len(fh.features)
	}
	return 0
}
func (fh *FHInput) fetchFeaturesWithRequest(fullUrl string) error {
	var resp, err = fh.client.Get(fullUrl)
	if err != nil {
		return err
	}
	if resp.StatusCode != 200 {
		return fmt.Errorf("non-200 status code from feature-hub server: %d", resp.StatusCode)
	}

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	fh.lock.Lock()
	fh.features, err = fh.fromJsonToMap(responseBody)
	fh.lock.Unlock()
	if err != nil {
		return err
	}

	return nil
}

func (fh *FHInput) fromJsonToMap(b []byte) (map[string]FHValue, error) {
	if b == nil {
		return nil, errors.New("incoming byte is nil, no config can be created")
	}
	var cnfs = make([]FeatureHubEnvironment, 0)
	if err := json.Unmarshal(b, &cnfs); err != nil {
		return nil, err
	}
	if cnfs == nil || len(cnfs) == 0 {
		return nil, errors.New("no config found, though no json parsing error neither, check server or connection")
	} else if cnfs[0].ID == "" {
		return nil, errors.New("no environment ID found in parsed feature-hub config")
	} else if cnfs[0].Features == nil {
		return nil, errors.New("no features found in response")
	}
	var mappedKeys = make(map[string]FHValue)

	for _, v := range cnfs[0].Features {
		mappedKeys[v.Key] = v
	}
	return mappedKeys, nil
}

func (fh *FHInput) getUrl() string {
	return fmt.Sprintf("%s/features/?apiKey=%s", fh.server, fh.apiKey)
}

func (fh *FHInput) CanRefresh() bool {
	return true
}

func (fh *FHInput) GetString(key string) (string, error) {
	if !fh.Has(key) {
		return "", errors.New("not found")
	}
	fh.lock.RLock()
	defer fh.lock.RUnlock()

	val := fh.features[key]
	if val.Type != "STRING" {
		return "", errors.New("incompatible type for key=" + key)
	} else if v, ok := val.Value.(string); ok {
		return v, nil
	}
	return "", errors.New("incompatible type for key=" + key)
}

func (fh *FHInput) GetNumber(key string) (float64, error) {
	if !fh.Has(key) {
		return 0, errors.New("not found")
	}
	fh.lock.RLock()
	defer fh.lock.RUnlock()

	val := fh.features[key]
	if val.Type != "NUMBER" {
		return 0, errors.New("incompatible type for key=" + key)
	} else if v, ok := val.Value.(float64); ok {
		return v, nil
	}
	return 0, errors.New("incompatible type for key=" + key)
}

func (fh *FHInput) GetBoolean(key string) (bool, error) {
	if !fh.Has(key) {
		return false, errors.New("not found")
	}
	fh.lock.RLock()
	defer fh.lock.RUnlock()

	val := fh.features[key]
	if val.Type != "BOOLEAN" {
		return false, errors.New("incompatible type for key=" + key)
	} else if v, ok := val.Value.(bool); ok {
		return v, nil
	}
	return false, errors.New("incompatible type for key=" + key)
}
func (fh *FHInput) Has(key string) bool {
	fh.lock.RLock()
	defer fh.lock.RUnlock()
	if fh.features != nil {
		if _, ok := fh.features[key]; ok {
			return ok
		}
	}
	return false
}

func (fh *FHInput) Reload() error {
	var err = fh.fetchFeaturesWithRequest(fh.getUrl())
	if err != nil {
		return fmt.Errorf("failed to reload: %s", err.Error())
	}
	return nil
}

func (fh *FHInput) GetInputName() string {
	return InputFHName
}

type FHValue struct {
	ID      string      `json:"id"`
	Key     string      `json:"key"`
	L       bool        `json:"l"`
	Version int64       `json:"version"`
	Type    string      `json:"type"`
	Value   interface{} `json:"value"`
}

type FeatureHubEnvironment struct {
	ID       string `json:"id"`
	Features []FHValue
}
