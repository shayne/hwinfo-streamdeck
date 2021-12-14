package hwinfostreamdeckplugin

import (
	"fmt"
	"sync"
	"time"
)

type actionManager struct {
	mux     sync.RWMutex
	actions map[string]*actionData
}

func newActionManager() *actionManager {
	return &actionManager{actions: make(map[string]*actionData)}
}

func (tm *actionManager) Run(updateTiles func(*actionData)) {
	go func() {
		ticker := time.NewTicker(time.Second)
		for range ticker.C {
			tm.mux.RLock()
			for _, data := range tm.actions {
				if data.settings.IsValid {
					updateTiles(data)
				}
			}
			tm.mux.RUnlock()
		}
	}()
}

func (tm *actionManager) SetAction(action, context string, settings *actionSettings) {
	tm.mux.Lock()
	tm.actions[context] = &actionData{action, context, settings}
	tm.mux.Unlock()
}

func (tm *actionManager) RemoveAction(context string) {
	tm.mux.Lock()
	delete(tm.actions, context)
	tm.mux.Unlock()
}

func (tm *actionManager) getSettings(context string) (actionSettings, error) {
	tm.mux.RLock()
	data, ok := tm.actions[context]
	tm.mux.RUnlock()
	if !ok {
		return actionSettings{}, fmt.Errorf("getSettings invalid key: %s", context)
	}
	// return full copy of settings, not reference to stored settings
	return *data.settings, nil
}
