// Code generated by mockery (devel). DO NOT EDIT.

package mocks

import (
	crane "github.com/google/go-containerregistry/pkg/crane"
	mock "github.com/stretchr/testify/mock"
)

// ICraneWrapper is an autogenerated mock type for the ICraneWrapper type
type ICraneWrapper struct {
	mock.Mock
}

// Digest provides a mock function with given fields: ref, opt
func (_m *ICraneWrapper) Digest(ref string, opt ...crane.Option) (string, error) {
	_va := make([]interface{}, len(opt))
	for _i := range opt {
		_va[_i] = opt[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, ref)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	var r0 string
	if rf, ok := ret.Get(0).(func(string, ...crane.Option) string); ok {
		r0 = rf(ref, opt...)
	} else {
		r0 = ret.Get(0).(string)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string, ...crane.Option) error); ok {
		r1 = rf(ref, opt...)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}
