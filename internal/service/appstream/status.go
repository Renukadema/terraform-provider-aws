package appstream

import (
	"context"

	"github.com/aws/aws-sdk-go/service/appstream"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

//statusStackState fetches the fleet and its state
func statusStackState(ctx context.Context, conn *appstream.AppStream, name string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		stack, err := findStackByName(ctx, conn, name)
		if err != nil {
			return nil, "Unknown", err
		}

		if stack == nil {
			return stack, "NotFound", nil
		}

		return stack, "AVAILABLE", nil
	}
}
