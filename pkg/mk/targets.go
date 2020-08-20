package mk

import (
	"fmt"
	"net/url"
)

type targetFactory struct {
	factories map[string]func(url.URL)Target
}

var defaultTargetFactory =  targetFactory{
	factories: make(map[string]func(url.URL)Target),
}

func (tf *targetFactory) RegisterTarget(scheme string, factory func (url.URL) Target) {
	if _, exists := tf.factories[scheme]; exists {
		panic(fmt.Errorf("duplicate registration for scheme %s", scheme))
	}
	tf.factories[scheme] = factory
}

func RegisterTarget(scheme string, factory func (url.URL) Target) {
	defaultTargetFactory.RegisterTarget(scheme, factory)
}