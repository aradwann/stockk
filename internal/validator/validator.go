package validator

import "errors"

func ValidateID(value int) error {
	if value <= 0 {
		return errors.New("ID must be a positive integer")
	}
	return nil
}

func ValidateQuantity(value int) error {
	if value <= 0 {
		return errors.New("Quatity must be a positive integer")
	}
	return nil
}
