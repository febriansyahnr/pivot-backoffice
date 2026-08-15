package liveFeature

import "reflect"

type AppVersion struct {
	Versions map[string]string `json:"appVersion"`
}

func (a *AppVersion) Equals(b *AppVersion) bool {
	return reflect.DeepEqual(a.Versions, b.Versions)
}
