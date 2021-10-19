// Code generated by mockery v1.0.0. DO NOT EDIT.

package mocks

import mock "github.com/stretchr/testify/mock"

// IACRTokenProvider is an autogenerated mock type for the IACRTokenProvider type
type IACRTokenProvider struct {
	mock.Mock
}

// GetACRRefreshToken provides a mock function with given fields: registry
func (_m *IACRTokenProvider) GetACRRefreshToken(registry string) (string, error) {
	ret := _m.Called(registry)

	var r0 string
	if rf, ok := ret.Get(0).(func(string) string); ok {
		r0 = rf(registry)
	} else {
		r0 = ret.Get(0).(string)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string) error); ok {
		r1 = rf(registry)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}
