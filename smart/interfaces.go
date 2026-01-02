// Package smart provides high-level helpers for the rest package.
package smart

import (
	"reflect"

	"github.com/ByteSizedMarius/go-fritzbox-api/rest"
)

// GetInterface returns the interface of type I from the unit's interfaces.
// The type parameter must be a pointer type matching one of the IFUnitInterfaces fields.
// Returns (value, true) if found and non-nil, or (zero, false) otherwise.
//
// Example:
//
//	therm, ok := smart.GetInterface[*rest.IFThermostatOverview](unit)
//	if ok {
//	    fmt.Println(therm.SetPointTemperature.Celsius)
//	}
func GetInterface[I any](u *rest.HelperOverviewUnit) (I, bool) {
	var zero I
	if u == nil {
		return zero, false
	}

	targetType := reflect.TypeOf(zero)
	v := reflect.ValueOf(u.Interfaces)

	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		if field.Type() == targetType && !field.IsNil() {
			return field.Interface().(I), true
		}
	}
	return zero, false
}

// HasInterface returns true if the unit has a non-nil interface of type I.
//
// Example:
//
//	if smart.HasInterface[*rest.IFThermostatOverview](unit) {
//	    // unit is a thermostat
//	}
func HasInterface[I any](u *rest.HelperOverviewUnit) bool {
	_, ok := GetInterface[I](u)
	return ok
}
