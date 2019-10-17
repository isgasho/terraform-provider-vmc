package vmc

import (
	"fmt"
	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	"gitlab.eng.vmware.com/het/vmware-vmc-sdk/vapi/bindings/vmc/orgs/sddcs/publicips"
	"gitlab.eng.vmware.com/het/vmware-vmc-sdk/vapi/runtime/protocol/client"
	"os"
	"testing"
)

func TestAccResourceVmcPublicIP_basic(t *testing.T) {
	vmName := "terraform_test_vm_" + acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	vmList := []string {vmName}
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testCheckVmcSddcDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccVmcPublicIPConfigBasic(vmList),
				Check: resource.ComposeTestCheckFunc(
					testCheckVmcPublicIPExists("vmc_publicips.publicip_1"),
				),
			},
		},
	})
}

func testCheckVmcPublicIPExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		sddcID := rs.Primary.Attributes["sddc_id"]
		orgID := rs.Primary.Attributes["org_id"]
		vmName := rs.Primary.Attributes["names"]
		connector := testAccProvider.Meta().(client.Connector)
		publicIPClient := publicips.NewPublicipsClientImpl(connector)

		publicIPList , err := publicIPClient.List(orgID, sddcID)
		if err != nil {
			return fmt.Errorf("Bad: List on publicIP: %s", err)
		}
		allocationID := publicIPList[0].AllocationId
		publicIP, err := publicIPClient.Get(orgID, sddcID, *allocationID)
		if err != nil {
			return fmt.Errorf("Bad: Get on publicIP: %s", err)
		}
        fmt.Println("Inside test for public IPs")
		fmt.Println(vmName)
		if *publicIP.Name != vmName {
			return fmt.Errorf("Bad: Public IP %q does not exist", *allocationID)
		}

		fmt.Printf("Public IP created successfully with id %s ", *allocationID)
		return nil
	}
}

/*
func testCheckVmcSddcDestroy(s *terraform.State) error {

	connector := testAccProvider.Meta().(client.Connector)
	sddcClient := sddcs.NewSddcsClientImpl(connector)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "vmc_sddc" {
			continue
		}

		sddcID := rs.Primary.Attributes["id"]
		orgID := rs.Primary.Attributes["org_id"]
		task, err := sddcClient.Delete(orgID, sddcID, nil, nil, nil)
		if err != nil {
			return fmt.Errorf("Error while deleting sddc %s, %s", sddcID, err)
		}
		err = WaitForTask(connector, orgID, task.Id)
		if err != nil {
			return fmt.Errorf("Error while waiting for task %q: %v", task.Id, err)
		}
	}

	return nil
}*/

func testAccVmcPublicIPConfigBasic(name []string) string {
	return fmt.Sprintf(`
provider "vmc" {
	refresh_token = %q
	
	# refresh_token = "ac5140ea-1749-4355-a892-56cff4893be0"
	 csp_url       = "https://console-stg.cloud.vmware.com"
    vmc_url = "https://stg.skyscraper.vmware.com"
}
	
data "vmc_org" "my_org" {
	id = %q
}

resource "vmc_publicips" "publicip_1" {
	org_id = "${data.vmc_org.my_org.id}"
	sddc_id = "4251fb8e-6fba-4880-aedb-ce3485873941"
	names     = %q
	host_count = 1
	private_ips = ["10.105.167.133"]
}
`,
		os.Getenv("REFRESH_TOKEN"),
		os.Getenv("ORG_ID"),
		name,
	)
}