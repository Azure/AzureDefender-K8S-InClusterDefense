// Code generated by mockery v1.0.0. DO NOT EDIT.

package mocks

import mock "github.com/stretchr/testify/mock"

// IACRTokenExchanger is an autogenerated mock type for the IACRTokenExchanger type
type IACRTokenExchanger struct {
	mock.Mock
}

// ExchangeACRAccessToken provides a mock function with given fields: registry, armToken
func (_m *IACRTokenExchanger) ExchangeACRAccessToken(registry string, armToken string) (string, error) {
	ret := _m.Called(registry, armToken)

	var r0 string
	if rf, ok := ret.Get(0).(func(string, string) string); ok {
		r0 = rf(registry, armToken)
	} else {
		r0 = ret.Get(0).(string)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string, string) error); ok {
		r1 = rf(registry, armToken)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}
