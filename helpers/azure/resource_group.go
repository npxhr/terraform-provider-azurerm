package azure

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/hashicorp/terraform-provider-azurerm/internal/tf/pluginsdk"
	"github.com/hashicorp/terraform-provider-azurerm/internal/tf/suppress"
)

func SchemaResourceGroupName() *pluginsdk.Schema {
	return &pluginsdk.Schema{
		Type:         pluginsdk.TypeString,
		Required:     true,
		ForceNew:     true,
		ValidateFunc: ValidateResourceGroupName,
	}
}

func SchemaResourceGroupNameDiffSuppress() *pluginsdk.Schema {
	return &pluginsdk.Schema{
		Type:             pluginsdk.TypeString,
		Required:         true,
		ForceNew:         true,
		DiffSuppressFunc: suppress.CaseDifference,
		ValidateFunc:     ValidateResourceGroupName,
	}
}

// Deprecated: use `commonschema.ResourceGroupNameForDataSource()` instead
func SchemaResourceGroupNameForDataSource() *pluginsdk.Schema {
	return &pluginsdk.Schema{
		Type:         pluginsdk.TypeString,
		Required:     true,
		ValidateFunc: ValidateResourceGroupName,
	}
}

func ValidateResourceGroupName(v interface{}, k string) (warnings []string, errors []error) {
	value := v.(string)

	if len(value) > 90 {
		errors = append(errors, fmt.Errorf("%q may not exceed 90 characters in length", k))
	}

	if strings.HasSuffix(value, ".") {
		errors = append(errors, fmt.Errorf("%q may not end with a period", k))
	}

	if len(value) == 0 {
		errors = append(errors, fmt.Errorf("%q cannot be blank", k))
	} else if matched := regexp.MustCompile(`^[-\w._()]+$`).Match([]byte(value)); !matched {
		// regex pulled from https://docs.microsoft.com/en-us/rest/api/resources/resourcegroups/createorupdate
		errors = append(errors, fmt.Errorf("%q may only contain alphanumeric characters, dash, underscores, parentheses and periods", k))
	}

	return warnings, errors
}
