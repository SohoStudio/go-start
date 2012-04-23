package model

import (
	"reflect"
	"strings"
)

/*
MetaData of a model field.
Use of any methods on a nil pointer is valid.
*/
type MetaData struct {
	parent      *MetaData
	depth       int
	name        string
	index1      int // start with index 1 to use default 0 as invalid index
	tag         string
	attribCache map[string]string
}

func (self *MetaData) Parent() *MetaData {
	if self == nil {
		return nil
	}
	return self.parent
}

func (self *MetaData) Depth() int {
	if self == nil {
		return 0
	}
	return self.depth
}

func (self *MetaData) Name() string {
	if self == nil {
		return ""
	}
	return self.name
}

func (self *MetaData) HasIndex() bool {
	if self == nil {
		return false
	}
	return self.index1 > 0
}

func (self *MetaData) Index() int {
	if self == nil {
		return -1
	}
	return self.index - 1
}

func (self *MetaData) Attrib(name string) (value string, ok bool) {
	if self == nil {
		return "", false
	}
	if self.attribCache == nil {
		for _, s := range strings.Split(self.tag, "|") {
			if self.attribCache == nil {
				self.attribCache = make(map[string]string)
			}
			pos := strings.Index(s, "=")
			if pos == -1 {
				self.attribCache[s] = "true"
			} else {
				self.attribCache[s[:pos]] = s[pos+1:]
			}
		}
	}
	if self.attribCache == nil {
		return "", false
	}
	value, ok = self.attribCache[name]
	return value, ok
}

func (self *MetaData) BoolAttrib(name string) bool {
	if self == nil {
		return false
	}
	value, ok := self.Attrib(name)
	return ok && value == "true"
}

func (self *MetaData) Selector() string {
	if self == nil {
		return ""
	}
	names := make([]string, self.depth)
	for i, m := self.depth-1, self; i >= 0; i-- {
		names[i] = m.name
		m = m.parent
	}
	return strings.Join(names, ".")
}

func (self *MetaData) ArrayWildcardSelector() string {
	if self == nil {
		return ""
	}
	names := make([]string, self.depth)
	for i, m := self.depth-1, self; i >= 0; i-- {
		if m.HasIndex() {
			names[i] = "$"
		} else {
			names[i] = m.name
		}
		m = m.Parent
	}
	return strings.Join(names, ".")
}
