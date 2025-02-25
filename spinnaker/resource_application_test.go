package spinnaker

import (
	"context"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccSpinnakerApplication_basic(t *testing.T) {
	resourceName := "spinnaker_application.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccSpinnakerApplication_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckApplicationExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "application", rName),
					resource.TestCheckResourceAttr(resourceName, "email", "acceptance@test.com"),
				),
			},
		},
	})
}

func testAccCheckApplicationExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Application Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Application ID is set")
		}

		client, err := testAccProvider.Meta().(*clientConfig).Client()
		if err != nil {
			return err
		}

		err = retry.RetryContext(context.Background(), 1*time.Minute, func() *retry.RetryError {
			_, resp, err := client.ApplicationControllerApi.GetApplicationUsingGET(client.Context, rs.Primary.ID, nil)
			if resp != nil {
				if resp != nil && resp.StatusCode == http.StatusNotFound {
					return retry.RetryableError(fmt.Errorf("application does not exit"))
				} else if resp.StatusCode != http.StatusOK {
					return retry.NonRetryableError(fmt.Errorf("encountered an error getting application, status code: %d", resp.StatusCode))
				}
			}
			if err != nil {
				return retry.NonRetryableError(err)
			}
			return nil
		})
		if err != nil {
			return fmt.Errorf("Unable to find Application after retries: %s", err)
		}
		return nil
	}
}

func testAccSpinnakerApplication_basic(rName string) string {
	return fmt.Sprintf(`
resource "spinnaker_application" "test" {
	application  = %q
	email = "acceptance@test.com"
}
`, rName)
}
