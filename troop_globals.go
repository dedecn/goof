package goof

import (
	"debug/dwarf"
	"reflect"
	"regexp"
	"sort"
	"strings"
	"unsafe"
	"fmt"
	"github.com/zeebo/errs"
)

var statictmpRe = regexp.MustCompile(`statictmp_\d+$`)

func (t *Troop) addGlobals() error {
	reader := t.data.Reader()

	for {
		entry, err := reader.Next()
		if err != nil {
			return errs.Wrap(err)
		}
		if entry == nil {
			break
		}

		if entry.Tag != dwarf.TagVariable {
			continue
		}

		name, ok := entry.Val(dwarf.AttrName).(string)
		if !ok {
			continue
		}

		// filter out some values that aren't useful and just clutter stuff
		if strings.Contains(name, "·") {
			continue
		}
		if statictmpRe.MatchString(name) {
			continue
		}

		loc, err := entryLocation(t.data, entry)
		if err != nil {
			continue
		}
		t.variables[name] = uintptr(loc)

		dtyp, err := entryType(t.data, entry)
		if err != nil {
			continue
		}

		dname := dwarfTypeName(dtyp)
		if dname == "<unspecified>" || dname == "" {
			continue
		}

		rtyp := t.types[dname]
		if rtyp == nil {
			continue
		}

		ptr := unsafe.Pointer(uintptr(loc))
		t.globals[name] = reflect.NewAt(rtyp, ptr).Elem()
	}

	return nil
}

func (t *Troop) Globals() ([]string, error) {
	if err := t.check(); err != nil {
		return nil, err
	}
	out := make([]string, 0, len(t.globals))
	for name := range t.globals {
		out = append(out, name)
	}
	sort.Strings(out)
	return out, nil
}

func (t *Troop) Global(name string) (reflect.Value, error) {
	if err := t.check(); err != nil {
		return reflect.Value{}, t.err
	}
	return t.globals[name], nil
}

func (t *Troop) Variables() ([]string, error) {
	if err := t.check(); err != nil {
		return nil, err
	}
	out := make([]string, 0, len(t.variables))
	for name := range t.variables {
		out = append(out, name)
	}
	sort.Strings(out)
	return out, nil
}

func (t *Troop) Variable(name string) (uintptr, error) {
	if err := t.check(); err != nil {
		return 0, t.err
	}
	if ret, ok := t.variables[name]; ok {
		return ret, nil
	} else {
		return 0, fmt.Errorf("%s not found", name)
	}
}
