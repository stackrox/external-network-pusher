package azure

import (
	"fmt"
)

// InvalidAzureCompoundRegionName is returned when an invalid Azure compound region name is found
func InvalidAzureCompoundRegionName(regionName string) error {
	return fmt.Errorf("invalid compound region name found: %s", regionName)
}

// InvalidAzureCompoundServiceName is returned when an invalid Azure compound service name is found
func InvalidAzureCompoundServiceName(serviceName string) error {
	return fmt.Errorf("invalid compound service name found: %s", serviceName)
}
